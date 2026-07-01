package usecase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TheKingDevs/tamk/internal/tools"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

type RunUseCase struct {
	tools tools.ToolManager
}

func NewRunUseCase(tools tools.ToolManager) *RunUseCase {
	return &RunUseCase{tools: tools}
}

type RunInput struct {
	ProjectPath string
	FileName    string
}

type RunOutput struct {
	ExitCode int
	Output   string
}

func (uc *RunUseCase) Execute(ctx context.Context, input RunInput) (*RunOutput, error) {
	// Find the Kotlin/Java file
	filePath, err := uc.findSourceFile(input.ProjectPath, input.FileName)
	if err != nil {
		return nil, err
	}

	logger.Step(fmt.Sprintf("Compiling %s", filePath))

	// Get kotlinc path
	kotlincPath, err := uc.tools.KotlinCompiler(ctx)
	if err != nil {
		return nil, fmt.Errorf("kotlinc not found: %w", err)
	}

	// Create temp directory for compiled classes
	tmpDir := filepath.Join(input.ProjectPath, ".tamk-run")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Get SDK jar path for classpath
	sdkPath, err := uc.tools.SDKJar()
	if err != nil {
		logger.Warn(fmt.Sprintf("SDK jar not found, compiling without classpath: %v", err))
		sdkPath = ""
	}

	// Compile Kotlin file
	compileArgs := []string{
		filePath,
		"-d", tmpDir,
	}
	if sdkPath != "" {
		compileArgs = append(compileArgs, "-cp", sdkPath)
	}

	compileCmd := exec.CommandContext(ctx, kotlincPath, compileArgs...)
	output, err := compileCmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("compilation failed: %s\n%s", err.Error(), string(output))
	}

	logger.Success("Compilation successful")

	// Find the main class
	mainClass, err := uc.findMainClass(tmpDir, filePath)
	if err != nil {
		return nil, err
	}

	logger.Step(fmt.Sprintf("Running %s", mainClass))

	// Find kotlin runtime
	kotlinPath := filepath.Dir(kotlincPath)
	runtimeJar := uc.findKotlinRuntime(kotlinPath)

	// Execute the compiled class
	runArgs := []string{"-cp", tmpDir}
	if runtimeJar != "" {
		runArgs[1] = tmpDir + string(os.PathListSeparator) + runtimeJar
	}
	runArgs = append(runArgs, mainClass)

	runCmd := exec.CommandContext(ctx, "kotlin", runArgs...)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin

	if err := runCmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return &RunOutput{ExitCode: exitErr.ExitCode()}, nil
		}
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	return &RunOutput{ExitCode: 0}, nil
}

func (uc *RunUseCase) findSourceFile(projectPath, fileName string) (string, error) {
	if fileName == "" {
		// Look for Main.kt or Main.java in src/
		candidates := []string{
			filepath.Join(projectPath, "src", "Main.kt"),
			filepath.Join(projectPath, "src", "main.kt"),
			filepath.Join(projectPath, "src", "Main.java"),
			filepath.Join(projectPath, "src", "main.java"),
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
		return "", fmt.Errorf("no source file found in %s/src/", projectPath)
	}

	// Use provided filename
	filePath := filepath.Join(projectPath, fileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Try in src/
		filePath = filepath.Join(projectPath, "src", fileName)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", fileName)
		}
	}

	if !strings.HasSuffix(filePath, ".kt") && !strings.HasSuffix(filePath, ".java") {
		return "", fmt.Errorf("unsupported file type: %s (only .kt and .java supported)", filePath)
	}

	return filePath, nil
}

func (uc *RunUseCase) findMainClass(classDir, sourceFile string) (string, error) {
	// Extract class name from source file
	baseName := filepath.Base(sourceFile)
	classDir2 := strings.TrimSuffix(strings.TrimSuffix(baseName, ".kt"), ".java")

	// Look for .class file
	var mainClass string
	err := filepath.Walk(classDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".class") {
			// Check if it's the main class (has main method)
			classFile := strings.TrimPrefix(path, classDir+string(os.PathSeparator))
			classFile = strings.TrimSuffix(classFile, ".class")
			classFile = strings.ReplaceAll(classFile, string(os.PathSeparator), ".")

			if strings.Contains(classFile, classDir2) || strings.HasSuffix(classFile, "Kt") {
				mainClass = classFile
				return filepath.SkipDir
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if mainClass == "" {
		return "", fmt.Errorf("main class not found in %s", classDir)
	}

	return mainClass, nil
}

func (uc *RunUseCase) findKotlinRuntime(kotlinBinDir string) string {
	// kotlin binary is usually in bin/, lib is in ../lib/
	libDir := filepath.Join(filepath.Dir(kotlinBinDir), "lib")
	if info, err := os.Stat(libDir); err == nil && info.IsDir() {
		// Look for kotlin-stdlib.jar
		entries, _ := os.ReadDir(libDir)
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "kotlin-stdlib") && strings.HasSuffix(e.Name(), ".jar") {
				return filepath.Join(libDir, e.Name())
			}
		}
	}
	return ""
}
