// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

package anki

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ID represents an Anki object ID (deck, card, note, etc) as an int64.
type ID int64

// Scan implements the sql.Scanner interface for the ID type.
func (i *ID) Scan(src interface{}) error {
	var id int64
	switch x := src.(type) {
	case float64:
		id = int64(src.(float64))
	case int64:
		id = src.(int64)
	case string:
		var err error
		id, err = strconv.ParseInt(src.(string), 10, 64)
		if err != nil {
			return err
		}
	case nil:
		return nil
	default:
		return fmt.Errorf("Incompatible type for ID: %s", x)
	}
	*i = ID(id)
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for the ID type.
func (i *ID) UnmarshalJSON(src []byte) error {
	var id interface{}
	if err := json.Unmarshal(src, &id); err != nil {
		return err
	}
	return i.Scan(id)
}

// TimestampSeconds represents a time.Time value stored as seconds.
type TimestampSeconds time.Time

// TimestampMilliseconds represents a time.Time value stored as milliseconds.
type TimestampMilliseconds time.Time

// Scan implements the sql.Scanner interface for the TimestampSeconds type.
func (t *TimestampSeconds) Scan(src interface{}) error {
	var seconds int64
	switch x := src.(type) {
	case float64:
		seconds = int64(src.(float64))
	case int64:
		seconds = src.(int64)
	case nil:
		return nil
	default:
		return fmt.Errorf("Incompatible type for TimestampSeconds: %s", x)
	}
	*t = TimestampSeconds(time.Unix(seconds, 0).UTC())
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for the
// TimestampSeconds type.
func (t *TimestampSeconds) UnmarshalJSON(src []byte) error {
	var ts interface{}
	if err := json.Unmarshal(src, &ts); err != nil {
		return err
	}
	return t.Scan(ts)
}

// Scan implements the sql.Scanner interface for the TimestampMilliseconds
// type.
func (t *TimestampMilliseconds) Scan(src interface{}) error {
	var ms int64
	switch src.(type) {
	case float64:
		ms = int64(src.(float64))
	case int64:
		ms = src.(int64)
	case nil:
		return nil
	default:
		return errors.New("Incompatible type for TimestampMillieconds")
	}
	*t = TimestampMilliseconds(time.Unix(ms/1000, ms%1000).UTC())
	return nil
}

func scanInt64(src interface{}) (int64, error) {
	var num int64
	switch src.(type) {
	case float64:
		num = int64(src.(float64))
	case int64:
		num = src.(int64)
	default:
		return 0, errors.New("Incompatible type for int64")
	}
	return num, nil
}

// DurationMilliseconds represents a time.Duration value stored as
// milliseconds.
type DurationMilliseconds time.Duration

// Scan implements the sql.Scanner interface for the DurationMilliseconds type.
func (d *DurationMilliseconds) Scan(src interface{}) error {
	ms, err := scanInt64(src)
	*d = DurationMilliseconds(time.Duration(ms) * time.Millisecond)
	return err
}

// DurationSeconds represents a time.Duration value stored as seconds.
type DurationSeconds time.Duration

// Scan implements the sql.Scanner interface for the DurationSeconds type.
func (d *DurationSeconds) Scan(src interface{}) error {
	seconds, err := scanInt64(src)
	*d = DurationSeconds(time.Duration(seconds) * time.Second)
	return err
}

// DurationMinutes represents a time.Duration value stored as minutes.
type DurationMinutes time.Duration

// Scan implements the sql.Scanner interface for the DurationMinutes type.
func (d *DurationMinutes) Scan(src interface{}) error {
	min, err := scanInt64(src)
	*d = DurationMinutes(time.Duration(min) * time.Minute)
	return err
}

// DurationDays represents a duration in days.
type DurationDays int

// Scan implements the sql.Scanner interface for the DurationDays type.
func (d *DurationDays) Scan(src interface{}) error {
	days, err := scanInt64(src)
	*d = DurationDays(int(days))
	return err
}

// BoolInt represents a boolean value stored as an int
type BoolInt bool

// Scan implements the sql.Scanner interface for the BoolInt type.
func (b *BoolInt) Scan(src interface{}) error {
	var tf bool
	switch t := src.(type) {
	case bool:
		tf = t
	case float64:
		// Only 0 is false
		tf = t != 0
	case int64:
		// Only 0 is false
		tf = t != 0
	case nil:
		// Nil is false
		tf = false
	default:
		return fmt.Errorf("Incompatible type '%T' for BoolInt", src)
	}
	*b = BoolInt(tf)
	return nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for the BoolInt
// type.
func (b *BoolInt) UnmarshalJSON(src []byte) error {
	var tmp interface{}
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	return b.Scan(tmp)
}
