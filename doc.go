// Package dateparse provides GNU date --date compatible timestamp parsing.
//
// The parser uses a scanner/accumulator/resolver architecture:
// a greedy longest-match scanner consumes items left-to-right,
// an accumulator collects them with conflict detection,
// and a resolver combines accumulated items with a reference time
// to produce the final time.Time.
package dateparse

// Compile-time assertions to suppress unused warnings for types
// that are defined now but consumed by scanner/accumulator/resolver
// in later implementation stages.
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
