package subcommand

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/spf13/cobra"

	"github.com/spaulg/solo/internal/pkg/impl/host/app/context"
)

const RequireConfigLoadSuccessAnnotation = "RequireConfigLoadSuccess"
const RequireProjectLoadSuccessAnnotation = "RequireProjectLoadSuccess"

var ErrUserAbortedCommand = errors.New("user aborted command")

func Execute() int {
	cobra.EnableCommandSorting = false

	soloCtx, err := context.LoadCliContext()
	if err != nil {
		return 1
	}

	rootCmd, stopProfiling := NewRootCommand(soloCtx)
	defer func() {
		if err := stopProfiling(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to stop profiling operations: %v\n", err)
		}
	}()

	rootCmd.AddGroup(&cobra.Group{ID: "lifecycle", Title: "Life Cycle Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "tooling", Title: "Tooling Commands:"})
	rootCmd.AddGroup(&cobra.Group{ID: "config", Title: "Config Commands:"})

	// Lifecycle
	rootCmd.AddCommand(NewStartCommand(soloCtx))
	rootCmd.AddCommand(NewStopCommand(soloCtx))
	rootCmd.AddCommand(NewRestartCommand(soloCtx))
	rootCmd.AddCommand(NewRebuildCommand(soloCtx))
	rootCmd.AddCommand(NewDestroySubCommand(soloCtx))
	rootCmd.AddCommand(NewCleanSubCommand(soloCtx))

	// Tooling
	rootCmd.AddCommand(NewShCommand(soloCtx))
	rootCmd.AddCommand(NewLogsCommand(soloCtx))
	rootCmd.AddCommand(NewToolCommands(soloCtx)...)

	// Config
	rootCmd.AddCommand(NewDumpConfigCommand(soloCtx))
	rootCmd.AddCommand(NewDumpComposeConfigCommand(soloCtx))

	// Other
	rootCmd.AddCommand(NewVersionCommand(soloCtx))

	err = rootCmd.Execute()

	if err != nil && !errors.Is(err, ErrUserAbortedCommand) {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	return 0
}

func NewRootCommand(soloCtx *context.CliContext) (*cobra.Command, func() error) {
	var cpuProfile string
	var memProfile string
	var cpuFile *os.File

	stopProfilingCallback := func() error {
		pprof.StopCPUProfile()

		if cpuFile != nil {
			_ = cpuFile.Close()
		}

		if memProfile != "" {
			runtime.GC()

			f, err := os.OpenFile(memProfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				return fmt.Errorf("failed to create memory profile %q: %v", memProfile, err)
			}

			defer func() {
				if closeErr := f.Close(); closeErr != nil {
					_, _ = fmt.Fprintf(os.Stderr, "failed to close memory profile %q: %v\n", memProfile, closeErr)
				}
			}()

			if err := pprof.WriteHeapProfile(f); err != nil {
				return fmt.Errorf("failed to write memory profile %q: %v", memProfile, err)
			}
		}

		return nil
	}

	cmd := &cobra.Command{
		Use:           "solo",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if cpuProfile != "" && memProfile != "" && cpuProfile == memProfile {
				return fmt.Errorf("CPU and memory profiles cannot be the same filename")
			}

			if cpuProfile != "" {
				var err error

				cpuFile, err = os.OpenFile(cpuProfile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
				if err != nil {
					return fmt.Errorf("failed to create CPU profile %q: %v", cpuProfile, err)
				}

				if err := pprof.StartCPUProfile(cpuFile); err != nil {
					defer func(cpuFile *os.File) {
						if closeErr := cpuFile.Close(); closeErr != nil {
							_, _ = fmt.Fprintf(os.Stderr, "failed to close CPU profile %q: %v\n", cpuProfile, closeErr)
						}
					}(cpuFile)

					cpuFile = nil
					return fmt.Errorf("CPU profiling failed to start: %v", err)
				}
			}

			if cmd.Annotations == nil || cmd.Annotations[RequireConfigLoadSuccessAnnotation] != "true" {
				return nil
			}

			if soloCtx.ConfigLoadErr != nil {
				return soloCtx.ConfigLoadErr
			}

			if cmd.Annotations == nil || cmd.Annotations[RequireProjectLoadSuccessAnnotation] != "true" {
				return nil
			}

			if soloCtx.ProjectLoadErr != nil {
				return soloCtx.ProjectLoadErr
			}

			soloCtx.CommandPath = strings.Join(strings.Split(cmd.CommandPath(), " ")[1:], " ")
			soloCtx.CommandArgs = os.Args[1:]

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&cpuProfile, "cpu-profile", "", "Enable CPU profiling and write data to the supplied file")
	cmd.PersistentFlags().StringVar(&memProfile, "mem-profile", "", "Enable memory profiling and write data to the supplied file")

	return cmd, stopProfilingCallback
}
