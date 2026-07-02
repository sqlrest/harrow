// Command harrow formats PostgreSQL SQL. It reads SQL from the named files (or
// standard input), lays it out in a canonical style via gomatic/go-sql, and —
// because that library verifies every rendering — never changes a statement's
// meaning or drops a comment.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sort"

	app "github.com/gomatic/go-app"
	"github.com/gomatic/go-log"
	"github.com/urfave/cli/v3"

	format "github.com/sqlrest/harrow/internal/app/commands/format"
)

const envPrefix = `HARROW_`

var (
	loggerConfig  log.LoggerConfig
	appCreator    = createApp
	loggerCreator = productionLogger
)

// productionLogger builds the application logger from the parsed logging flags.
func productionLogger(_ *cli.Command) *slog.Logger {
	return loggerConfig.NewLogger(os.Stderr)
}

// version is the application version, set via ldflags: -X main.version=1.0.0.
var version = "dev"

// osExit is indirected so tests can observe the process exit code.
var osExit = os.Exit

func main() { osExit(run(os.Args)) }

// run builds and executes the CLI, returning the process exit code. Keeping the
// exit code as a return value (rather than calling os.Exit here) makes the whole
// run path testable.
func run(args []string) int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	if err := appCreator(loggerCreator).Run(ctx, args); err != nil {
		slog.Error("Formatting error", "error", err)
		return 1
	}
	return 0
}

// createApp assembles the CLI: harrow is a single-verb tool, so the format
// command is the root, and the composition root adds only the version, the
// shell completion, the logger hook, and the logging flags.
func createApp(getLogger app.GetLoggerFunc) *cli.Command {
	cliApp := format.Command()
	cliApp.Version = version
	cliApp.EnableShellCompletion = true
	cliApp.Before = app.LoggerBefore(getLogger)
	cliApp.Flags = append(
		cliApp.Flags,
		&cli.StringFlag{
			Name:        "log-level",
			Sources:     cli.EnvVars(envPrefix + "LOG_LEVEL"),
			Value:       "warn",
			Usage:       "Set the logging level (debug, info, warn, error)",
			Destination: (*string)(&loggerConfig.LogLevel),
		},
		&cli.StringFlag{
			Name:        "log-format",
			Sources:     cli.EnvVars(envPrefix + "LOG_FORMAT"),
			Value:       "text",
			Usage:       "Set the log output format (text, json)",
			Destination: (*string)(&loggerConfig.LogFormat),
		},
	)

	sort.Sort(cli.FlagsByName(cliApp.Flags))

	return cliApp
}
