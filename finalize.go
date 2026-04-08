package dateparse

import "time"

// finalize resolves the final time from the accumulated state.
//
//nolint:unparam // error return will be used when additional validation is added
func finalize(st *state, ref time.Time) (time.Time, error) {
	// Empty state (no tokens matched) → midnight of ref date.
	if st.anchor == nil && isDeltaZero(st.delta) && st.timeOfDay == nil {
		return time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, ref.Location()), nil
	}

	// If no anchor set, use ref as anchor.
	anchor := ref
	if st.anchor != nil {
		anchor = *st.anchor
	}

	// Apply delta to anchor. Direction has already been applied during scanning,
	// so sign is always +1 here.
	anchor = applyDeltaToTime(anchor, st.delta, 1)

	// If timeOfDay is set, override time component on the anchor.
	if st.timeOfDay != nil {
		tod := st.timeOfDay
		year, month, day := anchor.Date()
		loc := anchor.Location()
		if tod.tzOffset != nil {
			loc = time.FixedZone("", *tod.tzOffset)
		}
		anchor = time.Date(year, month, day, tod.hour, tod.minute, tod.second, tod.nanosecond, loc)
	}

	return anchor, nil
}
