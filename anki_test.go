package anki

import (
	"os"
	// 	"fmt"
	"testing"
	// 	"github.com/davecgh/go-spew/spew"
)

const ApkgFile = "t/Test.apkg"

func TestReadFile(t *testing.T) {
	if _, err := ReadFile("does not exist"); err.Error() != "open does not exist: no such file or directory" {
		t.Fatalf("Unexpected error trying to open non-existant file: %s", err)
	}
	apkg, err := ReadFile(ApkgFile)
	if err != nil {
		t.Fatalf("Error opening test file: %s", err)
	}

	if err := apkg.Close(); err != nil {
		t.Fatalf("Error closing apkg: %s", err)
	}
}

func TestReadReader(t *testing.T) {
	file, err := os.Open(ApkgFile)
	if err != nil {
		t.Fatalf("Error opening test file: %s", err)
	}
	fi, err := file.Stat()
	if err != nil {
		t.Fatalf("Error statting file: %s", err)
	}
	apkg, err := ReadReader(file, fi.Size())
	if err != nil {
		t.Fatalf("Error opening apkg from Reader: %s", err)
	}
	if err := apkg.Close(); err != nil {
		t.Fatalf("Error closing apkg from Reader: %s", err)
	}

	collection, err := apkg.Collection()
	if err != nil {
		t.Fatalf("Error getting collection: %s", err)
	}
	if collection.Config.CollapseTime != 1200 {
		t.Fatalf("Spot-check failed")
	}

	if err := apkg.Close(); err != nil {
		t.Fatalf("Error closing Apkg: %v", err)
	}
}
