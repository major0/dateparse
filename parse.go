package dateparse

import (
	"fmt"
	"time"
)

// Parse interprets s as a GNU date-compatible string relative to ref.
// Returns the resolved time or a descriptive error. Never panics.
func Parse(s string, ref time.Time) (t time.Time, err error) {
	defer func() {
		if r := recover(); r != nil {
			t = time.Time{}
			err = fmt.Errorf("internal error: %v", r)
		}
	}()

	input := asciiLower(s)
	sc := &scanner{input: input, pos: 0, ref: ref}
	st, scanErr := sc.scan()
	if scanErr != nil {
		return time.Time{}, scanErr
	}
	return finalize(st, ref)
}
