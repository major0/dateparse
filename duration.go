package dateparse

import (
	"fmt"
	"time"
)

// Duration holds a multi-field offset with calendar and sub-day components.
type Duration struct {
	Years, Months, Days     int
	Hours, Minutes, Seconds int
	Nanos                   int
}

// Apply applies the duration to t: AddDate first, then Add.
func (d Duration) Apply(t time.Time) time.Time {
	t = t.AddDate(d.Years, d.Months, d.Days)
	dur := time.Duration(d.Hours)*time.Hour +
		time.Duration(d.Minutes)*time.Minute +
		time.Duration(d.Seconds)*time.Second +
		time.Duration(d.Nanos)
	return t.Add(dur)
}

// ParseDuration interprets s as a relative duration expression.
// Returns the accumulated delta. Errors if an anchor token is present.
func ParseDuration(s string) (Duration, error) {
	input := asciiLower(s)
	sc := &scanner{
		input:             input,
		pos:               0,
		ref:               time.Time{},
		parseDurationMode: true,
	}
	st, err := sc.scan()
	if err != nil {
		return Duration{}, err
	}
	if st.anchorSet {
		return Duration{}, fmt.Errorf("ParseDuration: expression contains an anchor (use Parse instead)")
	}
	return Duration{
		Years:   st.delta.years,
		Months:  st.delta.months,
		Days:    st.delta.days,
		Hours:   st.delta.hours,
		Minutes: st.delta.minutes,
		Seconds: st.delta.seconds,
		Nanos:   st.delta.nanos,
	}, nil
}
