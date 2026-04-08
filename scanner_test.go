package dateparse

import (
	"testing"
	"time"
)

// intPtr is a helper to create *int values for tzOffset comparisons.
func intPtr(v int) *int { return &v }

// assertTimeOfDay compares a timeOfDay value field-by-field, handling the
// pointer-based tzOffset correctly.
func assertTimeOfDay(t *testing.T, label string, got, want timeOfDay) {
	t.Helper()
	if got.hour != want.hour {
		t.Errorf("%s: hour = %d, want %d", label, got.hour, want.hour)
	}
	if got.minute != want.minute {
		t.Errorf("%s: minute = %d, want %d", label, got.minute, want.minute)
	}
	if got.second != want.second {
		t.Errorf("%s: second = %d, want %d", label, got.second, want.second)
	}
	if got.nanosecond != want.nanosecond {
		t.Errorf("%s: nanosecond = %d, want %d", label, got.nanosecond, want.nanosecond)
	}
	switch {
	case got.tzOffset == nil && want.tzOffset == nil:
		// ok
	case got.tzOffset == nil && want.tzOffset != nil:
		t.Errorf("%s: tzOffset = nil, want %d", label, *want.tzOffset)
	case got.tzOffset != nil && want.tzOffset == nil:
		t.Errorf("%s: tzOffset = %d, want nil", label, *got.tzOffset)
	default:
		if *got.tzOffset != *want.tzOffset {
			t.Errorf("%s: tzOffset = %d, want %d", label, *got.tzOffset, *want.tzOffset)
		}
	}
}

// ref is a fixed reference time used across scanner tests.
var ref = time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)

func TestScanEmptyInput(t *testing.T) {
	sc := &scanner{input: "", pos: 0, ref: ref}
	st, err := sc.scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.anchorSet {
		t.Error("anchor should not be set for empty input")
	}
	if st.todSet {
		t.Error("timeOfDay should not be set for empty input")
	}
	if st.delta != (delta{}) {
		t.Errorf("delta should be zero, got %+v", st.delta)
	}
}

func TestScanWhitespaceOnly(t *testing.T) {
	sc := &scanner{input: "   \t\n  ", pos: 0, ref: ref}
	st, err := sc.scan()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.anchorSet {
		t.Error("anchor should not be set for whitespace-only input")
	}
	if st.todSet {
		t.Error("timeOfDay should not be set for whitespace-only input")
	}
	if st.delta != (delta{}) {
		t.Errorf("delta should be zero, got %+v", st.delta)
	}
}

func TestMatchComment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "simple comment", input: "(hello)"},
		{name: "nested comment", input: "(hello (world))"},
		{name: "empty comment", input: "()"},
		{name: "unmatched open paren", input: "(hello", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Comment should leave state unchanged.
			if st.anchorSet {
				t.Error("anchor should not be set after comment")
			}
			if st.todSet {
				t.Error("timeOfDay should not be set after comment")
			}
			if st.delta != (delta{}) {
				t.Errorf("delta should be zero, got %+v", st.delta)
			}
		})
	}
}

func TestMatchEpoch(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantSec int64
		wantNs  int
		wantErr bool
	}{
		{name: "positive", input: "@1705276800", wantSec: 1705276800, wantNs: 0},
		{name: "negative", input: "@-86400", wantSec: -86400, wantNs: 0},
		{name: "zero", input: "@0", wantSec: 0, wantNs: 0},
		{name: "fractional with dot", input: "@1078100502.5", wantSec: 1078100502, wantNs: 500000000},
		{name: "fractional with comma", input: "@1078100502,5", wantSec: 1078100502, wantNs: 500000000},
		{name: "just @ with no digits", input: "@", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !st.anchorSet {
				t.Fatal("anchor should be set after epoch")
			}
			expected := time.Unix(tt.wantSec, int64(tt.wantNs))
			if !st.anchor.Equal(expected) {
				t.Errorf("anchor = %v, want %v", st.anchor, expected)
			}
		})
	}
}

