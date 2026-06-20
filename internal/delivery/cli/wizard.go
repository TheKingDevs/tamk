package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/Shadw-Developer/tamk/internal/config"
	"github.com/Shadw-Developer/tamk/internal/domain/entity"
	"github.com/Shadw-Developer/tamk/internal/domain/valueobject"
	"github.com/Shadw-Developer/tamk/internal/usecase"
)

var stdinReader = bufio.NewReader(os.Stdin)

const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiCyan   = "\033[36m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiRed    = "\033[31m"
	ansiGray   = "\033[90m"
	ansiPurple = "\033[35m"
	clr        = "\033[2K\r"
	up         = "\033[A"
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
	return color + strings.Repeat(char, width) + ansiReset
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
	prompt := ansiCyan + ansiBold + question + ansiReset +
		" " + ansiGray + "(" + defaultVal + ")" + ansiReset +
		" " + ansiCyan + "❯" + ansiReset + " "
	os.Stdout.WriteString(clr + "  " + prompt)

	input := readLine()
	if input == "" {
		return defaultVal
	}
	return input
}

func askSecret(question string) string {
	prompt := ansiYellow + ansiBold + question + ansiReset +
		" " + ansiYellow + "❯" + ansiReset + " "
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
		fmt.Println(center(ansiCyan+"  ╔══════════════════════════════╗"+ansiReset, w))
		fmt.Println(center(ansiCyan+"  ║     T.A.M.K v"+config.Version+"     ║"+ansiReset, w))
		fmt.Println(center(ansiCyan+"  ╚══════════════════════════════╝"+ansiReset, w))
	}
	fmt.Println()
	fmt.Println(center(ansiGray+ansiDim+"Termux APK Manager Kit"+ansiReset, w))
	fmt.Println(center(ansiGray+ansiDim+"Terminal Wizard"+ansiReset, w))
	fmt.Println()
	fmt.Println(center(separator("─", ansiGray, 0), w))
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
		result = append(result, center(ansiCyan+line+ansiReset, width))
	}
	return strings.Join(result, "\n")
}

func showSuccessBanner(name, projType, projPath, pkg string) {
	fmt.Println()
	fmt.Println("  " + ansiGreen + separator("═", ansiGreen, 0) + ansiReset)
	fmt.Println(clr + "  " + ansiGreen + "✅  PROJETO '" + ansiBold + name + ansiReset + ansiGreen + "' CRIADO!" + ansiReset)
	fmt.Println("  " + ansiGreen + separator("═", ansiGreen, 0) + ansiReset)
	fmt.Println()
	fmt.Println("  " + ansiCyan + "▸" + ansiReset + " Tipo: " + ansiBold + projType + ansiReset)
	fmt.Println("  " + ansiCyan + "▸" + ansiReset + " Pacote: " + ansiBold + pkg + ansiReset)
	fmt.Println("  " + ansiCyan + "▸" + ansiReset + " Destino: " + ansiBold + projPath + ansiReset)
	fmt.Println()
	fmt.Println("  " + ansiGray + separator("─", ansiGray, 0) + ansiReset)
	fmt.Println("  " + ansiYellow + ansiBold + "Próximo passo:" + ansiReset + " " + ansiGray + "cd " + name + " && tamk build -p <senha>" + ansiReset)
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
		fmt.Printf(up + clr + "  " + ansiRed + "✖" + ansiReset + " Invalid name. Use letters, numbers, hyphens.\n")
	}

	author = ask("Author", "Developer")

	for {
		version = ask("Version (SEMVER)", "1.0.0")
		if _, err := valueobject.ParseVersion(version); err == nil {
			break
		}
		fmt.Printf(up + clr + "  " + ansiRed + "✖" + ansiReset + " Invalid version. Use MAJOR.MINOR.PATCH (e.g. 1.0.0)\n")
	}

	fmt.Println()
	fmt.Println("  " + ansiBold + ansiYellow + "SELECT ENGINE" + ansiReset)
	fmt.Println()
	fmt.Printf("  %s[1]%s Standard Console\n", ansiCyan, ansiReset)
	fmt.Printf("    %s└─%s CLI scripts and automation\n\n", ansiGray, ansiReset)
	fmt.Printf("  %s[2]%s Native Android (UI/APK)\n", ansiGreen, ansiReset)
	fmt.Printf("    %s└─%s Native XML/Kotlin interface\n\n", ansiGray, ansiReset)
	fmt.Printf("  %s[3]%s Universal WebApp (HTML/JS)\n", ansiPurple, ansiReset)
	fmt.Printf("    %s└─%s Hybrid WebView\n\n", ansiGray, ansiReset)

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
		fmt.Println("  " + ansiBold + ansiYellow + "WEB CONTENT TYPE" + ansiReset)
		fmt.Println()
		fmt.Printf("  %s[1]%s Internal (assets/ folder)\n", ansiCyan, ansiReset)
		fmt.Printf("  %s[2]%s External (Remote URL)\n\n", ansiPurple, ansiReset)

		mode := ask("Option", "1")
		if strings.TrimSpace(mode) == "2" {
			webMode = entity.WebContentExternal
			for {
				webURL = ask("App URL", "https://example.com")
				if strings.HasPrefix(webURL, "http://") || strings.HasPrefix(webURL, "https://") {
					break
				}
				fmt.Printf(up + clr + "  " + ansiRed + "✖" + ansiReset + " URL must start with http:// or https://\n")
			}
		} else {
			webMode = entity.WebContentInternal
		}
	}

	pwd := password
	if pwd == "" && (projType == entity.ProjectTypeWebApp || projType == entity.ProjectTypeUIAPK) {
		fmt.Println()
		fmt.Println("  " + ansiBold + ansiYellow + "⚠ SECURITY: Keystore password" + ansiReset)
		fmt.Println("    " + ansiGray + "(minimum 6 characters)" + ansiReset)
		for {
			pwd = askSecret("Keystore password")
			if len(pwd) >= 6 {
				break
			}
			fmt.Printf(up + clr + "  " + ansiRed + "✖" + ansiReset + " Password must be at least 6 characters\n")
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
