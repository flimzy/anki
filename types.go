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
	ID             int                    `db:"id"`
	Created        *TimestampSeconds      `db:"crt"`
	Modified       *TimestampMilliseconds `db:"mod"`
	SchemaModified *TimestampMilliseconds `db:"scm"`
	Ver            int                    `db:"ver"`
	UpdateSequence int                    `db:"usn"`
	LastSync       *TimestampMilliseconds `db:"ls"`
	Config         string                 `db:"conf"`
	Models         string                 `db:"models"`
	Decks          string                 `db:"decks"`
	DeckConfig     string                 `db:"dconf"`
	Tags           string                 `db:"tags"`
	Dty            int                    `db:"dty"` // Unused
}
