package usecase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/domain/entity"
	"github.com/TheKingDevs/tamk/internal/domain/repository"
	"github.com/TheKingDevs/tamk/internal/tools"
	"github.com/TheKingDevs/tamk/pkg/errors"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

func apkFilename(project *entity.Project, env string) string {
	slug := strings.ToLower(strings.ReplaceAll(project.Name, " ", "-"))
	return fmt.Sprintf("%s-%s-%s.apk", slug, project.Version, env)
}

// resolveKeystorePath finds the keystore: project-local first, then fallback to config.
func resolveKeystorePath(projectPath, cfgKeystore string) string {
	keystorePath := filepath.Join(projectPath, "secret", "project.keystore")
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return cfgKeystore
	}
	return keystorePath
}

// validateKeystorePassword verifies the keystore password before build starts.
// Returns nil if valid, error if invalid or keystore not found.
func validateKeystorePassword(keystorePath, password string) error {
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		return fmt.Errorf("keystore not found at %s: %w", keystorePath, errors.ErrKeystoreNotFound)
	}

	if password == "" {
		return fmt.Errorf("keystore password required: %w", errors.ErrKeystoreInvalidPass)
	}

	// Use keytool to verify password by listing keystore contents
	cmd := exec.Command("keytool", "-list",
		"-keystore", keystorePath,
		"-storepass", password,
		"-noprompt",
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.Debug("Keystore validation failed", "output", string(output))
		return fmt.Errorf("keystore password incorrect: %w", errors.ErrKeystoreInvalidPass)
	}

	// Verify we actually got a valid response
	if !strings.Contains(string(output), "Keystore type") && !strings.Contains(string(output), "Entry count") {
		logger.Debug("Keystore validation failed: unexpected output", "output", string(output))
		return fmt.Errorf("keystore password incorrect: %w", errors.ErrKeystoreInvalidPass)
	}

	return nil
}

type BuildProjectUseCase struct {
	cfg       *config.Config
	buildRepo repository.BuildRepository
	projRepo  repository.ProjectRepository
	tools     tools.ToolManager
}

func NewBuildProjectUseCase(
	cfg *config.Config,
	buildRepo repository.BuildRepository,
	projRepo repository.ProjectRepository,
) *BuildProjectUseCase {
	toolMgr := tools.New(tools.Config{
		DevDir:  cfg.DevDir,
		SDKPath: cfg.SDKPath,
	})

	return &BuildProjectUseCase{
		cfg:       cfg,
		buildRepo: buildRepo,
		projRepo:  projRepo,
		tools:     toolMgr,
	}
}

type BuildInput struct {
	ProjectPath string
	Password    string
	Guardian    bool // Enable Guardian security protection
}

