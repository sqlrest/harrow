package main

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

const appName = "harrow"

// testApp builds the CLI wired to a quiet logger and a captured stdout buffer.
func testApp(stdout *bytes.Buffer) *cli.Command {
	logger := slog.New(slog.NewTextHandler(&bytes.Buffer{}, &slog.HandlerOptions{Level: slog.LevelWarn}))
	cliApp := createApp(func(_ *cli.Command) *slog.Logger { return logger })
	cliApp.Writer = stdout
	return cliApp
}

func TestVersionFlagPrintsVersion(t *testing.T) {
	var stdout bytes.Buffer
	require.NoError(t, testApp(&stdout).Run(context.Background(), []string{appName, "--version"}))
	assert.Contains(t, stdout.String(), version)
}

func TestCreateAppNameVersionAndFlags(t *testing.T) {
	cliApp := testApp(&bytes.Buffer{})
	assert.Equal(t, appName, cliApp.Name)
	assert.Equal(t, version, cliApp.Version)
	flagNames := map[string]bool{}
	for _, f := range cliApp.Flags {
		for _, n := range f.Names() {
			flagNames[n] = true
		}
	}
	for _, want := range []string{"write", "w", "list", "l", "log-level", "log-format"} {
		assert.True(t, flagNames[want], "missing flag %q", want)
	}
}

func TestFormatActionFormatsFileToStdout(t *testing.T) {
	path := filepath.Join(t.TempDir(), "in.sql")
	require.NoError(t, os.WriteFile(path, []byte("SELECT   a   FROM t"), 0o600))

	var stdout bytes.Buffer
	require.NoError(t, testApp(&stdout).Run(context.Background(), []string{appName, path}))
	assert.Equal(t, "select a from t\n", stdout.String())
}

func TestFormatActionPropagatesError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.sql")
	require.NoError(t, os.WriteFile(path, []byte("not valid (("), 0o600))
	err := testApp(&bytes.Buffer{}).Run(context.Background(), []string{appName, path})
	assert.ErrorIs(t, err, constants.ErrFormat)
}

func TestRunExitCodes(t *testing.T) {
	original := appCreator
	t.Cleanup(func() { appCreator = original })

	appCreator = func(app.GetLoggerFunc) *cli.Command {
		return &cli.Command{Name: appName, Writer: &bytes.Buffer{}}
	}
	assert.Equal(t, 0, run([]string{appName}), "successful run exits 0")

	appCreator = func(app.GetLoggerFunc) *cli.Command {
		return &cli.Command{
			Name:   appName,
			Writer: &bytes.Buffer{},
			Action: func(context.Context, *cli.Command) error { return constants.ErrFormat },
		}
	}
	assert.Equal(t, 1, run([]string{appName}), "failed run exits 1")
}

func TestMainEntry(t *testing.T) {
	originalCreator, originalExit, originalArgs := appCreator, osExit, os.Args
	t.Cleanup(func() { appCreator, osExit, os.Args = originalCreator, originalExit, originalArgs })

	var code int
	osExit = func(c int) { code = c }
	appCreator = func(app.GetLoggerFunc) *cli.Command {
		return &cli.Command{Name: appName, Writer: &bytes.Buffer{}}
	}
	os.Args = []string{appName}

	main()
	assert.Equal(t, 0, code)
}

func TestProductionLogger(t *testing.T) {
	assert.NotNil(t, productionLogger(nil))
}
