// Package harrow is the module root marker for the harrow CLI. harrow formats
// PostgreSQL SQL the way a harrow levels a field: it reads SQL, lays it out in a
// canonical style, and — by leaning on gomatic/go-sql's verification gate — never
// changes a statement's meaning or drops a comment.
package harrow
