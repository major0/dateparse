// Package dateparse provides GNU date --date compatible timestamp parsing.
//
// The parser uses a scanner/accumulator/resolver architecture:
// a greedy longest-match scanner consumes items left-to-right,
// an accumulator collects them with conflict detection,
// and a resolver combines accumulated items with a reference time
// to produce the final time.Time.
package dateparse

// TODO: Remove these compile-time assertions once scanner/accumulator/resolver
// consume all types. They exist only to suppress unused warnings during
// incremental development.
var (
	_ itemType     = itemComment
	_ item         = item{}
	_ calendarDate = calendarDate{}
	_ timeOfDay    = timeOfDay{}
	_ timeZone     = timeZone{}
	_ relativeUnit = unitMoment
	_ relativeItem = relativeItem{}
	_ namedRefType = namedTomorrow
	_ dayOfWeekRef = dayOfWeekRef{}
	_ epochSeconds = epochSeconds{}
)
