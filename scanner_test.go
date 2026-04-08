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
	if st.anchor != nil {
		t.Error("anchor should be nil for empty input")
	}
	if st.timeOfDay != nil {
		t.Error("timeOfDay should be nil for empty input")
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
	if st.anchor != nil {
		t.Error("anchor should be nil for whitespace-only input")
	}
	if st.timeOfDay != nil {
		t.Error("timeOfDay should be nil for whitespace-only input")
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
			if st.anchor != nil {
				t.Error("anchor should be nil after comment")
			}
			if st.timeOfDay != nil {
				t.Error("timeOfDay should be nil after comment")
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
			if st.anchor == nil {
				t.Fatal("anchor should be set after epoch")
			}
			expected := time.Unix(tt.wantSec, int64(tt.wantNs))
			if !st.anchor.Equal(expected) {
				t.Errorf("anchor = %v, want %v", *st.anchor, expected)
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
			if st.anchor == nil {
				t.Fatal("anchor should be set after RFC 3339")
			}
			expectedAnchor := time.Date(tt.wantYear, time.Month(tt.wantMonth), tt.wantDay, 0, 0, 0, 0, time.UTC)
			if !st.anchor.Equal(expectedAnchor) {
				t.Errorf("anchor = %v, want %v", *st.anchor, expectedAnchor)
			}

			// Check time-of-day is set.
			if st.timeOfDay == nil {
				t.Fatal("timeOfDay should be set after RFC 3339")
			}
			assertTimeOfDay(t, tt.name, *st.timeOfDay, tt.wantTime)
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
			if st.timeOfDay == nil {
				t.Fatal("timeOfDay should be set")
			}
			assertTimeOfDay(t, tt.name, *st.timeOfDay, tt.wantTime)
		})
	}
}
