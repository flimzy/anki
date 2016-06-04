// +build js,!webworker

package anki

import (
	"io"

	"database/sql"
	"github.com/flimzy/go-sql.js"
)

type DB struct {
	*sql.DB
}

var sqliteReader io.Reader

func init() {
	sql.Register("anki-reader", &sqljs.SQLJSDriver{Reader: sqliteReader})
}

func OpenDB(dbFile io.Reader) (*sql.DB, error) {
	sqliteReader = dbFile
	return sql.Open("anki-reader", "")
}
