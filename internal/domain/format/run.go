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
	Changed []FilePath `json:"changed"`
}

// Run formats the SQL named in paths, or standard input when no paths are
// given. With no paths it reads in and writes the formatted SQL to out. With
// paths it formats each file: WriteEnabled rewrites changed files in place,
// ListEnabled prints the paths that would change, and otherwise the formatted
// SQL goes to out. A file that won't parse stops the run with
// [constants.ErrFormat].
func Run(
	_ context.Context,
	logger *slog.Logger,
	cfg Config,
	in io.Reader,
	out io.Writer,
	paths ...FilePath,
) (Result, error) {
	if len(paths) == 0 {
		return Result{}, formatStream(in, out)
	}

	var result Result
	for _, path := range paths {
		next, err := formatFile(cfg, out, path, result)
		if err != nil {
			return Result{}, err
		}
		result = next
	}
	logger.Info("Formatting complete.", "files", len(paths), "changed", len(result.Changed))
	return result, nil
}

// formatStream formats everything read from in and writes it to out.
func formatStream(in io.Reader, out io.Writer) error {
	content, err := io.ReadAll(in)
	if err != nil {
		return constants.ErrReadInput.With(err)
	}
	formatted, err := format(sqlText(content))
	if err != nil {
		return err
	}
	return writeOut(out, formatted)
}

// formatFile formats one file and routes the output according to cfg, returning
// result grown by the path when its formatting differs from what's on disk.
func formatFile(cfg Config, out io.Writer, path FilePath, result Result) (Result, error) {
	content, err := os.ReadFile(string(path))
	if err != nil {
		return Result{}, constants.ErrOpenFile.With(err, string(path))
	}
	original := sqlText(content)
	formatted, err := format(original)
	if err != nil {
		return Result{}, err
	}
	if string(formatted) != string(original) {
		result.Changed = append(result.Changed, path)
	}
	return result, emit(cfg, out, path, original, formatted)
}

// emit routes a formatted file: rewritten in place, listed if changed, or
// printed to out.
func emit(cfg Config, out io.Writer, path FilePath, original sqlText, formatted formattedSQL) error {
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
func writeBack(path FilePath, original sqlText, formatted formattedSQL) error {
	if string(formatted) == string(original) {
		return nil
	}
	if err := writeFile(string(path), []byte(formatted), fileMode); err != nil {
		return constants.ErrWriteFile.With(err, string(path))
	}
	return nil
}

// listChanged prints path when its formatting differs from disk.
func listChanged(out io.Writer, path FilePath, original sqlText, formatted formattedSQL) error {
	if string(formatted) == string(original) {
		return nil
	}
	return writeLine(out, path)
}

// format renders content through gomatic/go-sql and appends the trailing newline
// a source file carries. A parse failure comes back wrapped in
// [constants.ErrFormat].
func format(content sqlText) (formattedSQL, error) {
	out, err := formatter.New().Format(sql.SQL(content))
	if err != nil {
		return "", constants.ErrFormat.With(err)
	}
	return formattedSQL(out + "\n"), nil
}

// writeOut writes formatted SQL to out.
func writeOut(out io.Writer, formatted formattedSQL) error {
	if _, err := io.WriteString(out, string(formatted)); err != nil {
		return constants.ErrWriteFile.With(err, "stdout")
	}
	return nil
}

// writeLine writes one line followed by a newline to out.
func writeLine(out io.Writer, line FilePath) error {
	if _, err := fmt.Fprintln(out, string(line)); err != nil {
		return constants.ErrWriteFile.With(err, "stdout")
	}
	return nil
}