func (uc *BuildProjectUseCase) FullBuild(ctx context.Context, input BuildInput) (*entity.BuildResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if !uc.projRepo.Exists(ctx, input.ProjectPath) {
		return nil, fmt.Errorf("no project found at %s", input.ProjectPath)
	}

	project, err := uc.projRepo.Load(ctx, input.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	if input.Password != "" && len(input.Password) < 6 {
		return nil, fmt.Errorf("keystore password too short: %w", errors.ErrKeystoreInvalidPass)
	}

	// Validate keystore password before starting build
	keystorePath := resolveKeystorePath(input.ProjectPath, uc.cfg.Keystore)
	if err := validateKeystorePassword(keystorePath, input.Password); err != nil {
		return nil, err
	}

	currentHash, err := uc.buildRepo.CalculateProjectHash(ctx, input.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	mustRecompile, err := uc.buildRepo.MustRecompile(ctx, input.ProjectPath, currentHash)
	if err == nil && !mustRecompile {
		expectedAPK := filepath.Join(input.ProjectPath, apkFilename(project, "release"))
		if _, err := os.Stat(expectedAPK); err == nil {
			logger.Info("Nothing changed, skipping build")
			return &entity.BuildResult{Success: true, APKPath: expectedAPK}, nil
		}
		logger.Warn("APK missing, forcing rebuild", "apk", expectedAPK)
	}

	// Apply Guardian security if enabled
	if input.Guardian {
		project.Security = entity.SecurityConfigForLevel(entity.SecurityLevelStandard)
		logger.Step("Applying Guardian security protection...")

		// Encrypt assets
		encryptor := &AssetEncryptor{}
		if err := encryptor.EncryptAssets(input.ProjectPath, input.Password); err != nil {
			logger.Warn("Asset encryption failed", "error", err)
		}
	}

	result := uc.executeBuild(ctx, project, input)
	if result.Success {
		uc.buildRepo.SaveCache(ctx, input.ProjectPath, &entity.BuildCache{Hash: currentHash})
	}

	return result, nil
}

func (uc *BuildProjectUseCase) AssetsOnlyBuild(ctx context.Context, input BuildInput) (*entity.BuildResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	project, err := uc.projRepo.Load(ctx, input.ProjectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load project: %w", err)
	}

	// Validate keystore password before starting build
	keystorePath := resolveKeystorePath(input.ProjectPath, uc.cfg.Keystore)
	if err := validateKeystorePassword(keystorePath, input.Password); err != nil {
		logger.Error("Keystore validation failed", "error", err)
		return nil, err
	}

	releaseAPK := apkFilename(project, "release")
	devAPKName := apkFilename(project, "dev")
	baseAPK := filepath.Join(input.ProjectPath, releaseAPK)
	if _, err := os.Stat(baseAPK); os.IsNotExist(err) {
		baseAPK = filepath.Join(input.ProjectPath, devAPKName)
		if _, err := os.Stat(baseAPK); os.IsNotExist(err) {
			return nil, fmt.Errorf("apk base not found: %w", errors.ErrAPKBaseNotFound)
		}
	}

	assetsDir := filepath.Join(input.ProjectPath, "src", "main", "assets")
	if _, err := os.Stat(assetsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("assets directory not found: %w", errors.ErrBuildFailed)
	}

	devAPK := filepath.Join(input.ProjectPath, devAPKName)
	extractDir := filepath.Join(input.ProjectPath, "assets", "cache", "extract")
	os.RemoveAll(extractDir)
	os.MkdirAll(extractDir, 0o755)

	if err := uc.execute(exec.CommandContext(ctx, "unzip", "-o", baseAPK, "-d", extractDir)); err != nil {
		return nil, fmt.Errorf("failed to extract APK: %w", err)
	}

	oldAssets := filepath.Join(extractDir, "assets")
	os.RemoveAll(oldAssets)

	if err := uc.execute(exec.CommandContext(ctx, "cp", "-r", assetsDir, filepath.Join(extractDir, "assets"))); err != nil {
		return nil, fmt.Errorf("failed to copy assets: %w", err)
	}

	if err := zipDir(extractDir, devAPK); err != nil {
		return nil, fmt.Errorf("failed to repackage APK: %w", err)
	}

	apksignerPath, err := uc.tools.ApkSigner(ctx)
	if err != nil {
		return nil, fmt.Errorf("apksigner not available: %w", err)
	}

	signArgs := strings.Fields(apksignerPath)
	signArgs = append(signArgs, "sign",
		"--ks", keystorePath,
		"--ks-pass", "pass:"+input.Password,
		"--out", devAPK+".signed",
		devAPK,
	)
	signCmd := exec.CommandContext(ctx, signArgs[0], signArgs[1:]...)
	if err := uc.execute(signCmd); err != nil {
		return nil, fmt.Errorf("failed to sign APK: %w", err)
	}

	os.Rename(devAPK+".signed", devAPK)

	logger.Info("Incremental assets build complete")
	return &entity.BuildResult{Success: true, APKPath: devAPK}, nil
}

func (uc *BuildProjectUseCase) executeBuild(ctx context.Context, project *entity.Project, input BuildInput) *entity.BuildResult {
	projPath := input.ProjectPath
	finalName := apkFilename(project, "release")
	apkPath := filepath.Join(projPath, "app.apk")
	unsignedPath := filepath.Join(projPath, "app-unsigned.apk")
	finalPath := filepath.Join(projPath, finalName)
	resZip := filepath.Join(projPath, "res.zip")
	objDir := filepath.Join(projPath, "obj")

	// Resolve tool paths via ToolManager
	aapt2Path, err := uc.tools.AAPT2(ctx)
	if err != nil {
		return failedResult(entity.BuildPhaseAAPT2Compile, err.Error())
	}

	kotlincPath, err := uc.tools.KotlinCompiler(ctx)
	if err != nil {
		return failedResult(entity.BuildPhaseKotlinCompile, err.Error())
	}

	d8Path, err := uc.tools.D8(ctx)
	if err != nil {
		return failedResult(entity.BuildPhaseD8, err.Error())
	}

	zipalignPath, err := uc.tools.Zipalign(ctx)
	if err != nil {
		return failedResult(entity.BuildPhaseZipalign, err.Error())
	}

	apksignerPath, err := uc.tools.ApkSigner(ctx)
	if err != nil {
		return failedResult(entity.BuildPhaseApkSign, err.Error())
	}

	sdkPath, err := uc.tools.SDKJar()
	if err != nil {
		return failedResult(entity.BuildPhaseAAPT2Link, err.Error())
	}

	// Build pipeline
	logger.Step("Compiling resources (AAPT2)...")
	if out, err := uc.executeOutput(exec.CommandContext(ctx, aapt2Path, "compile", "--dir",
		filepath.Join(projPath, "res"), "-o", resZip)); err != nil {
		return failedResult(entity.BuildPhaseAAPT2Compile, string(out))
	}

	logger.Step("Linking resources (AAPT2)...")
	genDir := filepath.Join(projPath, "gen")
	os.MkdirAll(genDir, 0o755)
	linkArgs := []string{
		"link",
		"-I", sdkPath,
		"--manifest", filepath.Join(projPath, "AndroidManifest.xml"),
		"-o", apkPath,
		"--java", genDir,
		resZip,
		"--auto-add-overlay",
	}
	assetsDir := filepath.Join(projPath, "src", "main", "assets")
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		linkArgs = append(linkArgs, "-A", assetsDir)
	}
	if out, err := uc.executeOutput(exec.CommandContext(ctx, aapt2Path, linkArgs...)); err != nil {
		return failedResult(entity.BuildPhaseAAPT2Link, string(out))
	}

	kotlinDir := filepath.Join(projPath, "src", "main", "kotlin")

	// If Guardian is enabled, copy security templates to project
	if input.Guardian {
		uc.injectSecurityTemplates(ctx, project, projPath)
	}

	logger.Step("Compiling Kotlin sources...")
	os.MkdirAll(objDir, 0o755)
	kotlinCmd := exec.CommandContext(ctx, kotlincPath,
		kotlinDir, genDir,
		"-cp", sdkPath,
		"-d", objDir,
	)
	if out, err := uc.executeOutput(kotlinCmd); err != nil {
		return failedResult(entity.BuildPhaseKotlinCompile, string(out))
	}

	// Obfuscate code with ProGuard if Guardian is enabled
	if input.Guardian {
		obfuscator := NewProGuardObfuscator()
		if obfuscator.IsAvailable() {
			if err := obfuscator.Obfuscate(ctx, objDir, sdkPath, ""); err != nil {
				logger.Warn("ProGuard obfuscation failed", "error", err)
			}
		}
	}

	logger.Step("Converting to DEX (D8)...")
	var classFiles []string
	filepath.Walk(objDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(path, ".class") {
			classFiles = append(classFiles, path)
		}
		return nil
	})
	if len(classFiles) == 0 {
		return failedResult(entity.BuildPhaseD8, "no .class files found in "+objDir)
	}
	args := []string{"--lib", sdkPath, "--release", "--output", projPath}
	args = append(args, classFiles...)
	d8Cmd := exec.CommandContext(ctx, d8Path, args...)
	if out, err := uc.executeOutput(d8Cmd); err != nil {
		return failedResult(entity.BuildPhaseD8, string(out))
	}

	logger.Step("Packaging DEX into APK...")
	if out, err := uc.executeOutput(exec.CommandContext(ctx, "zip", "-j", apkPath,
		filepath.Join(projPath, "classes.dex"))); err != nil {
		return failedResult(entity.BuildPhasePackageDEX, string(out))
	}

	logger.Step("Aligning (zipalign)...")
	if out, err := uc.executeOutput(exec.CommandContext(ctx, zipalignPath, "-f", "4", apkPath, unsignedPath)); err != nil {
		return failedResult(entity.BuildPhaseZipalign, string(out))
	}

	keystorePath := resolveKeystorePath(projPath, uc.cfg.Keystore)

	logger.Step("Signing APK...")
	signArgs := strings.Fields(apksignerPath)
	signArgs = append(signArgs, "sign",
		"--ks", keystorePath,
		"--ks-pass", "pass:"+input.Password,
		"--out", finalPath,
		unsignedPath,
	)
	signCmd := exec.CommandContext(ctx, signArgs[0], signArgs[1:]...)
	if out, err := uc.executeOutput(signCmd); err != nil {
		return failedResult(entity.BuildPhaseApkSign, string(out))
	}

	os.Remove(resZip)
	os.RemoveAll(objDir)
	os.Remove(apkPath)
	os.Remove(unsignedPath)
	classesDex := filepath.Join(projPath, "classes.dex")
	if _, err := os.Stat(classesDex); err == nil {
		os.Remove(classesDex)
	}

	logger.Success("Build completed")
	return &entity.BuildResult{Success: true, APKPath: finalPath}
}

