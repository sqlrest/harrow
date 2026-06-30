// Package constants declares harrow's sentinel error values. The error mechanism
// (the matchable string type) lives in the shared gomatic/go-error library;
// these values are harrow's own.
package constants

// Imported bare (the package is named error); this file declares only sentinels.
import errs "github.com/gomatic/go-error"

// Keep these constants sorted alphabetically.
const (
	ErrFormat    errs.Const = "failed to format SQL"
	ErrOpenFile  errs.Const = "failed to open file"
	ErrReadInput errs.Const = "failed to read input"
	ErrWriteFile errs.Const = "failed to write file"
)
