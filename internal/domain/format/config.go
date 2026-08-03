package format

import "io"

// Config holds the formatting command's flags and injected streams. Input
// paths arrive as positional arguments, not config.
type Config struct {
	// In is the stream formatted when no paths are given.
	In io.Reader
	// Out receives formatted SQL and --list paths.
	Out          io.Writer
	WriteEnabled writeEnabled
	ListEnabled  listEnabled
}