func (uc *BuildProjectUseCase) PushAssetToDevice(ctx context.Context, projectPath, assetPath string) error {
	if !strings.HasPrefix(assetPath, filepath.Join(projectPath, "src", "main", "assets")) {
		return fmt.Errorf("asset must be inside src/main/assets/")
	}

	relPath, _ := filepath.Rel(filepath.Join(projectPath, "src", "main", "assets"), assetPath)
	devicePath := "/data/local/tmp/tamk/" + relPath

	deviceDir := filepath.Dir(devicePath)
	mkdirCmd := exec.CommandContext(ctx, "adb", "shell", "mkdir", "-p", deviceDir)
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create device directory: %w", err)
	}

	pushCmd := exec.CommandContext(ctx, "adb", "push", assetPath, devicePath)
	if out, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push asset: %s: %w", string(out), err)
	}

	broadcastCmd := exec.CommandContext(ctx, "adb", "shell", "am", "broadcast",
		"-a", "tamk.ACTION_REFRESH_ASSET",
		"--es", "asset_path", devicePath,
	)
	broadcastCmd.Run()

	return nil
}

// injectSecurityTemplates copies security templates to the project.
func (uc *BuildProjectUseCase) injectSecurityTemplates(ctx context.Context, project *entity.Project, projPath string) {
	securityDir := filepath.Join(projPath, "src", "main", "kotlin",
		strings.ReplaceAll(project.PackageName, ".", "/"), "security")
	os.MkdirAll(securityDir, 0o755)

	// List of security templates to inject
	securityTemplates := []string{
		"GuardianBridge.kt.tmpl",
		"RASPSecurityModule.kt.tmpl",
		"CertificatePinner.kt.tmpl",
		"IntegrityVerifier.kt.tmpl",
		"StringObfuscator.kt.tmpl",
	}

	for _, tmpl := range securityTemplates {
		tmplPath := filepath.Join(uc.cfg.TAMKHome, "templates", "security", "kotlin", tmpl)
		tmplContent, err := os.ReadFile(tmplPath)
		if err != nil {
			logger.Debug("Security template not found", "template", tmpl)
			continue
		}

		// Replace package placeholder
		content := strings.ReplaceAll(string(tmplContent), "{{PACKAGE}}", project.PackageName)

		// Write to security directory
		outName := strings.Replace(tmpl, ".tmpl", "", 1)
		if err := os.WriteFile(filepath.Join(securityDir, outName), []byte(content), 0o644); err != nil {
			logger.Debug("Failed to write security template", "template", outName, "error", err)
		}
	}

	logger.Debug("Security templates injected", "dir", securityDir, "count", len(securityTemplates))
}

func (uc *BuildProjectUseCase) execute(cmd *exec.Cmd) error {
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (uc *BuildProjectUseCase) executeOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func failedResult(phase entity.BuildPhase, msg string) *entity.BuildResult {
	err := &errors.BuildError{Phase: string(phase), Err: fmt.Errorf("%s", msg)}
	logger.Error("Build failed", "phase", string(phase), "error", err)
	return &entity.BuildResult{
		Success:  false,
		Phase:    phase,
		ErrorMsg: err.Error(),
	}
}