func TestMatchRFC3339(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantMonth int
		wantDay   int
		wantTime  timeOfDay
		wantErr   bool
	}{
		{
			name:     "with T and Z",
			input:    "2024-01-15t00:00:00z",
			wantYear: 2024, wantMonth: 1, wantDay: 15,
			wantTime: timeOfDay{0, 0, 0, 0, intPtr(0)},
		},
		{
			name:     "with T and offset",
			input:    "2024-01-15t20:02:00-05:00",
			wantYear: 2024, wantMonth: 1, wantDay: 15,
			wantTime: timeOfDay{20, 2, 0, 0, intPtr(-18000)},
		},
		{
			name:     "with space separator",
			input:    "2024-01-15 20:02:00-05:00",
			wantYear: 2024, wantMonth: 1, wantDay: 15,
			wantTime: timeOfDay{20, 2, 0, 0, intPtr(-18000)},
		},
		{
			name:     "with fractional seconds dot",
			input:    "2024-01-15t14:30:00.123z",
			wantYear: 2024, wantMonth: 1, wantDay: 15,
			wantTime: timeOfDay{14, 30, 0, 123000000, intPtr(0)},
		},
		{
			name:     "with fractional seconds comma",
			input:    "2024-01-15t14:30:00,456z",
			wantYear: 2024, wantMonth: 1, wantDay: 15,
			wantTime: timeOfDay{14, 30, 0, 456000000, intPtr(0)},
		},
		{
			name:    "invalid month",
			input:   "2024-13-01t00:00:00z",
			wantErr: true,
		},
		{
			name:    "invalid day",
			input:   "2024-02-30t00:00:00z",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check anchor is set to the correct date at midnight UTC.
			if !st.anchorSet {
				t.Fatal("anchor should be set after RFC 3339")
			}
			expectedAnchor := time.Date(tt.wantYear, time.Month(tt.wantMonth), tt.wantDay, 0, 0, 0, 0, time.UTC)
			if !st.anchor.Equal(expectedAnchor) {
				t.Errorf("anchor = %v, want %v", st.anchor, expectedAnchor)
			}

			// Check time-of-day is set.
			if !st.todSet {
				t.Fatal("timeOfDay should be set after RFC 3339")
			}
			assertTimeOfDay(t, tt.name, st.tod, tt.wantTime)
		})
	}
}

func TestMatchTimeOfDay(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTime timeOfDay
		wantErr  bool
	}{
		{name: "24-hour HH:MM", input: "14:30", wantTime: timeOfDay{14, 30, 0, 0, nil}},
		{name: "24-hour HH:MM:SS", input: "14:30:45", wantTime: timeOfDay{14, 30, 45, 0, nil}},
		{name: "24-hour with fraction", input: "14:30:45.123", wantTime: timeOfDay{14, 30, 45, 123000000, nil}},
		{name: "12-hour pm", input: "3pm", wantTime: timeOfDay{15, 0, 0, 0, nil}},
		{name: "12-hour am", input: "3am", wantTime: timeOfDay{3, 0, 0, 0, nil}},
		{name: "12-hour with minutes", input: "3:30pm", wantTime: timeOfDay{15, 30, 0, 0, nil}},
		{name: "dotted a.m.", input: "3a.m.", wantTime: timeOfDay{3, 0, 0, 0, nil}},
		{name: "dotted p.m.", input: "3p.m.", wantTime: timeOfDay{15, 0, 0, 0, nil}},
		{name: "12am is midnight", input: "12am", wantTime: timeOfDay{0, 0, 0, 0, nil}},
		{name: "12pm is noon", input: "12pm", wantTime: timeOfDay{12, 0, 0, 0, nil}},
		{name: "24-hour with tz offset", input: "14:30-0400", wantTime: timeOfDay{14, 30, 0, 0, intPtr(-14400)}},
		{name: "24-hour with tz colon", input: "14:30+05:30", wantTime: timeOfDay{14, 30, 0, 0, intPtr(19800)}},
		{name: "am/pm with tz is error", input: "3pm-0400", wantErr: true},
		{name: "invalid hour", input: "25:00", wantErr: true},
		{name: "invalid minute", input: "14:61", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !st.todSet {
				t.Fatal("timeOfDay should be set")
			}
			assertTimeOfDay(t, tt.name, st.tod, tt.wantTime)
		})
	}
}

// ---------------------------------------------------------------------------
// TestMatchCalendarDate
// ---------------------------------------------------------------------------

