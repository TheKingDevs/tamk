package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/TheKingDevs/tamk/internal/config"
	"github.com/TheKingDevs/tamk/internal/domain/entity"
	repoRemote "github.com/TheKingDevs/tamk/internal/repository"
	repoFS "github.com/TheKingDevs/tamk/internal/repository/filesystem"
	"github.com/TheKingDevs/tamk/internal/tools"
	"github.com/TheKingDevs/tamk/internal/usecase"
	"github.com/TheKingDevs/tamk/pkg/errors"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

var (
	verbose  bool
	password string
	guardian bool
	optimize bool
)

func NewRootCmd() *cobra.Command {
	cfg := config.New()
	projRepo := repoFS.NewProjectRepository()
	tmplRepo := repoFS.NewTemplateRepository(cfg)
	buildRepo := repoFS.NewBuildRepository()
	updateRepo := repoRemote.NewUpdateRepository()
	toolMgr := tools.New(tools.Config{
		DevDir:  cfg.DevDir,
		SDKPath: cfg.SDKPath,
	})

	createUC := usecase.NewCreateProjectUseCase(cfg, projRepo, tmplRepo)
	buildUC := usecase.NewBuildProjectUseCase(cfg, buildRepo, projRepo)
	setupUC := usecase.NewSetupEnvironmentUseCase(cfg)
	devUC := usecase.NewDevModeUseCase(cfg, buildUC, projRepo)
	updateUC := usecase.NewUpdateUseCase(cfg, updateRepo)
	runUC := usecase.NewRunUseCase(toolMgr)

	cmd := &cobra.Command{
		Use:   "tamk",
		Short: "Termux APK Manager Kit — build Android APKs from any platform",
		Long:  `T.A.M.K (Termux APK Manager Kit) v` + config.Version + ` — Professional automation framework for native Android app development.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger.Init(verbose)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "Debug-level logging")
	cmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Keystore password")
	cmd.PersistentFlags().BoolVar(&guardian, "guardian", false, "Enable Guardian security protection")
	cmd.PersistentFlags().BoolVar(&optimize, "optimize", false, "Use R8 for optimized DEX output (replaces D8)")

	cmd.AddCommand(newCreateCmd(createUC))
	cmd.AddCommand(newBuildCmd(buildUC, projRepo))
	cmd.AddCommand(newDevCmd(devUC, projRepo, cfg))
	cmd.AddCommand(newSetupCmd(setupUC))
	cmd.AddCommand(newRunCmd(toolMgr))
	cmd.AddCommand(newInstallCmd(projRepo))
	cmd.AddCommand(newUpdateCmd(updateUC))
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newShellCmd(cfg, createUC, buildUC, devUC, setupUC, updateUC, runUC, projRepo))
	cmd.AddCommand(newLibsCmd(cfg))

	return cmd
}

func newCreateCmd(uc *usecase.CreateProjectUseCase) *cobra.Command {
	var (
		name      string
		projType  string
		version   string
		author    string
		webURL    string
		webMode   string
		framework string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new project",
		Long:  `Create a new T.A.M.K project. Run without flags for interactive wizard.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name != "" && projType != "" && version != "" && author != "" {
				return createProjectFromFlags(context.Background(), uc, name, projType, version, author, webURL, webMode, framework)
			}
			return createProjectInteractive(context.Background(), uc)
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "Project name")
	cmd.Flags().StringVarP(&projType, "type", "t", "", "Project type: ui_apk, webapp, console")
	cmd.Flags().StringVarP(&version, "version", "v", "1.0.0", "Version (SEMVER)")
	cmd.Flags().StringVarP(&author, "author", "a", "Developer", "Author name")
	cmd.Flags().StringVar(&webURL, "url", "file:///android_asset/index.html", "WebApp URL (for webapp type)")
	cmd.Flags().StringVar(&webMode, "web-mode", "internal", "Web content mode: internal, external")
	cmd.Flags().StringVar(&framework, "framework", "xml", "UI framework: xml, compose")

	return cmd
}

func newBuildCmd(uc *usecase.BuildProjectUseCase, projRepo *repoFS.ProjectRepository) *cobra.Command {
	var buildTarget string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build APK or AAB from current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			ctx := cmd.Context()
			if !projRepo.Exists(ctx, cwd) {
				return fmt.Errorf("no project found in current directory: %w", errors.ErrProjectNotFound)
			}

			pwd := password
			if pwd == "" {
				fmt.Print("Keystore password: ")
				fmt.Scan(&pwd)
			}

			targets, err := parseBuildTargets(buildTarget)
			if err != nil {
				return fmt.Errorf("invalid build target %q: %w", buildTarget, err)
			}

			result, err := uc.FullBuild(ctx, usecase.BuildInput{
				ProjectPath: cwd,
				Password:    pwd,
				Targets:     targets,
				Guardian:    guardian,
				Optimize:    optimize,
			})
			if err != nil {
				return err
			}
			if !result.Success {
				return fmt.Errorf("build failed: %s: %w", result.ErrorMsg, errors.ErrBuildFailed)
			}
			if result.APKPath != "" {
				logger.Success(fmt.Sprintf("APK ready: %s", result.APKPath))
			}
			if result.AABPath != "" {
				logger.Success(fmt.Sprintf("AAB ready: %s", result.AABPath))
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&buildTarget, "target", "apk", "Build target: apk, aab, or apk,aab")
	cmd.Flags().BoolVar(&guardian, "guardian", false, "Enable Guardian security protection")
	return cmd
}

