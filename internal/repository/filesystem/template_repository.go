package filesystem

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/domain/entity"
)

type TemplateRepository struct {
	cfg *config.Config
}

func NewTemplateRepository(cfg *config.Config) *TemplateRepository {
	return &TemplateRepository{cfg: cfg}
}

func (r *TemplateRepository) LoadTemplate(ctx context.Context, name string) (string, error) {
	return r.LoadTemplateForType(ctx, "", name)
}

func (r *TemplateRepository) LoadTemplateForType(ctx context.Context, projectType, name string) (string, error) {
	if projectType != "" {
		dir := r.cfg.GetTemplateDir(projectType)
		path := filepath.Join(dir, name)
		if !strings.HasSuffix(path, ".tmpl") {
			path += ".tmpl"
		}
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}
	}

	for _, tmplDir := range r.getTemplateDirCandidates() {
		path := filepath.Join(tmplDir, name)
		if !strings.HasSuffix(path, ".tmpl") {
			path += ".tmpl"
		}
		data, err := os.ReadFile(path)
		if err == nil {
			return string(data), nil
		}
	}
	return "", fmt.Errorf("template %s not found", name)
}

func (r *TemplateRepository) LoadTemplateWithFallback(ctx context.Context, name, fallbackType string) (string, error) {
	content, err := r.LoadTemplate(ctx, name)
	if err == nil {
		return content, nil
	}
	fallbackDir := r.cfg.GetTemplateDir(fallbackType)
	path := filepath.Join(fallbackDir, name)
	if !strings.HasSuffix(path, ".tmpl") {
		path += ".tmpl"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("template %s not found in %s or %s fallback", name, name, fallbackType)
	}
	return string(data), nil
}

func (r *TemplateRepository) ApplyPlaceholders(content string, placeholders map[string]string) string {
	for key, val := range placeholders {
		content = strings.ReplaceAll(content, "{{"+key+"}}", val)
	}
	return content
}

func (r *TemplateRepository) WriteTemplate(ctx context.Context, projectPath, destPath, content string) error {
	fullPath := filepath.Join(projectPath, destPath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return os.WriteFile(fullPath, []byte(content), 0o644)
}

func (r *TemplateRepository) GetTemplateDir(ctx context.Context, projectType string) (string, error) {
	dir := r.cfg.GetTemplateDir(projectType)
	info, err := os.Stat(dir)
	if err != nil {
		return "", fmt.Errorf("template directory %s not found: %w", dir, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", dir)
	}
	return dir, nil
}

func (r *TemplateRepository) GetMappings(ctx context.Context, projectType string, isInternal bool) ([]entity.TemplateMapping, error) {
	switch projectType {
	case "webapp":
		return r.getWebAppMappings(isInternal), nil
	case "ui_apk":
		return r.getUIAPKMappings(), nil
	case "console":
		return r.getConsoleMappings(), nil
	default:
		return nil, fmt.Errorf("unknown project type: %s", projectType)
	}
}

func (r *TemplateRepository) getWebAppMappings(internal bool) []entity.TemplateMapping {
	mappings := []entity.TemplateMapping{
		{Dest: "AndroidManifest.xml", Template: "xml/AndroidManifest.xml.tmpl"},
		{Dest: "res/values/strings.xml", Template: "xml/strings.xml.tmpl"},
		{Dest: "res/values/styles.xml", Template: "xml/styles.xml.tmpl"},
		{Dest: "res/drawable/ic_launcher.xml", Template: "xml/icon.xml.tmpl"},
		{Dest: "res/drawable/ic_launcher_round.xml", Template: "xml/icon.xml.tmpl"},
		{Dest: "res/xml/network_security_config.xml", Template: "xml/network_security_config.xml.tmpl"},
		{Dest: "src/main/kotlin/", Template: "kotlin/MainActivity.kt.tmpl"},
		{Dest: ".gitignore", Template: "gitignore_root.tmpl"},
	}
	if internal {
		mappings = append(mappings,
			entity.TemplateMapping{Dest: "src/main/assets/index.html", Template: "index.html.tmpl", Internal: true},
			entity.TemplateMapping{Dest: "src/main/assets/css/styles.css", Template: "css/styles.css.tmpl", Internal: true},
			entity.TemplateMapping{Dest: "src/main/assets/js/app.js", Template: "js/app.js.tmpl", Internal: true},
			entity.TemplateMapping{Dest: "src/main/assets/.gitignore", Template: "gitignore_assets.tmpl", Internal: true},
		)
	}
	return mappings
}

func (r *TemplateRepository) getUIAPKMappings() []entity.TemplateMapping {
	return []entity.TemplateMapping{
		{Dest: "AndroidManifest.xml", Template: "xml/AndroidManifest.xml.tmpl"},
		{Dest: "res/layout/activity_main.xml", Template: "xml/activity_main.xml.tmpl"},
		{Dest: "res/values/strings.xml", Template: "xml/strings.xml.tmpl"},
		{Dest: "res/values/styles.xml", Template: "xml/styles.xml.tmpl"},
		{Dest: "res/mipmap/ic_launcher.xml", Template: "xml/icon.xml.tmpl"},
		{Dest: "res/mipmap/ic_launcher_round.xml", Template: "xml/icon.xml.tmpl"},
		{Dest: "src/main/kotlin/", Template: "kotlin/MainActivity.kt.tmpl"},
	}
}

func (r *TemplateRepository) getConsoleMappings() []entity.TemplateMapping {
	return []entity.TemplateMapping{
		{Dest: "src/Main.kt", Template: "Main.kt.tmpl"},
	}
}

func (r *TemplateRepository) getTemplateDirCandidates() []string {
	return []string{
		r.cfg.GetTemplateDir("webapp"),
		r.cfg.GetTemplateDir("ui_apk"),
		r.cfg.GetTemplateDir("console"),
	}
}
