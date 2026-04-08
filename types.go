package dateparse

import "time"

// delta accumulates offsets with separate calendar and sub-day fields.
// Calendar fields (years, months, days) are applied via time.AddDate.
// Sub-day fields (hours, minutes, seconds, nanos) are applied via time.Add.
type delta struct {
	years, months, days     int
	hours, minutes, seconds int
	nanos                   int
}

// pendingOp holds a delta and direction that is waiting for an anchor.
type pendingOp struct {
	d   delta
	dir int // -1 or +1
}

// state is the scanner's accumulation register.
type state struct {
	delta        delta
	anchor       time.Time
	anchorSet    bool
	tod          timeOfDay
	todSet       bool
	direction    int         // -1 (before), +1 (after), 0 = no pending direction
	thisModifier bool        // true when "this" keyword was seen
	pendingOps   []pendingOp // stack of outer direction ops for chaining
}

// deltaField identifies which field of a delta a unit maps to.
type deltaField int

const (
	fieldYears deltaField = iota
	fieldMonths
	fieldDays
	fieldHours
	fieldMinutes
	fieldSeconds
	fieldNanos
)

// unitEntry maps a unit keyword to a delta field and scale factor.
type unitEntry struct {
	field deltaField
	scale int
}

// timeOfDay holds a parsed time.
type timeOfDay struct {
	hour       int // 0-23
	minute     int // 0-59
	second     int // 0-59
	nanosecond int
	tzOffset   *int // seconds east of UTC, nil = no explicit tz
}
