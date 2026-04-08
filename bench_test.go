package dateparse

import (
	"testing"
	"time"
)

var benchRef = time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)

// Sink prevents compiler from optimizing away results.
var benchResult time.Time
var benchDur Duration
var benchErr error //nolint:errname // benchmark sink, not a sentinel error

func BenchmarkParse_RFC3339(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("2024-01-15T14:30:00Z", benchRef)
	}
}

func BenchmarkParse_Epoch(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("@1705276800", benchRef)
	}
}

func BenchmarkParse_CalendarDate_ISO(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("2024-01-15", benchRef)
	}
}

func BenchmarkParse_CalendarDate_Literal(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("Jan 15, 2024", benchRef)
	}
}

func BenchmarkParse_NamedRef(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("yesterday", benchRef)
	}
}

func BenchmarkParse_Relative_Simple(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("3 days ago", benchRef)
	}
}

func BenchmarkParse_Relative_Multi(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("1 year 2 months 3 days ago", benchRef)
	}
}

func BenchmarkParse_Direction_Before(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("3 days before 2024-01-15", benchRef)
	}
}

func BenchmarkParse_Direction_Chained(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("7 hours before 2 weeks after 2024-07-13", benchRef)
	}
}

func BenchmarkParse_Composite(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("yesterday at 3pm", benchRef)
	}
}

func BenchmarkParse_DayOfWeek(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("last monday", benchRef)
	}
}

func BenchmarkParse_Empty(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("", benchRef)
	}
}

func BenchmarkParseDuration_Simple(b *testing.B) {
	for b.Loop() {
		benchDur, benchErr = ParseDuration("3 days")
	}
}

func BenchmarkParseDuration_Multi(b *testing.B) {
	for b.Loop() {
		benchDur, benchErr = ParseDuration("1 year 2 months and 3 days")
	}
}

func BenchmarkScan_LongInput(b *testing.B) {
	for b.Loop() {
		benchResult, benchErr = Parse("3 hours and 15 minutes before 2 weeks after Jan 15, 2024 3pm", benchRef)
	}
}
