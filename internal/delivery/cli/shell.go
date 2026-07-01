package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/TheKingDevs/tamk/internal/config"
	repoFS "github.com/TheKingDevs/tamk/internal/repository/filesystem"
	"github.com/TheKingDevs/tamk/internal/usecase"
	"github.com/TheKingDevs/tamk/pkg/errors"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

type shellState struct {
	history  []string
	cwd      string
	cfg      *config.Config
	createUC *usecase.CreateProjectUseCase
	buildUC  *usecase.BuildProjectUseCase
	devUC    *usecase.DevModeUseCase
	setupUC  *usecase.SetupEnvironmentUseCase
	updateUC *usecase.UpdateUseCase
	runUC    *usecase.RunUseCase
	projRepo *repoFS.ProjectRepository
}

func newShellCmd(
	cfg *config.Config,
	createUC *usecase.CreateProjectUseCase,
	buildUC *usecase.BuildProjectUseCase,
	devUC *usecase.DevModeUseCase,
	setupUC *usecase.SetupEnvironmentUseCase,
	updateUC *usecase.UpdateUseCase,
	runUC *usecase.RunUseCase,
	projRepo *repoFS.ProjectRepository,
) *cobra.Command {
	return &cobra.Command{
		Use:   "shell",
		Short: "Start interactive development shell",
		Long:  `Open an interactive REPL session to create, build, and manage projects without leaving the shell.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			s := &shellState{
				cwd:      getCwd(),
				cfg:      cfg,
				createUC: createUC,
				buildUC:  buildUC,
				devUC:    devUC,
				setupUC:  setupUC,
				updateUC: updateUC,
				runUC:    runUC,
				projRepo: projRepo,
			}
			return s.run()
		},
	}
}

func getCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func (s *shellState) run() error {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("\033[36m\033[1m╔══════════════════════════════════════╗\033[0m")
	fmt.Println("\033[36m\033[1m║     T.A.M.K Interactive Shell       ║\033[0m")
	fmt.Println("\033[36m\033[1m║     Type 'help' for commands        ║\033[0m")
	fmt.Println("\033[36m\033[1m╚══════════════════════════════════════╝\033[0m")
	fmt.Println()

	for {
		fmt.Printf("\033[36mtamk\033[0m> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if len(s.history) == 0 || s.history[len(s.history)-1] != input {
			s.history = append(s.history, input)
		}

		if err := s.exec(input); err != nil {
			fmt.Fprintf(os.Stderr, "\033[31mError: %v\033[0m\n", err)
		}
		fmt.Println()
	}

	fmt.Println("\033[32mGoodbye!\033[0m")
	return scanner.Err()
}

func (s *shellState) exec(input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case "help":
		return s.showHelp()
	case "exit", "quit":
		fmt.Println("\033[32mGoodbye!\033[0m")
		os.Exit(0)
		return nil
	case "history":
		return s.showHistory()
	case "pwd":
		fmt.Println(s.cwd)
		return nil
	case "cd":
		return s.changeDir(args)
	case "version":
		return s.showVersion()
	case "create":
		return s.doCreate()
	case "build":
		return s.doBuild(args)
	case "dev":
		return s.doDev()
	case "setup":
		return s.doSetup()
	case "update":
		return s.doUpdate()
	case "run":
		return s.doRun(args)
	case "install":
		return s.doInstall()
	default:
		return fmt.Errorf("unknown command: %s (type 'help' for available commands)", cmd)
	}
}

func (s *shellState) showHelp() error {
	fmt.Println()
	fmt.Println("\033[1mAvailable commands:\033[0m")
	fmt.Println()
	type entry struct{ cmd, desc string }
	entries := []entry{
		{"help", "Show this help message"},
		{"create", "Create a new project via interactive wizard"},
		{"build [-p <password>]", "Build APK from current directory project"},
		{"dev", "Start development mode with HMR"},
		{"run <code>", "Execute Kotlin/Java snippet"},
		{"install", "Serve APK download via local HTTP server"},
		{"setup", "Download SDK and generate debug keystore"},
		{"update", "Check for T.A.M.K updates"},
		{"version", "Show version and environment info"},
		{"cd <dir>", "Change working directory"},
		{"pwd", "Show current working directory"},
		{"history", "Show command history"},
		{"exit / quit", "Exit the interactive shell"},
	}
	for _, e := range entries {
		fmt.Printf("  \033[36m%-28s\033[0m %s\n", e.cmd, e.desc)
	}
	return nil
}

func (s *shellState) showHistory() error {
	if len(s.history) == 0 {
		fmt.Println("\033[90m(no commands yet)\033[0m")
		return nil
	}
	for i, h := range s.history {
		fmt.Printf("  \033[90m%3d\033[0m  %s\n", i+1, h)
	}
	return nil
}

func (s *shellState) changeDir(args []string) error {
	target := "."
	if len(args) > 0 {
		target = args[0]
	}

	newDir := target
	if !filepath.IsAbs(target) {
		newDir = filepath.Join(s.cwd, target)
	}

	info, err := os.Stat(newDir)
	if err != nil {
		return fmt.Errorf("cannot access %s: %w", target, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", target)
	}

	if err := os.Chdir(newDir); err != nil {
		return fmt.Errorf("failed to change directory: %w", err)
	}
	s.cwd = newDir
	fmt.Printf("\033[90mchanged to %s\033[0m\n", s.cwd)
	return nil
}

func (s *shellState) showVersion() error {
	fmt.Printf("T.A.M.K \033[1m%s\033[0m\n", config.Version)
	fmt.Printf("Environment: %s\n", s.cfg.Env)
	fmt.Printf("TAMK Home: %s\n", s.cfg.TAMKHome)
	fmt.Printf("Working dir: %s\n", s.cwd)
	return nil
}

func (s *shellState) doCreate() error {
	logger.Init(false)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return createProjectInteractive(ctx, s.createUC)
}

func (s *shellState) doBuild(args []string) error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if !s.projRepo.Exists(ctx, s.cwd) {
		return fmt.Errorf("no project found in %s: %w", s.cwd, errors.ErrProjectNotFound)
	}

	pwd := password
	for i := 0; i < len(args); i++ {
		if args[i] == "-p" && i+1 < len(args) {
			pwd = args[i+1]
			break
		}
	}

	if pwd == "" {
		fmt.Print("Keystore password: ")
		pwd = readLine()
	}

	result, err := s.buildUC.FullBuild(ctx, usecase.BuildInput{
		ProjectPath: s.cwd,
		Password:    pwd,
	})
	if err != nil {
		return err
	}
	if !result.Success {
		return fmt.Errorf("build failed: %s: %w", result.ErrorMsg, errors.ErrBuildFailed)
	}
	logger.Success(fmt.Sprintf("APK ready: %s", result.APKPath))
	return nil
}

func (s *shellState) doDev() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if !s.projRepo.Exists(ctx, s.cwd) {
		return fmt.Errorf("no project found in %s: %w", s.cwd, errors.ErrProjectNotFound)
	}

	fmt.Println("\033[33mStarting dev mode (press Ctrl+C to stop and return to shell)...\033[0m")

	if err := s.devUC.Start(ctx, s.cwd, password); err != nil {
		return err
	}
	defer s.devUC.Stop(ctx)

	<-ctx.Done()
	signal.Reset(syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\033[32mDev mode stopped. Back to shell.\033[0m")
	return nil
}

func (s *shellState) doSetup() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	return s.setupUC.Execute(ctx)
}

func (s *shellState) doUpdate() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	info, err := s.updateUC.Check(ctx)
	if err != nil {
		return fmt.Errorf("update check failed: %w", err)
	}
	if info == nil {
		fmt.Println("\033[32mYou are up to date!\033[0m")
		return nil
	}
	s.updateUC.PrintUpdateInfo(info)
	return nil
}

func (s *shellState) doRun(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: run <filename>")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	result, err := s.runUC.Execute(ctx, usecase.RunInput{
		ProjectPath: s.cwd,
		FileName:    args[0],
	})
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("process exited with code %d", result.ExitCode)
	}
	return nil
}

func (s *shellState) doInstall() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	installUC := usecase.NewInstallUseCase()
	if _, err := installUC.Serve(ctx, s.cwd, 9090); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
