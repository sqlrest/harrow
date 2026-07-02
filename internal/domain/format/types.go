package format

// FilePath is a SQL file to format. Paths arrive as positional arguments, so
// the app tier converts each one and passes it to [Run]; it is the one domain
// type the CLI must name, hence the export.
type FilePath string

// Named types for the command's flags, bound by the CLI via pointer conversion.
type (
	writeEnabled bool // writeEnabled rewrites changed files in place (--write).
	listEnabled  bool // listEnabled prints the paths that would change (--list).
)

// SQL text at its two stages: as read, and as canonically rendered.
type (
	sqlText      string // sqlText is SQL source exactly as read from a file or stream.
	formattedSQL string // formattedSQL is SQL rendered in harrow's canonical layout.
)
