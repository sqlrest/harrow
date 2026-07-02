// Package format wires harrow's formatting verb: it declares the write/list
// flags, binds them to the domain Config, and routes the action to the domain
// Run. harrow is a single-verb tool, so this command is shaped to serve as the
// application root; the composition root (cmd/harrow) adds only the version,
// the logging flags, and the logger hook.
package format

import (
	"context"
	"io"
	"os"

	app "github.com/gomatic/go-app"
	"github.com/urfave/cli/v3"

	domain "github.com/sqlrest/harrow/internal/domain/format"
)

const (
	name        = `harrow`
	usage       = `Format PostgreSQL SQL.`
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
	envPrefix = `HARROW_`
)

const (
	listFlag  = "list"
	writeFlag = "write"
)

var (
	cfg       domain.Config
	runFormat = domain.Run

	// stdin is the stream formatted when no files are named, injected so a test
	// can feed the command without touching the process's real standard input.
	stdin io.Reader = os.Stdin
)

// Command returns the formatting command, shaped to serve as harrow's root.
func Command() *cli.Command {
	return &cli.Command{
		Name:        name,
		Usage:       usage,
		ArgsUsage:   argUsage,
		Description: description,
		Action:      action,
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
		},
	}
}

// action runs the formatting command over the positional file arguments,
// reading standard input when there are none.
func action(ctx context.Context, c *cli.Command) error {
	_, err := runFormat(ctx, app.GetLogger(c), cfg, stdin, c.Root().Writer, paths(c.Args().Slice())...)
	return err
}

// paths converts the raw positional arguments into domain file paths.
func paths(args []string) []domain.FilePath {
	converted := make([]domain.FilePath, len(args))
	for i, arg := range args {
		converted[i] = domain.FilePath(arg)
	}
	return converted
}
