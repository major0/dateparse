package dateparse

import (
	"testing"
	"time"
)

// parseRef is a fixed reference time used across Parse tests.
var parseRef = time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)

// ---------------------------------------------------------------------------
// Task 6.4 — Parse end-to-end tests
// ---------------------------------------------------------------------------

func TestParse_RFC3339(t *testing.T) {
	got, err := Parse("2024-01-15T14:30:00Z", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_RFC3339_WithOffset(t *testing.T) {
	got, err := Parse("2024-01-15T20:02:00-05:00", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2024, 1, 15, 20, 2, 0, 0, time.FixedZone("", -5*3600))
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_Epoch(t *testing.T) {
	got, err := Parse("@1705276800", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Unix(1705276800, 0)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_CalendarDateWithTime(t *testing.T) {
	got, err := Parse("20 Jul 2020 14:30", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2020, 7, 20, 14, 30, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_NamedRef_Now(t *testing.T) {
	got, err := Parse("now", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Equal(parseRef) {
		t.Errorf("got %v, want %v", got, parseRef)
	}
}

func TestParse_NamedRef_Yesterday(t *testing.T) {
	got, err := Parse("yesterday", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := parseRef.AddDate(0, 0, -1)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_NamedRef_Tomorrow(t *testing.T) {
	got, err := Parse("tomorrow", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := parseRef.AddDate(0, 0, 1)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_DayOfWeek_LastMonday(t *testing.T) {
	got, err := Parse("last monday", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// parseRef is 2024-07-15 (Monday). "last monday" = most recent Monday before ref.
	// Since ref IS a Monday, "last monday" should be 7 days before = 2024-07-08.
	want := time.Date(2024, 7, 8, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_Relative_3DaysAgo(t *testing.T) {
	got, err := Parse("3 days ago", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := parseRef.AddDate(0, 0, -3)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_Direction_3DaysBefore(t *testing.T) {
	got, err := Parse("3 days before 2024-01-15", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_Chained(t *testing.T) {
	// "7 hours before 2 weeks after 2024-07-13"
	// Step 1: parse "2024-07-13" → anchor = Jul 13 2024 midnight UTC
	// Step 2: "2 weeks after" → anchor = Jul 13 + 14 days = Jul 27 2024
	// Step 3: "7 hours before" → anchor = Jul 27 - 7 hours = Jul 26 17:00
	got, err := Parse("7 hours before 2 weeks after 2024-07-13", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2024, 7, 26, 17, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_TripleChain(t *testing.T) {
	// "1 day before 2 days after 3 days before 2024-01-15"
	// = Jan 15 - 3d + 2d - 1d = Jan 13
	got, err := Parse("1 day before 2 days after 3 days before 2024-01-15", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2024, 1, 13, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_Composite_YesterdayAt3PM(t *testing.T) {
	got, err := Parse("yesterday at 3pm", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2024, 7, 14, 15, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_EmptyInput(t *testing.T) {
	got, err := Parse("", parseRef)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParse_ErrorInvalidDate(t *testing.T) {
	_, err := Parse("2024-13-01", parseRef)
	if err == nil {
		t.Fatal("expected error for invalid month, got nil")
	}
}

func TestParse_ErrorUnrecognized(t *testing.T) {
	_, err := Parse("gobbledygook", parseRef)
	if err == nil {
		t.Fatal("expected error for unrecognized input, got nil")
	}
}

func TestParse_ErrorReturnsZeroTime(t *testing.T) {
	got, err := Parse("2024-13-01", parseRef)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !got.IsZero() {
		t.Errorf("expected zero time on error, got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Task 6.5 — ParseDuration tests
// ---------------------------------------------------------------------------

func TestParseDuration_3Days(t *testing.T) {
	got, err := ParseDuration("3 days")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Duration{Days: 3}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseDuration_1Year2Months(t *testing.T) {
	got, err := ParseDuration("1 year 2 months")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Duration{Years: 1, Months: 2}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseDuration_5HoursAgo(t *testing.T) {
	got, err := ParseDuration("5 hours ago")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Duration{Hours: -5}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseDuration_2WeeksAnd3Days(t *testing.T) {
	got, err := ParseDuration("2 weeks and 3 days")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Duration{Days: 17} // 2*7 + 3
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseDuration_Fortnight(t *testing.T) {
	got, err := ParseDuration("fortnight")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := Duration{Days: 14}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}

func TestParseDuration_AnchorPresent_Error(t *testing.T) {
	_, err := ParseDuration("3 days before Jan 15")
	if err == nil {
		t.Fatal("expected error for anchor in duration expression, got nil")
	}
}

func TestParseDuration_NamedRef_Error(t *testing.T) {
	_, err := ParseDuration("yesterday")
	if err == nil {
		t.Fatal("expected error for named ref in duration expression, got nil")
	}
}

func TestParseDuration_ApplyConsistency(t *testing.T) {
	dur, err := ParseDuration("3 days")
	if err != nil {
		t.Fatalf("ParseDuration error: %v", err)
	}
	fromDuration := dur.Apply(parseRef)

	fromParse, err := Parse("3 days", parseRef)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if !fromDuration.Equal(fromParse) {
		t.Errorf("ParseDuration.Apply = %v, Parse = %v", fromDuration, fromParse)
	}
}