func parseBuildTargets(s string) ([]entity.BuildTarget, error) {
	parts := strings.Split(s, ",")
	targets := make([]entity.BuildTarget, 0, len(parts))
	seen := make(map[string]bool)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" || seen[p] {
			continue
		}
		switch p {
		case string(entity.BuildTargetAPK):
			targets = append(targets, entity.BuildTargetAPK)
		case string(entity.BuildTargetAAB):
			targets = append(targets, entity.BuildTargetAAB)
		default:
			return nil, fmt.Errorf("unknown target %q (valid: apk, aab)", p)
		}
		seen[p] = true
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no valid targets specified (use: apk, aab, or apk,aab)")
	}
	return targets, nil
}

func newDevCmd(uc *usecase.DevModeUseCase, projRepo *repoFS.ProjectRepository, cfg *config.Config) *cobra.Command {
	return &cobra.Command{
		Use:   "dev",
		Short: "Start development mode with HMR",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			if !projRepo.Exists(ctx, cwd) {
				return fmt.Errorf("no project found in current directory: %w", errors.ErrProjectNotFound)
			}

			if err := uc.Start(ctx, cwd, password); err != nil {
				return err
			}
			defer uc.Stop(ctx)

			<-ctx.Done()
			return nil
		},
	}
}

func newSetupCmd(uc *usecase.SetupEnvironmentUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "Download SDK and generate debug keystore",
		RunE: func(cmd *cobra.Command, args []string) error {
			return uc.Execute(context.Background())
		},
	}
}

func newRunCmd(tools tools.ToolManager) *cobra.Command {
	runUC := usecase.NewRunUseCase(tools)
	return &cobra.Command{
		Use:   "run [file]",
		Short: "Execute Kotlin/Java snippet",
		Long:  `Compile and run a Kotlin or Java file. If no file is specified, looks for Main.kt in src/`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			fileName := ""
			if len(args) > 0 {
				fileName = args[0]
			}

			result, err := runUC.Execute(context.Background(), usecase.RunInput{
				ProjectPath: cwd,
				FileName:    fileName,
			})
			if err != nil {
				return err
			}
			if result.ExitCode != 0 {
				return fmt.Errorf("process exited with code %d", result.ExitCode)
			}
			return nil
		},
	}
}

func newInstallCmd(projRepo *repoFS.ProjectRepository) *cobra.Command {
	installUC := usecase.NewInstallUseCase()
	return &cobra.Command{
		Use:   "install [port]",
		Short: "Serve APK download via local HTTP server",
		Long:  `Start a local HTTP server to download the APK on your Android device.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, _ := os.Getwd()
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()

			if !projRepo.Exists(ctx, cwd) {
				return fmt.Errorf("no project found in current directory: %w", errors.ErrProjectNotFound)
			}

			port := 9090
			if len(args) > 0 {
				fmt.Sscanf(args[0], "%d", &port)
			}

			_, err := installUC.Serve(ctx, cwd, port)
			if err != nil {
				return err
			}

			<-ctx.Done()
			return nil
		},
	}
}

func newUpdateCmd(uc *usecase.UpdateUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Check and install updates",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := uc.Check(context.Background())
			if err != nil {
				return err
			}
			if info == nil {
				logger.Info("No updates available")
				return nil
			}
			uc.PrintUpdateInfo(info)

			if uc.ShouldAutoInstall(info) {
				logger.Info("Auto-installing update...")
				return uc.Install(context.Background(), info)
			}

			if uc.ShouldPrompt(info) {
				fmt.Print("\nInstall update? [y/N] ")
				var answer string
				fmt.Scan(&answer)
				if strings.ToLower(answer) == "y" || strings.ToLower(answer) == "yes" {
					return uc.Install(context.Background(), info)
				}
				logger.Info("Update skipped")
			}

			return nil
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := config.New()
			fmt.Printf("T.A.M.K v%s\n", config.Version)
			fmt.Printf("Environment: %s\n", cfg.Env)
			fmt.Printf("TAMK Home: %s\n", cfg.TAMKHome)
		},
	}
}

func newLibsCmd(cfg *config.Config) *cobra.Command {
	libMgr := usecase.NewLibraryManager(cfg)

	cmd := &cobra.Command{
		Use:   "libs",
		Short: "Manage Android libraries",
		Long:  `Download, install, and manage Android libraries (AndroidX, Jetpack, etc.)`,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List available libraries",
		RunE: func(cmd *cobra.Command, args []string) error {
			return libMgr.ListLibraries()
		},
	}

	installCmd := &cobra.Command{
		Use:   "install [artifact]",
		Short: "Install a library",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return libMgr.InstallLibrary(context.Background(), args[0])
		},
	}

	removeCmd := &cobra.Command{
		Use:   "remove [artifact]",
		Short: "Remove an installed library",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return libMgr.RemoveLibrary(args[0])
		},
	}

	cpCmd := &cobra.Command{
		Use:   "classpath",
		Short: "Show library classpath",
		RunE: func(cmd *cobra.Command, args []string) error {
			cp, err := libMgr.GetClasspath()
			if err != nil {
				return err
			}
			if cp == "" {
				fmt.Println("No libraries installed")
				return nil
			}
			fmt.Println(cp)
			return nil
		},
	}

	installedCmd := &cobra.Command{
		Use:   "installed",
		Short: "List installed libraries",
		RunE: func(cmd *cobra.Command, args []string) error {
			libs := libMgr.GetInstalledLibraries()
			if len(libs) == 0 {
				fmt.Println("No libraries installed")
				return nil
			}
			fmt.Println("Installed libraries:")
			for _, l := range libs {
				fmt.Printf("  - %s\n", l)
			}
			return nil
		},
	}

	cmd.AddCommand(listCmd)
	cmd.AddCommand(installCmd)
	cmd.AddCommand(removeCmd)
	cmd.AddCommand(cpCmd)
	cmd.AddCommand(installedCmd)

	return cmd
}
