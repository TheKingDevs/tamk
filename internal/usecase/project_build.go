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
	"github.com/TheKingDevs/tamk/pkg/errors"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

func apkFilename(project *entity.Project, env string) string {
	slug := strings.ToLower(strings.ReplaceAll(project.Name, " ", "-"))
	return fmt.Sprintf("%s-%s-%s.apk", slug, project.Version, env)
}

type BuildProjectUseCase struct {
	cfg       *config.Config
	buildRepo repository.BuildRepository
	projRepo  repository.ProjectRepository
}

func NewBuildProjectUseCase(
	cfg *config.Config,
	buildRepo repository.BuildRepository,
	projRepo repository.ProjectRepository,
) *BuildProjectUseCase {
	return &BuildProjectUseCase{
		cfg:       cfg,
		buildRepo: buildRepo,
		projRepo:  projRepo,
	}
}

type BuildInput struct {
	ProjectPath string
	Password    string
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

	if _, err := os.Stat(uc.cfg.SDKPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("sdk not found: %w", errors.ErrSDKNotFound)
	}

	if input.Password != "" && len(input.Password) < 6 {
		return nil, fmt.Errorf("keystore password too short: %w", errors.ErrKeystoreInvalidPass)
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

	keystorePath := filepath.Join(input.ProjectPath, "secret", "project.keystore")
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		keystorePath = uc.cfg.Keystore
	}

	signCmd := exec.CommandContext(ctx, "apksigner", "sign",
		"--ks", keystorePath,
		"--ks-pass", "pass:"+input.Password,
		"--out", devAPK+".signed",
		devAPK,
	)
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
	logger.Step("Compiling resources (AAPT2)...")
	if out, err := uc.executeOutput(exec.CommandContext(ctx, "aapt2", "compile", "--dir",
		filepath.Join(projPath, "res"), "-o", resZip)); err != nil {
		return failedResult(entity.BuildPhaseAAPT2Compile, string(out))
	}

	logger.Step("Linking resources (AAPT2)...")
	genDir := filepath.Join(projPath, "gen")
	os.MkdirAll(genDir, 0o755)
	linkArgs := []string{
		"link",
		"-I", uc.cfg.SDKPath,
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
	if out, err := uc.executeOutput(exec.CommandContext(ctx, "aapt2", linkArgs...)); err != nil {
		return failedResult(entity.BuildPhaseAAPT2Link, string(out))
	}

	kotlinDir := filepath.Join(projPath, "src", "main", "kotlin")

	logger.Step("Compiling Kotlin sources...")
	os.MkdirAll(objDir, 0o755)
	kotlinCmd := exec.CommandContext(ctx, "kotlinc",
		kotlinDir, genDir,
		"-cp", uc.cfg.SDKPath,
		"-d", objDir,
	)
	if out, err := uc.executeOutput(kotlinCmd); err != nil {
		return failedResult(entity.BuildPhaseKotlinCompile, string(out))
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
	args := []string{"--lib", uc.cfg.SDKPath, "--release", "--output", projPath}
	args = append(args, classFiles...)
	d8Cmd := exec.CommandContext(ctx, "d8", args...)
	if out, err := uc.executeOutput(d8Cmd); err != nil {
		return failedResult(entity.BuildPhaseD8, string(out))
	}

	logger.Step("Packaging DEX into APK...")
	if out, err := uc.executeOutput(exec.CommandContext(ctx, "zip", "-j", apkPath,
		filepath.Join(projPath, "classes.dex"))); err != nil {
		return failedResult(entity.BuildPhasePackageDEX, string(out))
	}

	logger.Step("Aligning (zipalign)...")
	if out, err := uc.executeOutput(exec.CommandContext(ctx, "zipalign", "-f", "4", apkPath, unsignedPath)); err != nil {
		return failedResult(entity.BuildPhaseZipalign, string(out))
	}

	keystorePath := filepath.Join(projPath, "secret", "project.keystore")
	if _, err := os.Stat(keystorePath); os.IsNotExist(err) {
		keystorePath = uc.cfg.Keystore
	}

	logger.Step("Signing APK...")
	signCmd := exec.CommandContext(ctx, "apksigner", "sign",
		"--ks", keystorePath,
		"--ks-pass", "pass:"+input.Password,
		"--out", finalPath,
		unsignedPath,
	)
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

	logger.Success("Build complete", "apk", finalPath)
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
	logger.Error("Build failed", "phase", string(phase), "error", msg)
	return &entity.BuildResult{
		Success:  false,
		Phase:    phase,
		ErrorMsg: err.Error(),
	}
}
