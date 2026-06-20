package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/Shadw-Developer/tamk/internal/config"
	repoRemote "github.com/Shadw-Developer/tamk/internal/repository"
	repoFS "github.com/Shadw-Developer/tamk/internal/repository/filesystem"
	"github.com/Shadw-Developer/tamk/internal/usecase"
	"github.com/Shadw-Developer/tamk/pkg/errors"
	"github.com/Shadw-Developer/tamk/pkg/logger"
)

var (
	verbose  bool
	password string
	noWS     bool
	wsPort   int
)

func NewRootCmd() *cobra.Command {
	cfg := config.New()
	projRepo := repoFS.NewProjectRepository()
	tmplRepo := repoFS.NewTemplateRepository(cfg)
	buildRepo := repoFS.NewBuildRepository()
	updateRepo := repoRemote.NewUpdateRepository()

	createUC := usecase.NewCreateProjectUseCase(cfg, projRepo, tmplRepo)
	buildUC := usecase.NewBuildProjectUseCase(cfg, buildRepo, projRepo)
	setupUC := usecase.NewSetupEnvironmentUseCase(cfg)
	devUC := usecase.NewDevModeUseCase(cfg, buildUC, projRepo)
	updateUC := usecase.NewUpdateUseCase(cfg, updateRepo)

	cmd := &cobra.Command{
		Use:   "tamk",
		Short: "Termux APK Manager Kit — build Android apps from Termux",
		Long:  `T.A.M.K (Termux APK Manager Kit) v` + config.Version + ` — Professional automation framework for native Android app development directly in Termux.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger.Init(verbose)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "V", false, "Debug-level logging")
	cmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Keystore password")
	cmd.PersistentFlags().BoolVar(&noWS, "no-ws", false, "Disable WebSocket (HTTP fallback)")
	cmd.PersistentFlags().IntVar(&wsPort, "ws-port", 8765, "WebSocket port")

	cmd.AddCommand(newCreateCmd(createUC))
	cmd.AddCommand(newBuildCmd(buildUC, projRepo))
	cmd.AddCommand(newDevCmd(devUC, projRepo, cfg))
	cmd.AddCommand(newSetupCmd(setupUC))
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newInstallCmd(projRepo))
	cmd.AddCommand(newUpdateCmd(updateUC))
	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newShellCmd(cfg, createUC, buildUC, devUC, setupUC, updateUC, projRepo))

	return cmd
}

func newCreateCmd(uc *usecase.CreateProjectUseCase) *cobra.Command {
	return &cobra.Command{
		Use:   "create",
		Short: "Create a new project wizard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createProjectInteractive(context.Background(), uc)
		},
	}
}

func newBuildCmd(uc *usecase.BuildProjectUseCase, projRepo *repoFS.ProjectRepository) *cobra.Command {
	return &cobra.Command{
		Use:   "build",
		Short: "Build APK from current project",
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

			result, err := uc.FullBuild(ctx, usecase.BuildInput{
				ProjectPath: cwd,
				Password:    pwd,
			})
			if err != nil {
				return err
			}
			if !result.Success {
				return fmt.Errorf("build failed: %s: %w", result.ErrorMsg, errors.ErrBuildFailed)
			}
			logger.Success("APK ready", "path", result.APKPath)
			return nil
		},
	}
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

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "run [file]",
		Short: "Execute Kotlin/Java snippet",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Run command (not yet fully migrated)")
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
