package format

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	errs "github.com/gomatic/go-error"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sqlrest/harrow/internal/constants"
)

const (
	unformatted = "SELECT   a,b   FROM t"
	formatted   = "select a, b from t\n"
)

// discardLogger is a logger that throws its output away.
func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
}

// run is Run with the throwaway context and logger filled in.
func run(cfg Config, in *bytes.Reader, out *bytes.Buffer, args ...string) (Result, error) {
	reader := in
	if reader == nil {
		reader = bytes.NewReader(nil)
	}
	paths := make([]FilePath, len(args))
	for i, arg := range args {
		paths[i] = FilePath(arg)
	}
	return Run(context.Background(), discardLogger(), cfg, reader, out, paths...)
}

// writeTemp writes content to a fresh temp file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in.sql")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// failer fails every read or write with its error.
type failer struct{ err error }

func (f failer) Read([]byte) (int, error)  { return 0, f.err }
func (f failer) Write([]byte) (int, error) { return 0, f.err }

func TestRunStreamFormatsStdinToStdout(t *testing.T) {
	var out bytes.Buffer
	_, err := run(Config{}, bytes.NewReader([]byte(unformatted)), &out)
	require.NoError(t, err)
	assert.Equal(t, formatted, out.String())
}

func TestRunFileToStdoutRecordsChange(t *testing.T) {
	var out bytes.Buffer
	result, err := run(Config{}, nil, &out, writeTemp(t, unformatted))
	require.NoError(t, err)
	assert.Equal(t, formatted, out.String())
	assert.Len(t, result.Changed, 1)
}

func TestRunWriteRewritesChangedFile(t *testing.T) {
	path := writeTemp(t, unformatted)
	var out bytes.Buffer
	result, err := run(Config{WriteEnabled: true}, nil, &out, path)
	require.NoError(t, err)
	assert.Empty(t, out.String(), "write mode writes the file, not stdout")
	onDisk, _ := os.ReadFile(path)
	assert.Equal(t, formatted, string(onDisk))
	assert.Len(t, result.Changed, 1)
}

func TestRunWriteLeavesFormattedFileUntouched(t *testing.T) {
	path := writeTemp(t, formatted)
	var out bytes.Buffer
	result, err := run(Config{WriteEnabled: true}, nil, &out, path)
	require.NoError(t, err)
	assert.Empty(t, result.Changed)
}

func TestRunListPrintsChangedPathOnly(t *testing.T) {
	changed := writeTemp(t, unformatted)
	already := writeTemp(t, formatted)
	var out bytes.Buffer
	_, err := run(Config{ListEnabled: true}, nil, &out, changed, already)
	require.NoError(t, err)
	assert.Equal(t, changed+"\n", out.String())
}

func TestRunParseErrorWrapsErrFormat(t *testing.T) {
	var out bytes.Buffer
	_, err := run(Config{}, bytes.NewReader([]byte("not valid ((")), &out)
	assert.ErrorIs(t, err, constants.ErrFormat)
}

func TestRunFileParseErrorWrapsErrFormat(t *testing.T) {
	var out bytes.Buffer
	_, err := run(Config{}, nil, &out, writeTemp(t, "not valid (("))
	assert.ErrorIs(t, err, constants.ErrFormat)
}

func TestRunMissingFileWrapsErrOpenFile(t *testing.T) {
	var out bytes.Buffer
	_, err := run(Config{}, nil, &out, "does-not-exist.sql")
	assert.ErrorIs(t, err, constants.ErrOpenFile)
}

func TestRunStdinReadErrorWrapsErrReadInput(t *testing.T) {
	const boom errs.Const = "boom"
	var out bytes.Buffer
	_, err := Run(context.Background(), discardLogger(), Config{}, failer{err: boom}, &out)
	assert.ErrorIs(t, err, constants.ErrReadInput)
}

func TestRunStdoutWriteErrorWrapsErrWriteFile(t *testing.T) {
	const boom errs.Const = "boom"
	_, err := Run(
		context.Background(),
		discardLogger(),
		Config{},
		bytes.NewReader([]byte(unformatted)),
		failer{err: boom},
	)
	assert.ErrorIs(t, err, constants.ErrWriteFile)
}

func TestRunListWriteErrorWrapsErrWriteFile(t *testing.T) {
	const boom errs.Const = "boom"
	_, err := Run(
		context.Background(),
		discardLogger(),
		Config{ListEnabled: true},
		bytes.NewReader(nil),
		failer{err: boom},
		FilePath(writeTemp(t, unformatted)),
	)
	assert.ErrorIs(t, err, constants.ErrWriteFile)
}

func TestRunWriteBackErrorWrapsErrWriteFile(t *testing.T) {
	const boom errs.Const = "boom"
	original := writeFile
	writeFile = func(string, []byte, os.FileMode) error { return boom }
	defer func() { writeFile = original }()

	var out bytes.Buffer
	_, err := run(Config{WriteEnabled: true}, nil, &out, writeTemp(t, unformatted))
	assert.ErrorIs(t, err, constants.ErrWriteFile)
}
