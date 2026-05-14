package util

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ErrInvalidDate indicates a string is not a valid YYYY-MM-DD date.
var ErrInvalidDate = errors.New("util: invalid calendar date (want YYYY-MM-DD)")

// Date is a calendar day (no clock, no location). JSON and text use "YYYY-MM-DD".
// Stored internally as UTC midnight of that day for stable equality and ordering.
type Date struct {
	t time.Time // always 00:00:00 UTC when non-zero
}

// ParseDate parses a date string. Whitespace is trimmed; layout is time.DateOnly (YYYY-MM-DD).
func ParseDate(s string) (Date, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Date{}, fmt.Errorf("%w: empty", ErrInvalidDate)
	}
	tt, err := time.Parse(time.DateOnly, s)
	if err != nil {
		return Date{}, fmt.Errorf("%w: %q", ErrInvalidDate, s)
	}
	y, m, d := tt.Date()
	return Date{t: time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}, nil
}

// MustParseDate is like ParseDate but panics on error (for tests and constants).
func MustParseDate(s string) Date {
	d, err := ParseDate(s)
	if err != nil {
		panic(err)
	}
	return d
}

// DateFromTime extracts the calendar date in UTC (ignores sub-day precision).
func DateFromTime(tt time.Time) Date {
	if tt.IsZero() {
		return Date{}
	}
	u := tt.UTC()
	y, m, d := u.Date()
	return Date{t: time.Date(y, m, d, 0, 0, 0, 0, time.UTC)}
}

// Today returns the current calendar date in UTC.
func Today() Date { return DateFromTime(time.Now()) }

func (d Date) IsZero() bool { return d.t.IsZero() }

// String returns "YYYY-MM-DD" or empty if zero.
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.t.UTC().Format(time.DateOnly)
}

// Time returns UTC midnight of this date; zero Date yields zero time.Time.
func (d Date) Time() time.Time { return d.t }

// MarshalJSON emits a JSON string "YYYY-MM-DD" or null for zero.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.String())
}

// UnmarshalJSON accepts a JSON string "YYYY-MM-DD", or null for zero.
func (d *Date) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		*d = Date{}
		return nil
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	if strings.TrimSpace(s) == "" {
		*d = Date{}
		return nil
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalText implements encoding.TextMarshaler (query / forms).
func (d Date) MarshalText() ([]byte, error) {
	if d.IsZero() {
		return nil, nil
	}
	return []byte(d.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (d *Date) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		*d = Date{}
		return nil
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// Value implements driver.Valuer (PostgreSQL date / text).
func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.String(), nil
}

// Scan implements sql.Scanner (nil, string, []byte, time.Time).
func (d *Date) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*d = Date{}
		return nil
	case time.Time:
		*d = DateFromTime(v)
		return nil
	case []byte:
		return d.Scan(string(v))
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			*d = Date{}
			return nil
		}
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	default:
		return fmt.Errorf("util.Date: cannot Scan from %T", src)
	}
}
