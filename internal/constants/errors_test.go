package constants

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The error mechanism (Error/With) is exercised in gomatic/go-error; this test
// only verifies that harrow's sentinels carry their text and stay matchable with
// errors.Is once wrapped — the contract consumers rely on.
func TestSentinels(t *testing.T) {
	t.Parallel()
	want := assert.New(t)

	want.Equal("failed to format SQL", ErrFormat.Error())
	want.Equal("failed to open file", ErrOpenFile.Error())

	wrapped := fmt.Errorf("%w: %s", ErrWriteFile, "out.sql")
	want.ErrorIs(wrapped, ErrWriteFile)
	want.NotErrorIs(wrapped, ErrFormat)
}
