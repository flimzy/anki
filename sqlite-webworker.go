// +build js,webworker

// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

package anki

import (
	"database/sql"
	"io"
)

type DB struct {
	*sql.DB
}

func readSQLite(file io.Reader) (*Collection, error) {
	return nil, nil
}
