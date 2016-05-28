package anki

import (
	"testing"
)

func TestReadFile(t *testing.T) {
	if apkg, err := ReadFile("t/Test.apkg"); err != nil {
		t.Fatalf("Error opening Test.apkg")
	} else {
		if err := apkg.Close(); err != nil {
			t.Fatalf("Error closing apkg: %s", err)
		}
	}
	if _, err := ReadFile("does not exist"); err.Error() != "open does not exist: no such file or directory" {
		t.Fatalf("Unexpected error trying to open non-existant file: %s", err)
	}
}

