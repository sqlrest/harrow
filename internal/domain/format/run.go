package format

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	sql "github.com/gomatic/go-sql"
	"github.com/gomatic/go-sql/formatter"

	"github.com/sqlrest/harrow/internal/constants"
)

// fileMode is the permission new in-place writes keep; it matches the common
// default for SQL source files.
const fileMode os.FileMode = 0o644

// writeFile is [os.WriteFile], injected so a test can drive the in-place write
// failure path without depending on filesystem permissions.
var writeFile = os.WriteFile

// Result reports which inputs were not already canonically formatted.
type Result struct {
	Changed []filePath `json:"changed"`
}

// Run formats the SQL named in args, or standard input when no paths are given.
// With no paths it reads in and writes the formatted SQL to out. With paths it
// formats each file: WriteEnabled rewrites changed files in place, ListEnabled
// prints the paths that would change, and otherwise the formatted SQL goes to
// out. A file that won't parse stops the run with [constants.ErrFormat].
func Run(
	_ context.Context,
	logger *slog.Logger,
	cfg Config,
	in io.Reader,
	out io.Writer,
	args ...string,
) (Result, error) {
	paths := pathsFrom(args)
	if len(paths) == 0 {
		return Result{}, formatStream(in, out)
	}

	var result Result
	for _, path := range paths {
		if err := formatFile(cfg, out, path, &result); err != nil {
			return Result{}, err
		}
	}
	logger.Info("Formatting complete.", "files", len(paths), "changed", len(result.Changed))
	return result, nil
}

// pathsFrom turns positional arguments into file paths.
func pathsFrom(args []string) []filePath {
	paths := make([]filePath, len(args))
	for i, arg := range args {
		paths[i] = filePath(arg)
	}
	return paths
}

// formatStream formats everything read from in and writes it to out.
func formatStream(in io.Reader, out io.Writer) error {
	content, err := io.ReadAll(in)
	if err != nil {
		return constants.ErrReadInput.With(err)
	}
	formatted, err := format(content)
	if err != nil {
		return err
	}
	return writeOut(out, formatted)
}

// formatFile formats one file and routes the result according to cfg, recording
// the path in result when its formatting differs from what's on disk.
func formatFile(cfg Config, out io.Writer, path filePath, result *Result) error {
	content, err := os.ReadFile(string(path))
	if err != nil {
		return constants.ErrOpenFile.With(err, string(path))
	}
	formatted, err := format(content)
	if err != nil {
		return err
	}
	if formatted != string(content) {
		result.Changed = append(result.Changed, path)
	}
	return emit(cfg, out, path, string(content), formatted)
}

// emit routes a formatted file: rewritten in place, listed if changed, or
// printed to out.
func emit(cfg Config, out io.Writer, path filePath, original, formatted string) error {
	switch {
	case bool(cfg.WriteEnabled):
		return writeBack(path, original, formatted)
	case bool(cfg.ListEnabled):
		return listChanged(out, path, original, formatted)
	default:
		return writeOut(out, formatted)
	}
}

// writeBack rewrites path only when its formatting actually changed, so an
// already-formatted file keeps its modification time.
func writeBack(path filePath, original, formatted string) error {
	if formatted == original {
		return nil
	}
	if err := writeFile(string(path), []byte(formatted), fileMode); err != nil {
		return constants.ErrWriteFile.With(err, string(path))
	}
	return nil
}

// listChanged prints path when its formatting differs from disk.
func listChanged(out io.Writer, path filePath, original, formatted string) error {
	if formatted == original {
		return nil
	}
	return writeLine(out, string(path))
}

// format renders content through gomatic/go-sql and appends the trailing newline
// a source file carries. A parse failure comes back wrapped in
// [constants.ErrFormat].
func format(content []byte) (string, error) {
	out, err := formatter.New().Format(sql.SQL(content))
	if err != nil {
		return "", constants.ErrFormat.With(err)
	}
	return out + "\n", nil
}

// writeOut writes formatted SQL to out.
func writeOut(out io.Writer, formatted string) error {
	if _, err := io.WriteString(out, formatted); err != nil {
		return constants.ErrWriteFile.With(err, "stdout")
	}
	return nil
}

// writeLine writes one line followed by a newline to out.
func writeLine(out io.Writer, line string) error {
	if _, err := fmt.Fprintln(out, line); err != nil {
		return constants.ErrWriteFile.With(err, "stdout")
	}
	return nil
}
