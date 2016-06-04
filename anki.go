package anki

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
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

// Close closes any opened resources (io.Reader, SQLite handles, etc)
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
	collection := &Collection{}
	if err := db.Get(collection, "SELECT * FROM col"); err != nil {
		return nil, err
	}
	fmt.Printf("%v\n", collection)
	return collection, nil
}
