package dateparse

import "time"

// itemType classifies a parsed token.
type itemType int

const (
	itemCalendarDate itemType = iota
	itemTimeOfDay
	itemTimeZone
	itemRelative
	itemNamedRef
	itemDayOfWeek
	itemEpoch
	itemPureNumber
	itemComment
)

// item represents a single parsed token from the input.
type item struct {
	typ   itemType
	value interface{} // type-specific payload
	pos   int         // byte offset in original input
}

// calendarDate holds a parsed date.
type calendarDate struct {
	year  int // 4-digit
	month int // 1-12
	day   int // 1-31
}

// timeOfDay holds a parsed time.
type timeOfDay struct {
	hour       int // 0-23
	minute     int // 0-59
	second     int // 0-59
	nanosecond int
	tzOffset   *int // seconds east of UTC, nil = no explicit tz
}

// timeZone holds a parsed timezone offset.
type timeZone struct {
	offsetSeconds int
}

// relativeUnit identifies a time unit for relative offsets.
type relativeUnit int

const (
	unitSecond relativeUnit = iota
	unitMinute
	unitHour
	unitDay
	unitWeek
	unitFortnight
	unitMonth
	unitYear
	unitGhurry
	unitScruple
	unitMileway
	unitMicrofortnight
	unitNundine
	unitDecade
	unitCentury
	unitMillennium
	unitOlympiad
	unitIndiction
	unitLustre
	unitJiffy
	unitMoment
)

// relativeItem holds a signed multiplier and unit for relative offsets.
type relativeItem struct {
	n    int // signed multiplier
	unit relativeUnit
}

// namedRefType identifies a named reference keyword.
type namedRefType int

const (
	namedNow       namedRefType = iota // zero displacement
	namedToday                         // zero displacement
	namedYesterday                     // -1 day
	namedTomorrow                      // +1 day
)

// dayOfWeekRef holds a day-of-week reference with ordinal modifier.
type dayOfWeekRef struct {
	day     time.Weekday
	ordinal int // 0=this, -1=last, 1=next/first, 2-12=Nth
}

// epochSeconds holds a parsed Unix epoch timestamp.
type epochSeconds struct {
	seconds    int64
	nanosecond int
}
