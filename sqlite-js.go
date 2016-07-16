// +build js,!webworker

// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html


package anki

import (
	"io"

	"github.com/flimzy/go-sql.js"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	*sqlx.DB
}

func OpenDB(dbFile io.Reader) (*DB, error) {
	sqljs.AddReader("collection.anki2", dbFile)
	db, err := sqlx.Connect("sqljs", "collection.anki2")
	if err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
