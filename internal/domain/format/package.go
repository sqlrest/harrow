// Package format orchestrates harrow's formatting command. Run resolves the
// inputs (files, or standard input when none are named), formats each through
// gomatic/go-sql's formatter, and routes the result: to standard output, back
// into the file (--write), or to a list of changed paths (--list). It holds no
// CLI parsing or flag-definition logic — that stays in the app tier
// (cmd/harrow) — and no formatting logic — that lives in gomatic/go-sql.
package format
