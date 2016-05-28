package anki

import (
	"io"
	"archive/zip"
)

type Apkg struct {
	readCloser *zip.ReadCloser
}


func ReadFile(f string) (*Apkg,error) {
	z, err := zip.OpenReader(f)
	if err != nil {
		return nil, err
	}
	return &Apkg{
		readCloser: z,
	}, nil
}

func ReadReader(r io.ReaderAt, size int64) (*Apkg,error) {
	return nil, nil
}

func (a *Apkg) Close() error {
	if a.readCloser == nil {
		return nil // Nothing to close
	}
	return a.readCloser.Close()
}
