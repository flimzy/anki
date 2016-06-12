package anki

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"io"
	// 	"github.com/davecgh/go-spew/spew"
)

type Apkg struct {
	reader *zip.Reader
	closer *zip.ReadCloser
	index  map[string]*zip.File
	db     *DB
}

func ReadFile(f string) (*Apkg, error) {
	z, err := zip.OpenReader(f)
	if err != nil {
		return nil, err
	}
	a := &Apkg{
		reader: &z.Reader,
		closer: z,
	}
	a.populateIndex()
	return a, nil
}

func ReadReader(r io.ReaderAt, size int64) (*Apkg, error) {
	z, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	a := &Apkg{
		reader: z,
	}
	a.populateIndex()
	return a, nil
}

func (a *Apkg) populateIndex() {
	a.index = make(map[string]*zip.File)
	for _, file := range a.reader.File {
		a.index[file.FileHeader.Name] = file
	}
}

// Close closes any opened resources (io.Reader, SQLite handles, etc). Any
// subsequent calls to extant objects (Collection, Decks, Notes, etc) which
// depend on these resources may fail. Only call this method after you're
// completely done reading the Apkg file.
func (a *Apkg) Close() (e error) {
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			e = err
		}
	}
	if a.closer != nil {
		if err := a.closer.Close(); err != nil {
			e = err
		}
	}
	return
}

func (a *Apkg) Collection() (*Collection, error) {
	file, ok := a.index["collection.anki2"]
	if !ok {
		return nil, errors.New("Did not find 'collection.anki2'. Invalid Anki package")
	}
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	db, err := OpenDB(rc)
	if err != nil {
		return nil, err
	}
	a.db = db
	var deletedDecks []ID
	if rows, err := db.Query("SELECT oid FROM graves WHERE type=2"); err != nil {
		return nil, err
	} else {
		for rows.Next() {
			var id *ID
			if err := rows.Scan(id); err != nil {
				return nil, err
			}
			deletedDecks = append(deletedDecks, *id)
		}
	}
	collection := &Collection{}
	if err := db.Get(collection, "SELECT * FROM col"); err != nil {
		return nil, err
	}
	for _, deck := range collection.Decks {
		for _, deleted := range deletedDecks {
			if deck.ID == deleted {
				delete(collection.Decks, deck.ID)
				continue
			}
		}
		conf, ok := collection.DeckConfigs[deck.ConfigID]
		if !ok {
			return nil, fmt.Errorf("Deck %d references non-existent config %d", deck.ID, deck.ConfigID)
		}
		deck.Config = conf
	}
	return collection, nil
}

// Notes is a wrapper around sqlx.Rows, which means that any standard sqlx.Rows
// or sql.Rows methods may be called on it. Generally, you should only ever
// need to call Next() and Close(), in addition to Note() which is defined in
// this package.
type Notes struct {
	*sqlx.Rows
}

// Notes returns a Notes struct representing all of the Note rows in the *.apkg
// package file.
func (a *Apkg) Notes() (*Notes, error) {
	rows, err := a.db.Queryx(`
		SELECT n.id, n.guid, n.mid, n.mod, n.usn, n.tags, n.flds, n.sfld,
			CAST(n.csum AS text) AS csum -- Work-around for SQL.js trying to treat this as a float
		FROM notes n
		LEFT JOIN graves g ON g.oid=n.id AND g.type=1
		ORDER BY id
	`)
	return &Notes{rows}, err
}

// Note is a simple wrapper around sqlx's StructScan(), which returns a Note
// struct populated from the database.
func (n *Notes) Note() (*Note, error) {
	note := &Note{}
	err := n.StructScan(note)
	return note, err
}

// Cards is a wrapper around sqlx.Rows, which means that any standard sqlx.Rows
// or sql.Rows methods may be called on it. Generally, you should only ever
// need to call Next() and Close(), in addition to Card() which is defined in
// this package.
type Cards struct {
	*sqlx.Rows
}

// Cards returns a Cards struct represeting all of the non-deleted cards in the
// *.apkg package file.
func (a *Apkg) Cards() (*Cards, error) {
	rows, err := a.db.Queryx(`
		SELECT c.id, c.nid, c.did, c.ord, c.mod, c.usn, c.type, c.queue, c.reps, c.lapses, c.left, c.odid,
			CAST(c.factor AS real)/1000 AS factor,
			CASE c.type
				WHEN 1 THEN 0
				WHEN 2 THEN c.due*24*60*60*(SELECT crt FROM col)
				ELSE c.due
			END AS due,
			CASE
				WHEN c.ivl < 0 THEN -ivl
				ELSE c.ivl*24*60*60
			END AS ivl,
			CASE c.type
				WHEN 1 THEN 0
				WHEN 2 THEN c.odue*24*60*60*(SELECT crt FROM col)
				ELSE c.odue
			END AS odue
		FROM cards c
		LEFT JOIN graves g ON g.oid=c.id AND g.type=0
		WHERE g.oid IS NULL
		ORDER BY id
	`)
	return &Cards{rows}, err
}

func (c *Cards) Card() (*Card, error) {
	card := &Card{}
	err := c.StructScan(card)
	return card, err
}
