package usecase

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/domain/entity"
	"github.com/TheKingDevs/tamk/internal/domain/repository"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

type CreateProjectUseCase struct {
	cfg      *config.Config
	projRepo repository.ProjectRepository
	tmplRepo repository.TemplateRepository
}

func NewCreateProjectUseCase(
	cfg *config.Config,
	projRepo repository.ProjectRepository,
	tmplRepo repository.TemplateRepository,
) *CreateProjectUseCase {
	return &CreateProjectUseCase{
		cfg:      cfg,
		projRepo: projRepo,
		tmplRepo: tmplRepo,
	}
}

type CreateProjectInput struct {
	Name     string
	Type     entity.ProjectType
	Version  string
	Author   string
	Password string
	WebURL   string
	WebMode  entity.WebContentMode
}

type CreateProjectOutput struct {
	ProjectPath string
	Project     *entity.Project
}

func (uc *CreateProjectUseCase) Execute(ctx context.Context, input CreateProjectInput) (*CreateProjectOutput, error) {
	projectPath := uc.cfg.GetProjectDir(input.Name)

	if uc.projRepo.Exists(ctx, projectPath) {
		return nil, fmt.Errorf("project '%s' already exists at %s", input.Name, projectPath)
	}

	packageName := uc.cfg.DetectPackageName(input.Author, input.Name)

	project := &entity.Project{
		Name:        input.Name,
		Type:        input.Type,
		Version:     input.Version,
		Author:      input.Author,
		PackageName: packageName,
		WebURL:      input.WebURL,
		WebMode:     input.WebMode,
		MinSDK:      21,
		TargetSDK:   30,
	}

	switch input.Type {
	case entity.ProjectTypeWebApp:
		if err := uc.createWebAppStructure(ctx, project, projectPath, input); err != nil {
			return nil, fmt.Errorf("failed to create webapp: %w", err)
		}
	case entity.ProjectTypeUIAPK:
		if err := uc.createUIAPKStructure(ctx, project, projectPath, input); err != nil {
			return nil, fmt.Errorf("failed to create UI APK: %w", err)
		}
	case entity.ProjectTypeConsole:
		// Console projects don't have web-related fields
		project.WebURL = ""
		project.WebMode = ""
		project.MinSDK = 0
		project.TargetSDK = 0
		if err := uc.createConsoleStructure(ctx, project, projectPath, input); err != nil {
			return nil, fmt.Errorf("failed to create console app: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported project type: %s", input.Type)
	}

	if input.Password != "" {
		if err := uc.generateKeystore(ctx, project, projectPath, input.Password); err != nil {
			return nil, fmt.Errorf("failed to generate keystore: %w", err)
		}
	}

	project.KeystorePath = filepath.Join(projectPath, "secret", "project.keystore")

	if err := uc.projRepo.Save(ctx, project, projectPath); err != nil {
		return nil, fmt.Errorf("failed to save project config: %w", err)
	}

	logger.Success("Project created successfully", "path", projectPath)

	return &CreateProjectOutput{
		ProjectPath: projectPath,
		Project:     project,
	}, nil
}

func (uc *CreateProjectUseCase) createWebAppStructure(ctx context.Context, project *entity.Project, projectPath string, input CreateProjectInput) error {
	isInternal := input.WebMode == entity.WebContentInternal

	dirs := []string{
		"secret",
		"res/values",
		"res/drawable",
		"res/xml",
		"src/main/kotlin/" + strings.ReplaceAll(project.PackageName, ".", "/"),
	}
	if isInternal {
		dirs = append(dirs, "src/main/assets/css", "src/main/assets/js")
	}

	for _, d := range dirs {
		path := config.SecurePath(projectPath, d)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}

	placeholders := uc.makePlaceholders(project, input)

	mappings, err := uc.tmplRepo.GetMappings(ctx, "webapp", isInternal)
	if err != nil {
		return err
	}

	for _, m := range mappings {
		tmplContent, err := uc.tmplRepo.LoadTemplateForType(ctx, string(project.Type), m.Template)
		if err != nil {
			if m.Internal {
				logger.Warn("Template not found (skipping)", "template", m.Template)
				continue
			}
			return fmt.Errorf("required template %s not found: %w", m.Template, err)
		}

		content := uc.tmplRepo.ApplyPlaceholders(tmplContent, placeholders)

		dest := m.Dest
		if strings.HasPrefix(dest, "src/main/kotlin/") {
			dest = "src/main/kotlin/" + strings.ReplaceAll(project.PackageName, ".", "/") + "/MainActivity.kt"
		}

		if err := uc.tmplRepo.WriteTemplate(ctx, projectPath, dest, content); err != nil {
			return fmt.Errorf("failed to write %s: %w", dest, err)
		}
	}

	return nil
}

func (uc *CreateProjectUseCase) createUIAPKStructure(ctx context.Context, project *entity.Project, projectPath string, input CreateProjectInput) error {
	dirs := []string{
		"secret",
		"res/layout",
		"res/values",
		"res/mipmap",
		"src/main/kotlin/" + strings.ReplaceAll(project.PackageName, ".", "/"),
	}
	for _, d := range dirs {
		path := config.SecurePath(projectPath, d)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}

	placeholders := uc.makePlaceholders(project, input)

	mappings, err := uc.tmplRepo.GetMappings(ctx, "ui_apk", false)
	if err != nil {
		return err
	}

	for _, m := range mappings {
		tmplContent, err := uc.tmplRepo.LoadTemplateForType(ctx, string(project.Type), m.Template)
		if err != nil {
			return fmt.Errorf("template %s not found: %w", m.Template, err)
		}
		content := uc.tmplRepo.ApplyPlaceholders(tmplContent, placeholders)

		dest := m.Dest
		if strings.HasPrefix(dest, "src/main/kotlin/") {
			dest = "src/main/kotlin/" + strings.ReplaceAll(project.PackageName, ".", "/") + "/MainActivity.kt"
		}

		if err := uc.tmplRepo.WriteTemplate(ctx, projectPath, dest, content); err != nil {
			return err
		}
	}

	return nil
}

func (uc *CreateProjectUseCase) createConsoleStructure(ctx context.Context, project *entity.Project, projectPath string, input CreateProjectInput) error {
	dirs := []string{"src", "libs"}
	for _, d := range dirs {
		path := config.SecurePath(projectPath, d)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return err
		}
	}

	placeholders := uc.makePlaceholders(project, input)

	tmplContent, err := uc.tmplRepo.LoadTemplateForType(ctx, "console", "Main.kt.tmpl")
	if err != nil {
		logger.Warn("Template Main.kt.tmpl not found, using fallback")
		tmplContent = `fun main() {
    println("Hello from {{NAME}} v{{VERSION}} by {{AUTHOR}}!")
}
`
	}
	content := uc.tmplRepo.ApplyPlaceholders(tmplContent, placeholders)

	return uc.tmplRepo.WriteTemplate(ctx, projectPath, "src/Main.kt", content)
}

