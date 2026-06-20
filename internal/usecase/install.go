package usecase

import (
	"context"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/TheKingDevs/tamk/pkg/logger"
	"github.com/TheKingDevs/tamk/pkg/qrcode"
)

type InstallOutput struct {
	URL      string
	Port     int
	APKPath  string
	ServeDir string
}

type InstallUseCase struct{}

func NewInstallUseCase() *InstallUseCase {
	return &InstallUseCase{}
}

func findAPK(projectPath string) (string, error) {
	patterns := []string{"*-release.apk", "*-dev.apk"}
	for _, p := range patterns {
		pattern := filepath.Join(projectPath, p)
		matches, err := filepath.Glob(pattern)
		if err == nil && len(matches) > 0 {
			return matches[0], nil
		}
	}
	return "", fmt.Errorf("APK not found in %s", projectPath)
}

func (uc *InstallUseCase) Serve(ctx context.Context, projectPath string, port int) (*InstallOutput, error) {
	apkPath, err := findAPK(projectPath)
	if err != nil {
		return nil, err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		listener, err = net.Listen("tcp", ":0")
		if err != nil {
			return nil, fmt.Errorf("failed to listen: %w", err)
		}
	}

	actualPort := listener.Addr().(*net.TCPAddr).Port
	serveDir := filepath.Dir(apkPath)
	apkName := filepath.Base(apkPath)

	localIP := getLocalIP()
	url := fmt.Sprintf("http://localhost:%d", actualPort)
	if localIP != "" {
		url = fmt.Sprintf("http://%s:%d", localIP, actualPort)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			renderInstallPage(w, apkName, url, actualPort)
			return
		}
		if r.URL.Path == "/"+apkName {
			w.Header().Set("Content-Type", "application/vnd.android.package-archive")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, apkName))
			http.ServeFile(w, r, apkPath)
			return
		}
		http.NotFound(w, r)
	})

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Install shutdown panicked", "recover", r)
			}
		}()
		<-ctx.Done()
		server.Close()
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Install server panicked", "recover", r)
			}
		}()
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			logger.Error("Install server error", "error", err)
		}
	}()

	fmt.Println()
	fmt.Printf("  \033[1;36m📦 APK:\033[0m        %s\n", apkPath)
	fmt.Printf("  \033[1;36m🌐 URL:\033[0m         %s\n", url)
	fmt.Println()
	fmt.Println("  \033[1;33mScan the QR code below with your phone\033[0m")
	fmt.Println("  \033[1;33mto download and install the APK.\033[0m")
	fmt.Println()

	if err := qrcode.PrintTerminal(url); err != nil {
		logger.Warn("QR code display failed", "error", err)
	}

	fmt.Println()
	fmt.Println("  \033[90mPress Ctrl+C to stop the server.\033[0m")
	fmt.Println()

	return &InstallOutput{
		URL:      url,
		Port:     actualPort,
		APKPath:  apkPath,
		ServeDir: serveDir,
	}, nil
}

func renderInstallPage(w http.ResponseWriter, apkName, url string, port int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	page := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>T.A.M.K — APK Download</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #0d1117; color: #c9d1d9;
    display: flex; flex-direction: column; align-items: center;
    justify-content: center; min-height: 100vh; padding: 2rem;
  }
  .card {
    background: #161b22; border: 1px solid #30363d;
    border-radius: 12px; padding: 3rem; max-width: 520px;
    text-align: center;
  }
  h1 { font-size: 1.5rem; margin-bottom: 0.5rem; color: #58a6ff; }
  p { color: #8b949e; margin-bottom: 1.5rem; word-break: break-all; }
  .apk-icon { font-size: 4rem; margin-bottom: 1rem; }
  a.button {
    display: inline-block; padding: 0.875rem 2rem;
    background: #238636; color: #fff; text-decoration: none;
    border-radius: 8px; font-size: 1.1rem; font-weight: 600;
    transition: background 0.2s;
  }
  a.button:hover { background: #2ea043; }
  .url { margin-top: 1.5rem; font-size: 0.85rem; color: #8b949e; }
  .url code { color: #58a6ff; font-size: 0.9rem; word-break: break-all; }
  .port { font-family: monospace; color: #484f58; }
</style>
</head>
<body>
<div class="card">
  <div class="apk-icon">📦</div>
  <h1>APK Ready</h1>
  <p>{{.APK}}</p>
  <a class="button" href="/{{.APK}}" download>Download</a>
  <div class="url">
    or open this URL on your device:<br>
    <code>{{.URL}}</code>
  </div>
</div>
</body>
</html>`

	tmpl, err := template.New("install").Parse(page)
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, map[string]any{"APK": apkName, "URL": url, "Port": port})
}

func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

func (uc *InstallUseCase) FindAPK(projectPath string) (string, error) {
	return findAPK(projectPath)
}

func (uc *InstallUseCase) FindAPKs(projectPath string) []string {
	var apks []string
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if !e.IsDir() && (strings.HasSuffix(e.Name(), ".apk")) {
			apks = append(apks, filepath.Join(projectPath, e.Name()))
		}
	}
	return apks
}
