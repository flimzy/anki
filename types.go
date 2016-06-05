package anki

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"
)

type ID int64

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

func (i *ID) UnmarshalJSON(src []byte) error {
	var id interface{}
	if err := json.Unmarshal(src, &id); err != nil {
		return err
	}
	return i.Scan(id)
}

type TimestampSeconds time.Time
type TimestampMilliseconds time.Time

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
	*t = TimestampSeconds(time.Unix(seconds, 0))
	return nil
}

func (t *TimestampSeconds) UnmarshalJSON(src []byte) error {
	var ts interface{}
	if err := json.Unmarshal(src, &ts); err != nil {
		return err
	}
	return t.Scan(ts)
}

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
	*t = TimestampMilliseconds(time.Unix(ms/1000, ms%1000))
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

type DurationSeconds time.Duration

func (d *DurationSeconds) Scan(src interface{}) error {
	seconds, err := scanInt64(src)
	*d = DurationSeconds(time.Duration(seconds) * time.Second)
	return err
}

type DurationMinutes time.Duration

func (d *DurationMinutes) Scan(src interface{}) error {
	min, err := scanInt64(src)
	*d = DurationMinutes(time.Duration(min) * time.Minute)
	return err
}

type DurationDays int

func (d *DurationDays) Scan(src interface{}) error {
	days, err := scanInt64(src)
	*d = DurationDays(int(days))
	return err
}

type BoolInt bool

func (b *BoolInt) Scan(src interface{}) error {
	var tf bool
	switch src.(type) {
	case float64:
		// Only 0 is false
		tf = src.(float64) != 0
	case int64:
		// Only 0 is false
		tf = src.(int64) != 0
	case nil:
		// Nil is false
		tf = false
	default:
		return errors.New("Incompatible type for BoolInt")
	}
	*b = BoolInt(tf)
	return nil
}

func (b *BoolInt) UnmarshalJSON(src []byte) error {
	var tmp interface{}
	if err := json.Unmarshal(src, &tmp); err != nil {
		return err
	}
	return b.Scan(tmp)
}
