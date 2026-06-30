package format

// Config holds the formatting command's flags. Input paths arrive as positional
// arguments, not config.
type Config struct {
	WriteEnabled writeEnabled
	ListEnabled  listEnabled
}
