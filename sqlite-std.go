// +build !js

// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

package anki

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // The sqlite3 driver
)

// DB is a wrapper for the underlying SQLite database.
type DB struct {
	*sqlx.DB
	tmpFile string
}

// Close closes the database handle.
func (db *DB) Close() (e error) {
	if db.tmpFile != "" {
		if err := os.Remove(db.tmpFile); err != nil {
			fmt.Printf("Cannot remove file: %s", err)
			e = err
		}
	}
	if db.DB != nil {
		if err := db.DB.Close(); err != nil {
			e = err
		}
	}
	return
}

// OpenDB reads an SQLite database file on src, and returns an opened database
// handle.
func OpenDB(src io.Reader) (db *DB, e error) {
	db = &DB{}
	dbFile, err := dumpToTemp(src)
	db.tmpFile = dbFile
	if err != nil {
		return db, err
	}
	sqldb, err := sqlx.Connect("sqlite3", dbFile)
	if err != nil {
		return db, err
	}
	db.DB = sqldb
	return db, nil
}

func dumpToTemp(src io.Reader) (string, error) {
	tmp, err := ioutil.TempFile("/tmp", "anki-sqlite3-")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := io.Copy(tmp, src); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}
