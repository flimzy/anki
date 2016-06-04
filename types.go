package anki

import (
	"errors"
	"fmt"
	"time"
)

type TimestampSeconds time.Time
type TimestampMilliseconds time.Time

func (t *TimestampSeconds) Scan(src interface{}) error {
	switch x := src.(type) {
	case float64:
		*t = TimestampSeconds(time.Unix(int64(src.(float64)), 0))
	case int64:
		*t = TimestampSeconds(time.Unix(src.(int64), 0))
	default:
		return fmt.Errorf("Incompatible type for TimestampSeconds: %s", x)
	}
	return nil
}

func (t *TimestampMilliseconds) Scan(src interface{}) error {
	switch src.(type) {
	case float64:
		ms := src.(float64)
		*t = TimestampMilliseconds(time.Unix(int64(ms/1000), int64(ms)%1000))
	case int64:
		ms := src.(int64)
		*t = TimestampMilliseconds(time.Unix(ms/1000, ms%1000))
	default:
		return errors.New("Incompatible type for TimestampMillieconds")
	}
	return nil
}

type Collection struct {
	ID             int                    `db:"id"`     // Primary key; should always be 1, as there's only ever one collection per *.apkg file
	Created        *TimestampSeconds      `db:"crt"`    // Created timestamp (seconds)
	Modified       *TimestampMilliseconds `db:"mod"`    // Last modified timestamp (milliseconds)
	SchemaModified *TimestampMilliseconds `db:"scm"`    // Schema modification time (milliseconds)
	Version        int                    `db:"ver"`    // Version?
	Dirty          int                    `db:"dty"`    // Dirty? No longer used. See https://github.com/dae/anki/blob/master/anki/collection.py#L90
	UpdateSequence int                    `db:"usn"`    // update sequence number. used to figure out diffs when syncing
	LastSync       *TimestampMilliseconds `db:"ls"`     // Last sync time (milliseconds)
	Config         string                 `db:"conf"`   // JSON blob containing configuration options
	Models         string                 `db:"models"` // JSON array of json objects containing the models (aka Note types)
	Decks          string                 `db:"decks"`  // JSON array of json objects containing decks
	DeckConfig     string                 `db:"dconf"`  // JSON blob containing deck configuration options
	Tags           string                 `db:"tags"`   // a cache of tags used in the collection
}