func (uc *CreateProjectUseCase) makePlaceholders(project *entity.Project, input CreateProjectInput) map[string]string {
	webURL := input.WebURL
	if webURL == "" {
		webURL = "file:///android_asset/index.html"
	}

	devMode := "false"
	if input.WebMode == entity.WebContentExternal {
		devMode = "true"
	}

	return map[string]string{
		"NAME":         project.Name,
		"PACKAGE":      project.PackageName,
		"VERSION":      project.Version,
		"AUTHOR":       project.Author,
		"WEB_URL":      webURL,
		"DEV_MODE":     devMode,
		"MIN_SDK":      fmt.Sprintf("%d", project.MinSDK),
		"TARGET_SDK":   fmt.Sprintf("%d", project.TargetSDK),
		"TAMK_VERSION": config.Version,
		"DEV_PORT":     "8765",
	}
}

func (uc *CreateProjectUseCase) generateKeystore(ctx context.Context, project *entity.Project, projectPath, password string) error {
	secretDir := filepath.Join(projectPath, "secret")
	os.MkdirAll(secretDir, 0o755)

	keystorePath := filepath.Join(secretDir, "project.keystore")

	args := []string{
		"-genkey",
		"-keystore", keystorePath,
		"-alias", project.Name,
		"-keyalg", "RSA",
		"-keysize", "2048",
		"-validity", "10000",
		"-storepass", password,
		"-keypass", password,
		"-dname", fmt.Sprintf("CN=%s, OU=TAMK, O=TAMK, L=Unknown, ST=Unknown, C=UN",
			project.Author),
	}

	cmd := exec.CommandContext(ctx, "keytool", args...)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to generate keystore: %w", err)
	}

	if err := os.Chmod(keystorePath, 0o600); err != nil {
		return fmt.Errorf("failed to set keystore permissions: %w", err)
	}

	logger.Success("Keystore generated", "path", keystorePath)
	return nil
}
