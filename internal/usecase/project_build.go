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

func artifactFilename(project *entity.Project, env, ext string) string {
	slug := strings.ToLower(strings.ReplaceAll(project.Name, " ", "-"))
	return fmt.Sprintf("%s-%s-%s.%s", slug, project.Version, env, ext)
}

func apkFilename(project *entity.Project, env string) string {
	return artifactFilename(project, env, "apk")
}

func aabFilename(project *entity.Project, env string) string {
	return artifactFilename(project, env, "aab")
}

// resolveKeystorePath finds the keystore: project-local first, then fallback to config.
// extractMinSdkVersion reads the minSdkVersion from AndroidManifest.xml.
func extractMinSdkVersion(manifestPath string) string {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return "21"
	}
	content := string(data)

	idx := strings.Index(content, "minSdkVersion")
	if idx == -1 {
		return "21"
	}
	rest := content[idx:]
	start := strings.IndexAny(rest, "\"")
	if start == -1 {
		return "21"
	}
	rest = rest[start+1:]
	end := strings.IndexAny(rest, "\"")
	if end == -1 {
		return "21"
	}
	return rest[:end]
}

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
		logger.Debug(fmt.Sprintf("Keystore validation failed: %s", string(output)))
		return fmt.Errorf("keystore password incorrect: %w", errors.ErrKeystoreInvalidPass)
	}

	// Verify we actually got a valid response
	if !strings.Contains(string(output), "Keystore type") && !strings.Contains(string(output), "Entry count") {
		logger.Debug(fmt.Sprintf("Keystore validation failed: unexpected output: %s", string(output)))
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
	Targets     []entity.BuildTarget
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
		allExist := true
		targets := input.Targets
		if len(targets) == 0 {
			targets = []entity.BuildTarget{entity.BuildTargetAPK}
		}
		for _, t := range targets {
			switch t {
			case entity.BuildTargetAPK:
				p := filepath.Join(input.ProjectPath, apkFilename(project, "release"))
				if _, err := os.Stat(p); err != nil {
					allExist = false
					logger.Warn(fmt.Sprintf("APK missing: %s, forcing rebuild", p))
				}
			case entity.BuildTargetAAB:
				p := filepath.Join(input.ProjectPath, aabFilename(project, "release"))
				if _, err := os.Stat(p); err != nil {
					allExist = false
					logger.Warn(fmt.Sprintf("AAB missing: %s, forcing rebuild", p))
				}
			}
		}
		if allExist {
			logger.Info("Nothing changed, skipping build")
			expectedAPK := filepath.Join(input.ProjectPath, apkFilename(project, "release"))
			return &entity.BuildResult{Success: true, APKPath: expectedAPK}, nil
		}
	}

	// Apply Guardian security if enabled
	if input.Guardian {
		project.Security = entity.SecurityConfigForLevel(entity.SecurityLevelStandard)
		logger.Step("Applying Guardian security protection...")

		// Encrypt assets
		encryptor := &AssetEncryptor{}
		if err := encryptor.EncryptAssets(input.ProjectPath, input.Password); err != nil {
			logger.Warn(fmt.Sprintf("Asset encryption failed: %v", err))
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
		logger.Error(fmt.Sprintf("Keystore validation failed: %v", err))
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
	resZip := filepath.Join(projPath, "res.zip")

	// Determine targets
	targets := input.Targets
	if len(targets) == 0 {
		targets = []entity.BuildTarget{entity.BuildTargetAPK}
	}

	wantAPK := false
	wantAAB := false
	for _, t := range targets {
		switch t {
		case entity.BuildTargetAPK:
			wantAPK = true
		case entity.BuildTargetAAB:
			wantAAB = true
		}
	}

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

	apksignerPath, err := uc.tools.ApkSigner(ctx)
	if err != nil {
		return failedResult(entity.BuildPhaseApkSign, err.Error())
	}

	sdkPath, err := uc.tools.SDKJar()
	if err != nil {
		return failedResult(entity.BuildPhaseAAPT2Link, err.Error())
	}

	// Shared: compile resources
	logger.Step("Compiling resources (AAPT2)...")
	if out, err := uc.executeOutput(exec.CommandContext(ctx, aapt2Path, "compile", "--dir",
		filepath.Join(projPath, "res"), "-o", resZip)); err != nil {
		return failedResult(entity.BuildPhaseAAPT2Compile, string(out))
	}

	assetsDir := filepath.Join(projPath, "src", "main", "assets")
	kotlinDir := filepath.Join(projPath, "src", "main", "kotlin")

	// If Guardian is enabled, copy security templates to project
	if input.Guardian {
		uc.injectSecurityTemplates(ctx, project, projPath)
	}

	result := &entity.BuildResult{Success: true}

	// Build APK target
	if wantAPK {
		r := uc.buildAPK(ctx, project, input, aapt2Path, kotlincPath, d8Path, apksignerPath, sdkPath, resZip, assetsDir, kotlinDir)
		if !r.Success {
			return r
		}
		result.APKPath = r.APKPath
	}

	// Build AAB target
	if wantAAB {
		r := uc.buildAAB(ctx, project, input, aapt2Path, kotlincPath, d8Path, apksignerPath, sdkPath, resZip, assetsDir, kotlinDir)
		if !r.Success {
			return r
		}
		result.AABPath = r.AABPath
	}

	// Cleanup shared temp files
	os.Remove(resZip)

	if result.APKPath != "" {
		logger.Success(fmt.Sprintf("APK ready: %s", result.APKPath))
	}
	if result.AABPath != "" {
		logger.Success(fmt.Sprintf("AAB ready: %s", result.AABPath))
	}
	return result
}

func (uc *BuildProjectUseCase) buildAPK(
	ctx context.Context, project *entity.Project, input BuildInput,
	aapt2Path, kotlincPath, d8Path, apksignerPath, sdkPath,
	resZip, assetsDir, kotlinDir string,
) *entity.BuildResult {
	projPath := input.ProjectPath
	apkPath := filepath.Join(projPath, "app.apk")
	unsignedPath := filepath.Join(projPath, "app-unsigned.apk")
	finalPath := filepath.Join(projPath, apkFilename(project, "release"))
	objDir := filepath.Join(projPath, "obj")

	zipalignPath, err := uc.tools.Zipalign(ctx)
	if err != nil {
		return failedResult(entity.BuildPhaseZipalign, err.Error())
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
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		linkArgs = append(linkArgs, "-A", assetsDir)
	}
	if out, err := uc.executeOutput(exec.CommandContext(ctx, aapt2Path, linkArgs...)); err != nil {
		return failedResult(entity.BuildPhaseAAPT2Link, string(out))
	}

	r := uc.compileAndDex(ctx, project, projPath, kotlincPath, d8Path, sdkPath, objDir, genDir, kotlinDir)
	if !r.Success {
		return r
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

	os.RemoveAll(objDir)
	os.Remove(apkPath)
	os.Remove(unsignedPath)
	classesDex := filepath.Join(projPath, "classes.dex")
	if _, err := os.Stat(classesDex); err == nil {
		os.Remove(classesDex)
	}
	if _, err := os.Stat(genDir); err == nil {
		os.RemoveAll(genDir)
	}

	return &entity.BuildResult{Success: true, APKPath: finalPath}
}

func (uc *BuildProjectUseCase) buildAAB(
	ctx context.Context, project *entity.Project, input BuildInput,
	aapt2Path, kotlincPath, d8Path, apksignerPath, sdkPath,
	resZip, assetsDir, kotlinDir string,
) *entity.BuildResult {
	projPath := input.ProjectPath
	protoAPK := filepath.Join(projPath, "base.apk")
	moduleDir := filepath.Join(projPath, "assets", "cache", "aab-module")
	moduleZip := filepath.Join(projPath, "module.zip")
	aabPath := filepath.Join(projPath, "bundle.aab")
	finalPath := filepath.Join(projPath, aabFilename(project, "release"))
	objDir := filepath.Join(projPath, "obj")

	logger.Step("Linking resources in proto format (AAPT2)...")
	genDir := filepath.Join(projPath, "gen")
	os.MkdirAll(genDir, 0o755)
	linkArgs := []string{
		"link", "--proto-format",
		"-I", sdkPath,
		"--manifest", filepath.Join(projPath, "AndroidManifest.xml"),
		"-o", protoAPK,
		"--java", genDir,
		resZip,
		"--auto-add-overlay",
	}
	if info, err := os.Stat(assetsDir); err == nil && info.IsDir() {
		linkArgs = append(linkArgs, "-A", assetsDir)
	}
	if out, err := uc.executeOutput(exec.CommandContext(ctx, aapt2Path, linkArgs...)); err != nil {
		return failedResult(entity.BuildPhaseAAPT2Link, string(out))
	}

	r := uc.compileAndDex(ctx, project, projPath, kotlincPath, d8Path, sdkPath, objDir, genDir, kotlinDir)
	if !r.Success {
		return r
	}

	// Build AAB module structure manually and sign directly,
	// avoiding bundletool which requires complex proto-APK restructuring.
	// An AAB is a ZIP with internal structure:
	//   base/manifest/AndroidManifest.xml
	//   base/dex/classes.dex
	//   base/resources.pb
	//   base/res/...
	//   base/assets/...

	logger.Step("Assembling AAB module structure...")
	os.RemoveAll(moduleDir)
	baseDir := filepath.Join(moduleDir, "base")
	os.MkdirAll(filepath.Join(baseDir, "manifest"), 0o755)
	os.MkdirAll(filepath.Join(baseDir, "dex"), 0o755)

	// Extract proto APK into module dir
	if out, err := uc.executeOutput(exec.CommandContext(ctx, "unzip", "-o", protoAPK, "-d", baseDir)); err != nil {
		return failedResult(entity.BuildPhaseAABBuild, string(out))
	}

	// Rearrange AndroidManifest into manifest/
	manifestSrc := filepath.Join(baseDir, "AndroidManifest.xml")
	manifestDst := filepath.Join(baseDir, "manifest", "AndroidManifest.xml")
	if _, err := os.Stat(manifestSrc); err == nil {
		if err := os.Rename(manifestSrc, manifestDst); err != nil {
			return failedResult(entity.BuildPhaseAABBuild, err.Error())
		}
	}

	// Move classes.dex into dex/
	dexSrc := filepath.Join(projPath, "classes.dex")
	dexDst := filepath.Join(baseDir, "dex", "classes.dex")
	if _, err := os.Stat(dexSrc); err == nil {
		if err := os.Rename(dexSrc, dexDst); err != nil {
			return failedResult(entity.BuildPhaseAABBuild, err.Error())
		}
	}

	// Zip module content (baseDir) into module.zip for bundletool
	os.Remove(moduleZip)
	if err := zipDir(baseDir, moduleZip); err != nil {
		return failedResult(entity.BuildPhaseAABBuild, err.Error())
	}

	// Build AAB via bundletool
	logger.Step("Building Android App Bundle...")
	os.Remove(aabPath)
	bundletoolPath, err := uc.tools.Bundletool(ctx)
	if err != nil {
		return failedResult(entity.BuildPhaseAABBuild, err.Error())
	}
	btArgs := strings.Fields(bundletoolPath)
	btArgs = append(btArgs, "build-bundle",
		"--modules", moduleZip,
		"--output", aabPath,
	)
	btCmd := exec.CommandContext(ctx, btArgs[0], btArgs[1:]...)
	if out, err := uc.executeOutput(btCmd); err != nil {
		return failedResult(entity.BuildPhaseAABBuild, string(out))
	}

	keystorePath := resolveKeystorePath(projPath, uc.cfg.Keystore)

	logger.Step("Signing AAB...")
	minSdk := extractMinSdkVersion(filepath.Join(projPath, "AndroidManifest.xml"))
	signArgs := strings.Fields(apksignerPath)
	signArgs = append(signArgs, "sign",
		"--ks", keystorePath,
		"--ks-pass", "pass:"+input.Password,
		"--min-sdk-version", minSdk,
		"--out", finalPath,
		aabPath,
	)
	signCmd := exec.CommandContext(ctx, signArgs[0], signArgs[1:]...)
	if out, err := uc.executeOutput(signCmd); err != nil {
		return failedResult(entity.BuildPhaseApkSign, string(out))
	}

	os.RemoveAll(objDir)
	os.Remove(protoAPK)
	os.Remove(aabPath)
	os.Remove(moduleZip)
	os.RemoveAll(moduleDir)
	if _, err := os.Stat(genDir); err == nil {
		os.RemoveAll(genDir)
	}

	return &entity.BuildResult{Success: true, AABPath: finalPath}
}

// compileAndDex compiles Kotlin sources and converts to DEX. Shared between APK and AAB builds.
func (uc *BuildProjectUseCase) compileAndDex(
	ctx context.Context, project *entity.Project, projPath,
	kotlincPath, d8Path, sdkPath, objDir, genDir, kotlinDir string,
) *entity.BuildResult {
	logger.Step("Compiling Kotlin sources...")
	os.MkdirAll(objDir, 0o755)

	classpath := sdkPath
	libMgr := NewLibraryManager(uc.cfg)
	if libCP, err := libMgr.GetClasspath(); err == nil && libCP != "" {
		classpath = sdkPath + string(os.PathListSeparator) + libCP
		logger.Debug(fmt.Sprintf("Library classpath added: %s", classpath))
	}

	kotlinCmd := exec.CommandContext(ctx, kotlincPath,
		kotlinDir, genDir,
		"-cp", classpath,
		"-d", objDir,
	)
	if out, err := uc.executeOutput(kotlinCmd); err != nil {
		logger.Debug(fmt.Sprintf("Kotlin compile output: %s", string(out)))
		return failedResult(entity.BuildPhaseKotlinCompile, string(out))
	}

	// Obfuscate code with ProGuard if Guardian is enabled
	if project.Security.Level != "" && project.Security.Level != entity.SecurityLevelNone {
		obfuscator := NewProGuardObfuscator()
		if obfuscator.IsAvailable() {
			if err := obfuscator.Obfuscate(ctx, objDir, sdkPath, ""); err != nil {
				logger.Warn(fmt.Sprintf("ProGuard obfuscation failed: %v", err))
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
	if libCP, err := libMgr.GetClasspath(); err == nil && libCP != "" {
		for _, jar := range strings.Split(libCP, string(os.PathListSeparator)) {
			args = append(args, "--lib", jar)
		}
	}
	args = append(args, classFiles...)
	d8Cmd := exec.CommandContext(ctx, d8Path, args...)
	if out, err := uc.executeOutput(d8Cmd); err != nil {
		return failedResult(entity.BuildPhaseD8, string(out))
	}

	return &entity.BuildResult{Success: true}
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
			logger.Debug(fmt.Sprintf("Security template not found: %s", tmpl))
			continue
		}

		// Replace package placeholder
		content := strings.ReplaceAll(string(tmplContent), "{{PACKAGE}}", project.PackageName)

		// Write to security directory
		outName := strings.Replace(tmpl, ".tmpl", "", 1)
		if err := os.WriteFile(filepath.Join(securityDir, outName), []byte(content), 0o644); err != nil {
			logger.Debug(fmt.Sprintf("Failed to write security template %s: %v", outName, err))
		}
	}

	logger.Debug(fmt.Sprintf("Security templates injected to %s: %d templates", securityDir, len(securityTemplates)))
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
	logger.Error(fmt.Sprintf("Build failed at phase %s: %s", string(phase), err))
	return &entity.BuildResult{
		Success:  false,
		Phase:    phase,
		ErrorMsg: err.Error(),
	}
}
