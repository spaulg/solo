package subcommand

import (
	"fmt"
	"github.com/spaulg/solo/internal/pkg/common/logging"
	"github.com/spaulg/solo/internal/pkg/entrypoint/context"
	"github.com/spf13/cobra"
	"log/slog"
	"os"
	"path"
	"time"
)

func NewRootCommand(_ *context.EntrypointContext) *cobra.Command {
	return &cobra.Command{
		Use:          "solo",
		SilenceUsage: true,
	}
}

func Execute() {
	cobra.EnableCommandSorting = false

	entrypointCtx := loadContext()

	rootCmd := NewRootCommand(entrypointCtx)
	rootCmd.AddCommand(NewEntrypointCommand(entrypointCtx))
	rootCmd.AddCommand(NewTriggerEventCommand(entrypointCtx))

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func loadContext() *context.EntrypointContext {
	logFileName := path.Join("/solo/service/logs", time.Now().Format("2006-01-02.log"))

	builder := logging.NewLogHandlerBuilder()
	handler, err := builder.
		WithLogFilePath(logFileName).
		WithLogLevel("info").
		WithLogHandlerName("text").
		Build()

	if err != nil {
		panic(fmt.Sprintf("%v\n", err))
	}

	return &context.EntrypointContext{
		Logger: slog.New(handler),
	}
}
