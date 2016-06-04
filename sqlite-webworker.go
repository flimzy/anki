// +build js,webworker

package anki

import (
	"io"
)

type DB struct {
	*sql.DB
}

func readSQLite(file io.Reader) (*Collection, error) {
	return nil, nil
}
