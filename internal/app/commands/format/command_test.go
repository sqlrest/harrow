package format

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	app "github.com/gomatic/go-app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"github.com/sqlrest/harrow/internal/constants"
)

// testCommand builds the command wired to a quiet logger and a captured stdout.
func testCommand(stdout *bytes.Buffer) *cli.Command {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelWarn}))
	c := Command()
	c.Writer = stdout
	c.Metadata = map[string]any{app.LoggerMetadataKey: logger}
	return c
}

// writeTemp writes content to a fresh temp file and returns its path.
func writeTemp(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in.sql")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestCommandNameAndFlags(t *testing.T) {
	c := testCommand(&bytes.Buffer{})
	assert.Equal(t, name, c.Name)
	flagNames := map[string]bool{}
	for _, f := range c.Flags {
		for _, n := range f.Names() {
			flagNames[n] = true
		}
	}
	for _, want := range []string{"write", "w", "list", "l"} {
		assert.True(t, flagNames[want], "missing flag %q", want)
	}
}

func TestActionFormatsFileToStdout(t *testing.T) {
	path := writeTemp(t, "SELECT   a   FROM t")
	var stdout bytes.Buffer
	require.NoError(t, testCommand(&stdout).Run(context.Background(), []string{name, path}))
	assert.Equal(t, "select a from t\n", stdout.String())
}

func TestActionFormatsStdinWhenNoArgs(t *testing.T) {
	original := stdin
	t.Cleanup(func() { stdin = original })
	stdin = bytes.NewReader([]byte("SELECT   a   FROM t"))

	var stdout bytes.Buffer
	require.NoError(t, testCommand(&stdout).Run(context.Background(), []string{name}))
	assert.Equal(t, "select a from t\n", stdout.String())
}

func TestActionListsChangedFiles(t *testing.T) {
	path := writeTemp(t, "SELECT   a   FROM t")
	var stdout bytes.Buffer
	require.NoError(t, testCommand(&stdout).Run(context.Background(), []string{name, "--list", path}))
	assert.Equal(t, path+"\n", stdout.String())
}

func TestActionPropagatesError(t *testing.T) {
	path := writeTemp(t, "not valid ((")
	err := testCommand(&bytes.Buffer{}).Run(context.Background(), []string{name, path})
	assert.ErrorIs(t, err, constants.ErrFormat)
}
