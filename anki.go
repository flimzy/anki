// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

package anki

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/jmoiron/sqlx"
)

// Apkg manages state of an Anki package file during processing.
type Apkg struct {
	reader *zip.Reader
	closer *zip.ReadCloser
	sqlite *zip.File
	media  *zipIndex
	db     *DB
}

// ReadFile reads an *.apkg file, returning an Apkg struct for processing.
func ReadFile(f string) (*Apkg, error) {
	z, err := zip.OpenReader(f)
	if err != nil {
		return nil, err
	}
	a := &Apkg{
		reader: &z.Reader,
		closer: z,
	}
	return a, a.open()
}

// ReadBytes reads an *.apkg file from a bytestring, returning an Apkg struct
// for processing.
func ReadBytes(b []byte) (*Apkg, error) {
	r := bytes.NewReader(b)
	return ReadReader(r, int64(len(b)))
}

// ReadReader reads an *.apkg file from an io.Reader, returning an Apkg struct
// for processing.
func ReadReader(r io.ReaderAt, size int64) (*Apkg, error) {
	z, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	a := &Apkg{
		reader: z,
	}
	return a, a.open()
}

func (a *Apkg) open() error {
	if err := a.populateIndex(); err != nil {
		return err
	}
	rc, err := a.sqlite.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	db, err := OpenDB(rc)
	if err != nil {
		return err
	}
	a.db = db
	return nil
}

type zipIndex struct {
	index map[string]*zip.File
}

func (a *Apkg) ReadMediaFile(name string) ([]byte, error) {
	return a.media.ReadFile(name)
}

func (zi *zipIndex) ReadFile(name string) ([]byte, error) {
	zipFile, ok := zi.index[name]
	if !ok {
		return nil, errors.New("File `" + name + "` not found in zip index")
	}
	fh, err := zipFile.Open()
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(fh)
	return buf.Bytes(), nil
}

func (a *Apkg) populateIndex() error {
	index := &zipIndex{
		index: make(map[string]*zip.File),
	}
	for _, file := range a.reader.File {
		index.index[file.FileHeader.Name] = file
	}

	if sqlite, ok := index.index["collection.anki2"]; !ok {
		return errors.New("Unable to find `collection.anki2` in archive")
	} else {
		a.sqlite = sqlite
	}

	mediaFile, err := index.ReadFile("media")
	if err != nil {
		return err
	}

	mediaMap := make(map[string]string)
	if err := json.Unmarshal(mediaFile, &mediaMap); err != nil {
		return err
	}
	a.media = &zipIndex{
		index: make(map[string]*zip.File),
	}
	for idx, filename := range mediaMap {
		a.media.index[filename] = index.index[idx]
	}
	return nil
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
	var deletedDecks []ID
	if rows, err := a.db.Query("SELECT oid FROM graves WHERE type=2"); err != nil {
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
	if err := a.db.Get(collection, "SELECT * FROM col"); err != nil {
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
		ORDER BY id DESC
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
				WHEN 0 THEN NULL
				WHEN 1 THEN c.due
				WHEN 2 THEN c.due*24*60*60+(SELECT crt FROM col)
			END AS due,
			CASE
				WHEN c.ivl == 0 THEN NULL
				WHEN c.ivl < 0 THEN -ivl
				ELSE c.ivl*24*60*60
			END AS ivl,
			CASE c.type
				WHEN 0 THEN NULL
				WHEN 1 THEN c.odue
				WHEN 2 THEN c.odue*24*60*60+(SELECT crt FROM col)
			END AS odue
		FROM cards c
		LEFT JOIN graves g ON g.oid=c.id AND g.type=0
		WHERE g.oid IS NULL
		ORDER BY id DESC
	`)
	return &Cards{rows}, err
}

func (c *Cards) Card() (*Card, error) {
	card := &Card{}
	err := c.StructScan(card)
	return card, err
}

type Reviews struct {
	*sqlx.Rows
}

// Reviews returns a Reviews struct representing all of the reviews of
// non-deleted cards in the *.apkg package file, in reverse chronological
// order (newest first).
func (a *Apkg) Reviews() (*Reviews, error) {
	rows, err := a.db.Queryx(`
		SELECT r.id, r.cid, r.usn, r.ease, r.time, r.type,
			CAST(r.factor AS real)/1000 AS factor,
			CASE
				WHEN r.ivl < 0 THEN -ivl
				ELSE r.ivl*24*60*60
			END AS ivl,
			CASE
				WHEN r.lastIvl < 0 THEN -ivl
				ELSE r.lastIvl*24*60*60
			END AS lastIvl
		FROM revlog r
		LEFT JOIN graves g ON g.oid=r.cid AND g.type=0
		WHERE g.oid IS NULL
		ORDER BY id DESc
	`)
	return &Reviews{rows}, err
}

func (r *Reviews) Review() (*Review, error) {
	review := &Review{}
	err := r.StructScan(review)
	return review, err
}
