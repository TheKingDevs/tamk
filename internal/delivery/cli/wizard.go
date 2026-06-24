package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/domain/entity"
	"github.com/TheKingDevs/tamk/internal/domain/valueobject"
	"github.com/TheKingDevs/tamk/internal/usecase"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

var stdinReader = bufio.NewReader(os.Stdin)

const (
	clr = "\033[2K\r"
	up  = "\033[A"
)

func termWidth() int {
	out, err := exec.Command("tput", "cols").Output()
	if err != nil {
		return 80
	}
	var w int
	fmt.Sscanf(string(out), "%d", &w)
	if w < 40 {
		return 40
	}
	return w
}

func center(text string, width int) string {
	clean := stripANSI(text)
	padding := (width - len(clean)) / 2
	if padding < 0 {
		padding = 0
	}
	return strings.Repeat(" ", padding) + text
}

func separator(char, color string, width int) string {
	if width <= 0 {
		width = termWidth()
	}
	return color + strings.Repeat(char, width) + logger.AnsiReset
}

func stripANSI(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*[mGKF]`)
	return re.ReplaceAllString(s, "")
}

func readLine() string {
	input, err := stdinReader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimRight(input, "\n\r")
}

func ask(question, defaultVal string) string {
	prompt := logger.AnsiCyan + logger.AnsiBold + question + logger.AnsiReset +
		" " + logger.AnsiGray + "(" + defaultVal + ")" + logger.AnsiReset +
		" " + logger.AnsiCyan + "❯" + logger.AnsiReset + " "
	os.Stdout.WriteString(clr + "  " + prompt)

	input := readLine()
	if input == "" {
		return defaultVal
	}
	return input
}

func askSecret(question string) string {
	prompt := logger.AnsiYellow + logger.AnsiBold + question + logger.AnsiReset +
		" " + logger.AnsiYellow + "❯" + logger.AnsiReset + " "
	os.Stdout.WriteString(clr + "  " + prompt)

	exec.Command("stty", "-echo").Run()
	input := readLine()
	exec.Command("stty", "echo").Run()
	fmt.Println()
	return strings.TrimSpace(input)
}

func showBanner() {
	os.Stdout.WriteString("\033[H\033[2J")
	w := termWidth()

	art := renderArt("TAMK", w)
	if art != "" {
		fmt.Println(art)
	} else {
		fmt.Println(center(logger.AnsiCyan+"  ╔══════════════════════════════╗"+logger.AnsiReset, w))
		fmt.Println(center(logger.AnsiCyan+"  ║     T.A.M.K v"+config.Version+"     ║"+logger.AnsiReset, w))
		fmt.Println(center(logger.AnsiCyan+"  ╚══════════════════════════════╝"+logger.AnsiReset, w))
	}
	fmt.Println()
	fmt.Println(center(logger.AnsiGray+logger.AnsiDim+"Termux APK Manager Kit"+logger.AnsiReset, w))
	fmt.Println(center(logger.AnsiGray+logger.AnsiDim+"Terminal Wizard"+logger.AnsiReset, w))
	fmt.Println()
	fmt.Println(center(separator("─", logger.AnsiGray, 0), w))
	fmt.Println()
}

func renderArt(text string, width int) string {
	out, err := exec.Command("toilet", "-f", "standard", "-F", "metal", text).Output()
	if err != nil || len(out) == 0 {
		return ""
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	var result []string
	for _, line := range lines {
		result = append(result, center(logger.AnsiCyan+line+logger.AnsiReset, width))
	}
	return strings.Join(result, "\n")
}

func showSuccessBanner(name, projType, projPath, pkg string) {
	fmt.Println()
	fmt.Println("  " + logger.AnsiGreen + separator("═", logger.AnsiGreen, 0) + logger.AnsiReset)
	fmt.Println(clr + "  " + logger.AnsiGreen + "✅  PROJETO '" + logger.AnsiBold + name + logger.AnsiReset + logger.AnsiGreen + "' CRIADO!" + logger.AnsiReset)
	fmt.Println("  " + logger.AnsiGreen + separator("═", logger.AnsiGreen, 0) + logger.AnsiReset)
	fmt.Println()
	fmt.Println("  " + logger.AnsiCyan + "▸" + logger.AnsiReset + " Tipo: " + logger.AnsiBold + projType + logger.AnsiReset)
	fmt.Println("  " + logger.AnsiCyan + "▸" + logger.AnsiReset + " Pacote: " + logger.AnsiBold + pkg + logger.AnsiReset)
	fmt.Println("  " + logger.AnsiCyan + "▸" + logger.AnsiReset + " Destino: " + logger.AnsiBold + projPath + logger.AnsiReset)
	fmt.Println()
	fmt.Println("  " + logger.AnsiGray + separator("─", logger.AnsiGray, 0) + logger.AnsiReset)
	fmt.Println("  " + logger.AnsiYellow + logger.AnsiBold + "Próximo passo:" + logger.AnsiReset + " " + logger.AnsiGray + "cd " + name + " && tamk build -p <senha>" + logger.AnsiReset)
	fmt.Println()
}

func createProjectInteractive(ctx context.Context, uc *usecase.CreateProjectUseCase) error {
	showBanner()

	var name, author, version, webURL string
	var projType entity.ProjectType
	var webMode entity.WebContentMode

	for {
		name = ask("Project name", "MyApp")
		if _, err := valueobject.NewProjectName(name); err == nil {
			break
		}
		fmt.Printf(up + clr + "  " + logger.AnsiRed + "✖" + logger.AnsiReset + " Invalid name. Use letters, numbers, hyphens.\n")
	}

	author = ask("Author", "Developer")

	for {
		version = ask("Version (SEMVER)", "1.0.0")
		if _, err := valueobject.ParseVersion(version); err == nil {
			break
		}
		fmt.Printf(up + clr + "  " + logger.AnsiRed + "✖" + logger.AnsiReset + " Invalid version. Use MAJOR.MINOR.PATCH (e.g. 1.0.0)\n")
	}

	fmt.Println()
	fmt.Println("  " + logger.AnsiBold + logger.AnsiYellow + "SELECT ENGINE" + logger.AnsiReset)
	fmt.Println()
	fmt.Printf("  %s[1]%s Standard Console\n", logger.AnsiCyan, logger.AnsiReset)
	fmt.Printf("    %s└─%s CLI scripts and automation\n\n", logger.AnsiGray, logger.AnsiReset)
	fmt.Printf("  %s[2]%s Native Android (UI/APK)\n", logger.AnsiGreen, logger.AnsiReset)
	fmt.Printf("    %s└─%s Native XML/Kotlin interface\n\n", logger.AnsiGray, logger.AnsiReset)
	fmt.Printf("  %s[3]%s Universal WebApp (HTML/JS)\n", logger.AnsiPurple, logger.AnsiReset)
	fmt.Printf("    %s└─%s Hybrid WebView\n\n", logger.AnsiGray, logger.AnsiReset)

	choice := ask("Engine (1-3)", "2")
	switch strings.TrimSpace(choice) {
	case "1":
		projType = entity.ProjectTypeConsole
	case "3":
		projType = entity.ProjectTypeWebApp
	default:
		projType = entity.ProjectTypeUIAPK
	}

	webURL = "file:///android_asset/index.html"
	if projType == entity.ProjectTypeWebApp {
		fmt.Println()
		fmt.Println("  " + logger.AnsiBold + logger.AnsiYellow + "WEB CONTENT TYPE" + logger.AnsiReset)
		fmt.Println()
		fmt.Printf("  %s[1]%s Internal (assets/ folder)\n", logger.AnsiCyan, logger.AnsiReset)
		fmt.Printf("  %s[2]%s External (Remote URL)\n\n", logger.AnsiPurple, logger.AnsiReset)

		mode := ask("Option", "1")
		if strings.TrimSpace(mode) == "2" {
			webMode = entity.WebContentExternal
			for {
				webURL = ask("App URL", "https://example.com")
				if strings.HasPrefix(webURL, "http://") || strings.HasPrefix(webURL, "https://") {
					break
				}
				fmt.Printf(up + clr + "  " + logger.AnsiRed + "✖" + logger.AnsiReset + " URL must start with http:// or https://\n")
			}
		} else {
			webMode = entity.WebContentInternal
		}
	}

	pwd := password
	if pwd == "" && (projType == entity.ProjectTypeWebApp || projType == entity.ProjectTypeUIAPK) {
		fmt.Println()
		fmt.Println("  " + logger.AnsiBold + logger.AnsiYellow + "⚠ SECURITY: Keystore password" + logger.AnsiReset)
		fmt.Println("    " + logger.AnsiGray + "(minimum 6 characters)" + logger.AnsiReset)
		for {
			pwd = askSecret("Keystore password")
			if len(pwd) >= 6 {
				break
			}
			fmt.Printf(up + clr + "  " + logger.AnsiRed + "✖" + logger.AnsiReset + " Password must be at least 6 characters\n")
		}
	}

	output, err := uc.Execute(ctx, usecase.CreateProjectInput{
		Name:     name,
		Type:     projType,
		Version:  version,
		Author:   author,
		Password: pwd,
		WebURL:   webURL,
		WebMode:  webMode,
	})
	if err != nil {
		return err
	}

	showSuccessBanner(name, string(projType), output.ProjectPath, output.Project.PackageName)

	return nil
}

func createProjectFromFlags(ctx context.Context, uc *usecase.CreateProjectUseCase, name, projTypeStr, version, author, webURL, webModeStr string) error {
	var projType entity.ProjectType
	switch projTypeStr {
	case "console":
		projType = entity.ProjectTypeConsole
	case "webapp":
		projType = entity.ProjectTypeWebApp
	default:
		projType = entity.ProjectTypeUIAPK
	}

	var webMode entity.WebContentMode
	if webModeStr == "external" {
		webMode = entity.WebContentExternal
	} else {
		webMode = entity.WebContentInternal
	}

	if projType == entity.ProjectTypeWebApp && webMode == entity.WebContentInternal {
		webURL = "file:///android_asset/index.html"
	}

	pwd := password
	if pwd == "" && (projType == entity.ProjectTypeWebApp || projType == entity.ProjectTypeUIAPK) {
		pwd = "test123"
	}

	output, err := uc.Execute(ctx, usecase.CreateProjectInput{
		Name:     name,
		Type:     projType,
		Version:  version,
		Author:   author,
		Password: pwd,
		WebURL:   webURL,
		WebMode:  webMode,
	})
	if err != nil {
		return err
	}

	showSuccessBanner(name, string(projType), output.ProjectPath, output.Project.PackageName)
	return nil
}
