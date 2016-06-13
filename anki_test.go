package anki

import (
	"os"
	"testing"
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

	collection, err := apkg.Collection()
	if err != nil {
		t.Fatalf("Error getting collection: %s", err)
	}
	if collection.Config.CollapseTime != 1200 {
		t.Fatalf("Spot-check failed")
	}

	notes, err := apkg.Notes()
	if err != nil {
		t.Fatalf("Error fetching notes: %s", err)
	}
	for notes.Next() {
		note, err := notes.Note()
		if err != nil {
			t.Fatalf("Error reading note: %s", err)
		}
		if note.ID != 1388721680877 {
			t.Fatalf("note spot-check failed. Expected ID 1388721680877, got %d", note.ID)
		}
		if note.Checksum != 1090091728 {
			t.Fatalf("note spot-check failed. Expected checksum 1090091728, got %d", note.Checksum)
		}
	}
	if err := notes.Close(); err != nil {
		t.Fatalf("Error closing Notes: %s", err)
	}

	reviews, err := apkg.Reviews()
	if err != nil {
		t.Fatalf("Error fetching reviews: %s", err)
	}
	for reviews.Next() {
		review, err := reviews.Review()
		if err != nil {
			t.Fatalf("Error reading review: %s", err)
		}
		if review.CardID != 1388721683902 {
			t.Fatalf("review spot-check failed. Expected cid 1388721683902, got %d", review.CardID)
		}
	}
	if err := reviews.Close(); err != nil {
		t.Fatalf("Error closing Reviews: %s", err)
	}

	cards, err := apkg.Cards()
	if err != nil {
		t.Fatalf("Error fetching cards: %s", err)
	}
	for cards.Next() {
		card, err := cards.Card()
		if err != nil {
			t.Fatalf("Error reading card: %s", err)
		}
		if card.ID != 1388721683902 {
			t.Fatalf("card spot-check failed. Expected ID 1388721683902, got %d", card.ID)
		}
	}
	if err := cards.Close(); err != nil {
		t.Fatalf("Error closing Cards: %s", err)
	}

	if err := apkg.Close(); err != nil {
		t.Fatalf("Error closing apkg from Reader: %s", err)
	}
}
