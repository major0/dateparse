package dateparse

import (
	"testing"
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

func TestMatchComment(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantTyp itemType
		wantVal string
		wantErr bool
	}{
		{
			name:    "simple comment",
			input:   "(hello)",
			wantTyp: itemComment,
			wantVal: "hello",
		},
		{
			name:    "nested comment",
			input:   "(hello (world))",
			wantTyp: itemComment,
			wantVal: "hello (world)",
		},
		{
			name:    "empty comment",
			input:   "()",
			wantTyp: itemComment,
			wantVal: "",
		},
		{
			name:    "unmatched open paren",
			input:   "(hello",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0}
			items, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != 1 {
				t.Fatalf("got %d items, want 1", len(items))
			}
			if items[0].typ != tt.wantTyp {
				t.Errorf("typ = %d, want %d", items[0].typ, tt.wantTyp)
			}
			if got := items[0].value.(string); got != tt.wantVal {
				t.Errorf("value = %q, want %q", got, tt.wantVal)
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
		{
			name:    "positive",
			input:   "@1705276800",
			wantSec: 1705276800,
			wantNs:  0,
		},
		{
			name:    "negative",
			input:   "@-86400",
			wantSec: -86400,
			wantNs:  0,
		},
		{
			name:    "zero",
			input:   "@0",
			wantSec: 0,
			wantNs:  0,
		},
		{
			name:    "fractional with dot",
			input:   "@1078100502.5",
			wantSec: 1078100502,
			wantNs:  500000000,
		},
		{
			name:    "fractional with comma",
			input:   "@1078100502,5",
			wantSec: 1078100502,
			wantNs:  500000000,
		},
		{
			name:    "just @ with no digits",
			input:   "@",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0}
			items, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != 1 {
				t.Fatalf("got %d items, want 1", len(items))
			}
			if items[0].typ != itemEpoch {
				t.Errorf("typ = %d, want %d (itemEpoch)", items[0].typ, itemEpoch)
			}
			ep := items[0].value.(epochSeconds)
			if ep.seconds != tt.wantSec {
				t.Errorf("seconds = %d, want %d", ep.seconds, tt.wantSec)
			}
			if ep.nanosecond != tt.wantNs {
				t.Errorf("nanosecond = %d, want %d", ep.nanosecond, tt.wantNs)
			}
		})
	}
}

func TestMatchRFC3339(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantDate calendarDate
		wantTime timeOfDay
		wantErr  bool
	}{
		{
			name:     "with T and Z",
			input:    "2024-01-15t00:00:00z",
			wantDate: calendarDate{2024, 1, 15},
			wantTime: timeOfDay{0, 0, 0, 0, intPtr(0)},
		},
		{
			name:     "with T and offset",
			input:    "2024-01-15t20:02:00-05:00",
			wantDate: calendarDate{2024, 1, 15},
			wantTime: timeOfDay{20, 2, 0, 0, intPtr(-18000)},
		},
		{
			name:     "with space separator",
			input:    "2024-01-15 20:02:00-05:00",
			wantDate: calendarDate{2024, 1, 15},
			wantTime: timeOfDay{20, 2, 0, 0, intPtr(-18000)},
		},
		{
			name:     "with fractional seconds dot",
			input:    "2024-01-15t14:30:00.123z",
			wantDate: calendarDate{2024, 1, 15},
			wantTime: timeOfDay{14, 30, 0, 123000000, intPtr(0)},
		},
		{
			name:     "with fractional seconds comma",
			input:    "2024-01-15t14:30:00,456z",
			wantDate: calendarDate{2024, 1, 15},
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
			sc := &scanner{input: tt.input, pos: 0}
			items, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != 2 {
				t.Fatalf("got %d items, want 2 (calendarDate + timeOfDay)", len(items))
			}

			// Check calendar date item.
			if items[0].typ != itemCalendarDate {
				t.Errorf("items[0].typ = %d, want %d (itemCalendarDate)", items[0].typ, itemCalendarDate)
			}
			gotDate := items[0].value.(calendarDate)
			if gotDate != tt.wantDate {
				t.Errorf("date = %+v, want %+v", gotDate, tt.wantDate)
			}

			// Check time-of-day item.
			if items[1].typ != itemTimeOfDay {
				t.Errorf("items[1].typ = %d, want %d (itemTimeOfDay)", items[1].typ, itemTimeOfDay)
			}
			gotTime := items[1].value.(timeOfDay)
			assertTimeOfDay(t, tt.name, gotTime, tt.wantTime)
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
		{
			name:     "24-hour HH:MM",
			input:    "14:30",
			wantTime: timeOfDay{14, 30, 0, 0, nil},
		},
		{
			name:     "24-hour HH:MM:SS",
			input:    "14:30:45",
			wantTime: timeOfDay{14, 30, 45, 0, nil},
		},
		{
			name:     "24-hour with fraction",
			input:    "14:30:45.123",
			wantTime: timeOfDay{14, 30, 45, 123000000, nil},
		},
		{
			name:     "12-hour pm",
			input:    "3pm",
			wantTime: timeOfDay{15, 0, 0, 0, nil},
		},
		{
			name:     "12-hour am",
			input:    "3am",
			wantTime: timeOfDay{3, 0, 0, 0, nil},
		},
		{
			name:     "12-hour with minutes",
			input:    "3:30pm",
			wantTime: timeOfDay{15, 30, 0, 0, nil},
		},
		{
			name:     "dotted a.m.",
			input:    "3a.m.",
			wantTime: timeOfDay{3, 0, 0, 0, nil},
		},
		{
			name:     "dotted p.m.",
			input:    "3p.m.",
			wantTime: timeOfDay{15, 0, 0, 0, nil},
		},
		{
			name:     "12am is midnight",
			input:    "12am",
			wantTime: timeOfDay{0, 0, 0, 0, nil},
		},
		{
			name:     "12pm is noon",
			input:    "12pm",
			wantTime: timeOfDay{12, 0, 0, 0, nil},
		},
		{
			name:     "24-hour with tz offset",
			input:    "14:30-0400",
			wantTime: timeOfDay{14, 30, 0, 0, intPtr(-14400)},
		},
		{
			name:     "24-hour with tz colon",
			input:    "14:30+05:30",
			wantTime: timeOfDay{14, 30, 0, 0, intPtr(19800)},
		},
		{
			name:    "am/pm with tz is error",
			input:   "3pm-0400",
			wantErr: true,
		},
		{
			name:    "invalid hour",
			input:   "25:00",
			wantErr: true,
		},
		{
			name:    "invalid minute",
			input:   "14:61",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := &scanner{input: tt.input, pos: 0}
			items, err := sc.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(items) != 1 {
				t.Fatalf("got %d items, want 1", len(items))
			}
			if items[0].typ != itemTimeOfDay {
				t.Errorf("typ = %d, want %d (itemTimeOfDay)", items[0].typ, itemTimeOfDay)
			}
			gotTime := items[0].value.(timeOfDay)
			assertTimeOfDay(t, tt.name, gotTime, tt.wantTime)
		})
	}
}
