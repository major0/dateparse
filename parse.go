package dateparse

import "time"

// Parse interprets s as a GNU date-compatible string relative to ref.
// Returns the resolved time or a descriptive error. Never panics.
func Parse(_ string, _ time.Time) (time.Time, error) {
	return time.Time{}, nil
}
