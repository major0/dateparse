// Package dateparse provides GNU date --date compatible timestamp parsing.
//
// The parser uses a single-pass scan+accumulate architecture:
// a greedy longest-match scanner reads left-to-right, matches one token
// at a time, and updates state inline. No intermediate token list.
// Units accumulate into a multi-field delta with separate calendar fields
// applied via time.AddDate and sub-day fields applied via time.Add.
package dateparse

// Compile-time assertions to suppress unused warnings during incremental
// development. Remove once all types are consumed by production code.
var (
	_ calendarDate = calendarDate{}
	_ dayOfWeekRef = dayOfWeekRef{}
	_ epochSeconds = epochSeconds{}
)
