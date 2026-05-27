package dto

import (
	"fmt"
	"strings"
	"time"
)

// FlexTime wraps time.Time and accepts several ISO 8601 variants in JSON in addition to RFC3339.
// Useful for request DTOs when clients omit seconds or timezone.
type FlexTime struct {
	time.Time
}

var flexTimeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04Z07:00",
	"2006-01-02T15:04Z",
	"2006-01-02T15:04",
}

func (ft *FlexTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	if s == "null" {
		return nil
	}
	for _, layout := range flexTimeFormats {
		t, err := time.Parse(layout, s)
		if err == nil {
			ft.Time = t
			return nil
		}
	}
	return fmt.Errorf("cannot parse %q as time", s)
}

// ToTimePtr returns nil when the receiver is nil, otherwise a pointer to the wrapped time.Time.
func (ft *FlexTime) ToTimePtr() *time.Time {
	if ft == nil {
		return nil
	}
	t := ft.Time
	return &t
}

// NewFlexTime wraps t in a FlexTime pointer.
func NewFlexTime(t time.Time) *FlexTime {
	return &FlexTime{Time: t}
}
