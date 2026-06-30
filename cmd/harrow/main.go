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

	"github.com/sqlrest/harrow/internal/domain/format"
)

const (
	argUsage    = `[file...]`
	description = `harrow reads PostgreSQL SQL, lays it out in a canonical style, and writes it
back out. With no files it reads standard input and writes to standard output,
so it composes in a pipe:

  cat schema.sql | harrow

Given files, it prints each formatted file to standard output, or:

  harrow --write *.sql      # rewrite changed files in place
  harrow --list *.sql       # print the paths that would change

harrow never changes what a statement means and never drops a comment: every
rendering is verified against the original, and anything it can't prove faithful
is left exactly as written.`
	envName   = "HARROW"
	envPrefix = envName + "_"
	name      = `harrow`
	usage     = `Format PostgreSQL SQL.`
)

const (
	listFlag  = "list"
	writeFlag = "write"
)

var (
	cfg           format.Config
	loggerConfig  log.LoggerConfig
	appCreator    = createApp
	loggerCreator = productionLogger
	runFormat     = format.Run
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

// createApp constructs the definition of the CLI.
func createApp(getLogger app.GetLoggerFunc) *cli.Command {
	cliApp := &cli.Command{
		Name:                  name,
		Usage:                 usage,
		ArgsUsage:             argUsage,
		Description:           description,
		Version:               version,
		EnableShellCompletion: true,
		Action:                formatAction,
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			c.Root().Metadata[app.LoggerMetadataKey] = getLogger(c)
			return ctx, nil
		},
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        writeFlag,
				Aliases:     []string{"w"},
				Sources:     cli.EnvVars(envPrefix + "WRITE"),
				Usage:       "Rewrite each changed file in place instead of writing to stdout",
				Destination: (*bool)(&cfg.WriteEnabled),
			},
			&cli.BoolFlag{
				Name:        listFlag,
				Aliases:     []string{"l"},
				Sources:     cli.EnvVars(envPrefix + "LIST"),
				Usage:       "Print the paths of files whose formatting would change",
				Destination: (*bool)(&cfg.ListEnabled),
			},
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
		},
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))

	return cliApp
}

// formatAction runs the formatting command over the positional file arguments,
// reading standard input when there are none.
func formatAction(ctx context.Context, c *cli.Command) error {
	_, err := runFormat(ctx, app.GetLogger(c), cfg, os.Stdin, c.Root().Writer, c.Args().Slice()...)
	return err
}
