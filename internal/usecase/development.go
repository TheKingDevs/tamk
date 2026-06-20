package usecase

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shadw-Developer/tamk/internal/config"
	"github.com/Shadw-Developer/tamk/internal/domain/repository"
	"github.com/Shadw-Developer/tamk/pkg/errors"
	"github.com/Shadw-Developer/tamk/pkg/logger"
	"github.com/Shadw-Developer/tamk/pkg/watcher"
)

type DevModeUseCase struct {
	cfg      *config.Config
	buildUC  *BuildProjectUseCase
	projRepo repository.ProjectRepository
	watcher  *watcher.Watcher

	password      string
	projectPath   string
	assetsDir     string
	mu            sync.Mutex
	buildCooldown atomic.Int64

	devBridgeInjected bool
	backupPath        string
}

func NewDevModeUseCase(
	cfg *config.Config,
	buildUC *BuildProjectUseCase,
	projRepo repository.ProjectRepository,
) *DevModeUseCase {
	return &DevModeUseCase{
		cfg:      cfg,
		buildUC:  buildUC,
		projRepo: projRepo,
	}
}

func (uc *DevModeUseCase) Start(ctx context.Context, projectPath, password string) error {
	uc.mu.Lock()
	uc.projectPath = projectPath
	uc.password = password
	uc.mu.Unlock()

	if !uc.projRepo.Exists(ctx, projectPath) {
		return fmt.Errorf("no project found at %s: %w", projectPath, errors.ErrProjectNotFound)
	}

	project, err := uc.projRepo.Load(ctx, projectPath)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	if project.Type != "webapp" {
		return fmt.Errorf("dev mode only supports WebApp projects: %w", errors.ErrNotAWebAppProject)
	}

	uc.mu.Lock()
	uc.assetsDir = filepath.Join(projectPath, "src", "main", "assets")
	uc.mu.Unlock()
	if _, err := os.Stat(uc.assetsDir); os.IsNotExist(err) {
		return fmt.Errorf("assets directory not found: %w", errors.ErrBuildFailed)
	}

	if err := uc.injectDevBridge(); err != nil {
		return fmt.Errorf("failed to inject dev bridge: %w", err)
	}

	uc.watcher = watcher.New(uc.onFileChanged, 500*time.Millisecond)
	if err := uc.watcher.Start(uc.assetsDir); err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}

	logger.Success("Dev mode started")
	logger.Info("Watching for changes", "dir", uc.assetsDir)
	logger.Info("Commands: b=rebuild, i=install, s=status, h=help, q=quit")

	return nil
}

func (uc *DevModeUseCase) Stop(ctx context.Context) {
	if uc.watcher != nil {
		uc.watcher.Stop()
	}
	uc.restoreBackup()
	logger.Info("Dev mode stopped")
}

func (uc *DevModeUseCase) onFileChanged(path string) {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".css", ".js":
		logger.Info(fmt.Sprintf("%s changed, HMR update", ext), "file", path)
	default:
		logger.Info("File changed, triggering rebuild", "file", path)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		uc.quickAssetsBuild(ctx)
	}
}

func (uc *DevModeUseCase) quickAssetsBuild(ctx context.Context) {
	now := time.Now().UnixNano()
	cooldown := uc.buildCooldown.Load()
	if now < cooldown {
		return
	}
	if !uc.buildCooldown.CompareAndSwap(cooldown, now+time.Second.Nanoseconds()) {
		return
	}

	logger.Step("Rebuilding assets...")

	uc.mu.Lock()
	projPath := uc.projectPath
	pass := uc.password
	uc.mu.Unlock()

	result, err := uc.buildUC.AssetsOnlyBuild(ctx, BuildInput{
		ProjectPath: projPath,
		Password:    pass,
	})
	if err != nil {
		logger.Error("Rebuild failed", "error", err)
		return
	}

	if result.Success {
		logger.Success("Assets rebuilt", "apk", result.APKPath)
	}
}

func (uc *DevModeUseCase) installAPK() {
	uc.mu.Lock()
	p := uc.projectPath
	uc.mu.Unlock()

	pattern := filepath.Join(p, "*-release.apk")
	matches, _ := filepath.Glob(pattern)
	apkPath := ""
	if len(matches) > 0 {
		apkPath = matches[0]
	} else {
		devPattern := filepath.Join(p, "*-dev.apk")
		devMatches, _ := filepath.Glob(devPattern)
		if len(devMatches) > 0 {
			apkPath = devMatches[0]
		} else {
			logger.Error("No APK found to install")
			return
		}
	}

	cmd := exec.Command("adb", "install", "-r", apkPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		logger.Error("Install failed", "error", string(out))
	} else {
		logger.Success("APK installed")
	}
}

func (uc *DevModeUseCase) injectDevBridge() error {
	uc.mu.Lock()
	aDir := uc.assetsDir
	uc.mu.Unlock()

	indexPath := filepath.Join(aDir, "index.html")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return fmt.Errorf("index.html not found: %w", err)
	}

	uc.backupPath = indexPath + ".tamk_backup"
	os.WriteFile(uc.backupPath, data, 0o644)

	bridgeScript := `<script>
(function() {
    var ws = new WebSocket('ws://localhost:8765');
    ws.onmessage = function(e) {
        var msg = JSON.parse(e.data);
        if (msg.type === 'reload') { location.reload(); }
        else if (msg.type === 'css-update') {
            document.querySelectorAll('link[rel="stylesheet"]').forEach(function(link) {
                link.href = link.href.split('?')[0] + '?t=' + Date.now();
            });
        }
    };
    ws.onopen = function() {
        ws.send(JSON.stringify({type: 'hello', url: window.location.href}));
    };
})();
</script>`

	content := string(data)
	if strings.Contains(content, "</body>") {
		content = strings.Replace(content, "</body>", bridgeScript+"\n</body>", 1)
	} else {
		content += bridgeScript
	}

	os.WriteFile(indexPath, []byte(content), 0o644)
	uc.devBridgeInjected = true

	return nil
}

func (uc *DevModeUseCase) restoreBackup() {
	uc.mu.Lock()
	inj := uc.devBridgeInjected
	backup := uc.backupPath
	assets := uc.assetsDir
	uc.mu.Unlock()

	if !inj {
		return
	}

	if _, err := os.Stat(backup); err == nil {
		indexPath := filepath.Join(assets, "index.html")
		data, _ := os.ReadFile(backup)
		os.WriteFile(indexPath, data, 0o644)
		os.Remove(backup)
		uc.mu.Lock()
		uc.devBridgeInjected = false
		uc.mu.Unlock()
		logger.Info("Index.html restored from backup")
	}
}

func (uc *DevModeUseCase) GetStatus() map[string]any {
	uc.mu.Lock()
	defer uc.mu.Unlock()

	return map[string]any{
		"project":  uc.projectPath,
		"watching": uc.assetsDir,
		"watcher":  uc.watcher != nil && uc.watcher.IsRunning(),
		"bridge":   uc.devBridgeInjected,
	}
}