func TestMatchCalendarDate(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAnchor time.Time
		wantErr    bool
	}{
		{
			name:       "ISO 2024-01-15",
			input:      "2024-01-15",
			wantAnchor: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "ISO standalone not RFC 3339",
			input:      "2024-01-15",
			wantAnchor: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "two-digit year 99",
			input:      "99-01-15",
			wantAnchor: time.Date(1999, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "two-digit year 24",
			input:      "24-01-15",
			wantAnchor: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "US format M/D/YYYY",
			input:      "7/20/2020",
			wantAnchor: time.Date(2020, 7, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "US format M/D ref year",
			input:      "7/20",
			wantAnchor: time.Date(2024, 7, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "literal D Mon YYYY",
			input:      "20 jul 2020",
			wantAnchor: time.Date(2020, 7, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "literal Mon D YYYY",
			input:      "jul 20 2020",
			wantAnchor: time.Date(2020, 7, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "literal Mon D, YYYY comma",
			input:      "jul 20, 2020",
			wantAnchor: time.Date(2020, 7, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "literal D-Mon-YYYY",
			input:      "20-jul-2020",
			wantAnchor: time.Date(2020, 7, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "literal D Mon ref year",
			input:      "20 jul",
			wantAnchor: time.Date(2024, 7, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "invalid date 2024-02-30",
			input:   "2024-02-30",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !st.anchorSet {
				t.Fatal("anchor should be set")
			}
			if !st.anchor.Equal(tt.wantAnchor) {
				t.Errorf("anchor = %v, want %v", st.anchor, tt.wantAnchor)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestMatchTimezone
// ---------------------------------------------------------------------------

func TestMatchTimezone(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantTZOffset int
	}{
		{name: "utc", input: "utc", wantTZOffset: 0},
		{name: "z", input: "z", wantTZOffset: 0},
		{name: "u.t.c.", input: "u.t.c.", wantTZOffset: 0},
		{name: "utc+05:30", input: "utc+05:30", wantTZOffset: 19800},
		{name: "utc-04", input: "utc-04", wantTZOffset: -14400},
		{name: "utc dst", input: "utc dst", wantTZOffset: 3600},
		{name: "standalone +0530", input: "+0530", wantTZOffset: 19800},
		{name: "standalone -04:00", input: "-04:00", wantTZOffset: -14400},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !st.todSet {
				t.Fatal("timeOfDay should be set by timezone matcher")
			}
			if st.tod.tzOffset == nil {
				t.Fatal("tzOffset should be set")
			}
			if *st.tod.tzOffset != tt.wantTZOffset {
				t.Errorf("tzOffset = %d, want %d", *st.tod.tzOffset, tt.wantTZOffset)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestMatchNamedRef
// ---------------------------------------------------------------------------

func TestMatchNamedRef(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAnchor time.Time
	}{
		{name: "now", input: "now", wantAnchor: ref},
		{name: "today", input: "today", wantAnchor: ref},
		{name: "yesterday", input: "yesterday", wantAnchor: ref.AddDate(0, 0, -1)},
		{name: "tomorrow", input: "tomorrow", wantAnchor: ref.AddDate(0, 0, 1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !st.anchorSet {
				t.Fatal("anchor should be set")
			}
			if !st.anchor.Equal(tt.wantAnchor) {
				t.Errorf("anchor = %v, want %v", st.anchor, tt.wantAnchor)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestMatchDayOfWeek
// ---------------------------------------------------------------------------

func TestMatchDayOfWeek(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAnchor time.Time
	}{
		{
			name:       "monday bare",
			input:      "monday",
			wantAnchor: time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC), // next Monday
		},
		{
			name:       "last friday",
			input:      "last friday",
			wantAnchor: time.Date(2024, 6, 14, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "next monday",
			input:      "next monday",
			wantAnchor: time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "this saturday",
			input:      "this saturday",
			wantAnchor: time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "third tuesday",
			input:      "third tuesday",
			wantAnchor: time.Date(2024, 6, 18, 0, 0, 0, 0, time.UTC).AddDate(0, 0, 14), // 3rd Tue after ref
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !st.anchorSet {
				t.Fatal("anchor should be set")
			}
			if !st.anchor.Equal(tt.wantAnchor) {
				t.Errorf("anchor = %v, want %v", st.anchor, tt.wantAnchor)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestMatchRelative
// ---------------------------------------------------------------------------

func TestMatchRelative(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantDelta delta
	}{
		{name: "3 days", input: "3 days", wantDelta: delta{days: 3}},
		{name: "day implicit", input: "day", wantDelta: delta{days: 1}},
		{name: "+2 hours", input: "+2 hours", wantDelta: delta{hours: 2}},
		{name: "-1 week", input: "-1 week", wantDelta: delta{days: -7}},
		{name: "1 year 2 months", input: "1 year 2 months", wantDelta: delta{years: 1, months: 2}},
		{name: "fortnight", input: "fortnight", wantDelta: delta{days: 14}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if st.delta != tt.wantDelta {
				t.Errorf("delta = %+v, want %+v", st.delta, tt.wantDelta)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestMatchDirectionOp
// ---------------------------------------------------------------------------

func TestMatchDirectionOp(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAnchor time.Time
		wantErr    bool
	}{
		{
			name:       "3 days ago",
			input:      "3 days ago",
			wantAnchor: ref.AddDate(0, 0, -3),
		},
		{
			name:       "2 weeks hence",
			input:      "2 weeks hence",
			wantAnchor: ref.AddDate(0, 0, 14),
		},
		{
			name:       "3 days before 2024-01-15",
			input:      "3 days before 2024-01-15",
			wantAnchor: time.Date(2024, 1, 12, 0, 0, 0, 0, time.UTC),
		},
		{
			name:       "2 weeks after 2024-01-15",
			input:      "2 weeks after 2024-01-15",
			wantAnchor: time.Date(2024, 1, 29, 0, 0, 0, 0, time.UTC),
		},
		{
			name:    "before without delta",
			input:   "before",
			wantErr: true,
		},
		{
			name:    "after without delta",
			input:   "after",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !st.anchorSet {
				t.Fatal("anchor should be set")
			}
			if !st.anchor.Equal(tt.wantAnchor) {
				t.Errorf("anchor = %v, want %v", st.anchor, tt.wantAnchor)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestMatchPureNumber
// ---------------------------------------------------------------------------

func TestMatchPureNumber(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantAnchor *time.Time
		wantTime   *timeOfDay
		wantErr    bool
	}{
		{
			name:       "8-digit date 20240115",
			input:      "20240115",
			wantAnchor: timePtr(time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)),
		},
		{
			name:     "4-digit time 1430",
			input:    "1430",
			wantTime: &timeOfDay{14, 30, 0, 0, nil},
		},
		{
			name:    "invalid month 20241301",
			input:   "20241301",
			wantErr: true,
		},
		{
			name:    "invalid hour 2561",
			input:   "2561",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantAnchor != nil {
				if !st.anchorSet {
					t.Fatal("anchor should be set")
				}
				if !st.anchor.Equal(*tt.wantAnchor) {
					t.Errorf("anchor = %v, want %v", st.anchor, *tt.wantAnchor)
				}
			}
			if tt.wantTime != nil {
				if !st.todSet {
					t.Fatal("timeOfDay should be set")
				}
				assertTimeOfDay(t, tt.name, st.tod, *tt.wantTime)
			}
		})
	}
}

// timePtr is a helper to create *time.Time values.
func timePtr(t time.Time) *time.Time { return &t }

// ---------------------------------------------------------------------------
// TestMatchNoise
// ---------------------------------------------------------------------------

func TestMatchNoise(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantDelta delta
	}{
		{
			name:      "and between relative items",
			input:     "2 weeks and 3 days",
			wantDelta: delta{days: 17},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			st, err := sc.scan()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if st.delta != tt.wantDelta {
				t.Errorf("delta = %+v, want %+v", st.delta, tt.wantDelta)
			}
		})
	}

	// Test stray hyphen as noise
	t.Run("stray hyphen", func(t *testing.T) {
		// "-" alone is noise (not followed by digit immediately)
		sc := &scanner{input: "-", pos: 0, ref: ref}
		st, err := sc.scan()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if st.delta != (delta{}) {
			t.Errorf("delta should be zero after stray hyphen, got %+v", st.delta)
		}
	})

	// Test "yesterday at 3pm" — anchor + time-of-day with noise "at"
	t.Run("yesterday at 3pm", func(t *testing.T) {
		sc := &scanner{input: "yesterday at 3pm", pos: 0, ref: ref}
		st, err := sc.scan()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		wantAnchor := ref.AddDate(0, 0, -1)
		if !st.anchorSet {
			t.Fatal("anchor should be set")
		}
		if !st.anchor.Equal(wantAnchor) {
			t.Errorf("anchor = %v, want %v", st.anchor, wantAnchor)
		}
		if !st.todSet {
			t.Fatal("timeOfDay should be set")
		}
		assertTimeOfDay(t, "yesterday at 3pm", st.tod, timeOfDay{15, 0, 0, 0, nil})
	})
}

// ---------------------------------------------------------------------------
// TestConflictDetection
// ---------------------------------------------------------------------------

func TestConflictDetection(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "two anchors", input: "2024-01-15 2024-02-20"},
		{name: "two time-of-day", input: "14:30 15:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0, ref: ref}
			_, err := sc.scan()
			if err == nil {
				t.Fatal("expected conflict error, got nil")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// TestScanError
// ---------------------------------------------------------------------------

func TestScanError(t *testing.T) {
	sc := &scanner{input: "xyzzy", pos: 0, ref: ref}
	_, err := sc.scan()
	if err == nil {
		t.Fatal("expected error for unrecognized input, got nil")
	}
}
