package format

// Named types for the command's inputs, bound by the CLI via pointer conversion.
type (
	filePath     string // filePath is a SQL file to format (positional arg).
	writeEnabled bool   // writeEnabled rewrites changed files in place (--write).
	listEnabled  bool   // listEnabled prints the paths that would change (--list).
)
