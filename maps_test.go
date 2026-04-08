package dateparse

import (
	"testing"
	"time"
)

func TestUnitTable(t *testing.T) {
	tests := []struct {
		key       string
		wantField deltaField
		wantScale int
	}{
		// millennium
		{"millennium", fieldYears, 1000}, {"millennia", fieldYears, 1000}, {"millenniums", fieldYears, 1000},
		// saeculum
		{"saeculum", fieldYears, 100}, {"saecula", fieldYears, 100},
		// century
		{"century", fieldYears, 100}, {"centuries", fieldYears, 100},
		// indiction
		{"indiction", fieldYears, 15}, {"indictions", fieldYears, 15},
		// decade
		{"decade", fieldYears, 10}, {"decades", fieldYears, 10},
		// lustre
		{"lustre", fieldYears, 5}, {"lustres", fieldYears, 5}, {"lustrum", fieldYears, 5}, {"lustra", fieldYears, 5},
		// olympiad / quadrennium
		{"olympiad", fieldYears, 4}, {"olympiads", fieldYears, 4},
		{"quadrennium", fieldYears, 4}, {"quadrennia", fieldYears, 4},
		// biennial
		{"biennial", fieldYears, 2}, {"biennials", fieldYears, 2},
		// annus
		{"annus", fieldYears, 1}, {"anni", fieldYears, 1},
		// year
		{"year", fieldYears, 1}, {"years", fieldYears, 1}, {"yr", fieldYears, 1}, {"yrs", fieldYears, 1},
		// semester / friedman
		{"semester", fieldMonths, 6}, {"semesters", fieldMonths, 6},
		{"friedman", fieldMonths, 6}, {"friedmans", fieldMonths, 6},
		// season
		{"season", fieldMonths, 3}, {"seasons", fieldMonths, 3},
		// trimester
		{"trimester", fieldMonths, 3}, {"trimesters", fieldMonths, 3},
		// month
		{"month", fieldMonths, 1}, {"months", fieldMonths, 1}, {"mo", fieldMonths, 1}, {"mos", fieldMonths, 1}, {"mon", fieldMonths, 1},
		// fortnight
		{"fortnight", fieldDays, 14}, {"fortnights", fieldDays, 14},
		// nundine
		{"nundine", fieldDays, 8}, {"nundines", fieldDays, 8},
		// week
		{"week", fieldDays, 7}, {"weeks", fieldDays, 7}, {"wk", fieldDays, 7}, {"wks", fieldDays, 7},
		// nychthemeron
		{"nychthemeron", fieldDays, 1}, {"nychthemera", fieldDays, 1},
		// day
		{"day", fieldDays, 1}, {"days", fieldDays, 1},
		// pahar
		{"pahar", fieldHours, 3}, {"pahars", fieldHours, 3},
		// hour
		{"hour", fieldHours, 1}, {"hours", fieldHours, 1}, {"hr", fieldHours, 1}, {"hrs", fieldHours, 1},
		// ghurry
		{"ghurry", fieldMinutes, 24}, {"ghurries", fieldMinutes, 24},
		// mileway
		{"mileway", fieldMinutes, 20}, {"mileways", fieldMinutes, 20},
		// minute
		{"minute", fieldMinutes, 1}, {"minutes", fieldMinutes, 1}, {"min", fieldMinutes, 1}, {"mins", fieldMinutes, 1},
		// moment
		{"moment", fieldSeconds, 90}, {"moments", fieldSeconds, 90},
		// scruple
		{"scruple", fieldSeconds, 60}, {"scruples", fieldSeconds, 60},
		// second
		{"second", fieldSeconds, 1}, {"seconds", fieldSeconds, 1}, {"sec", fieldSeconds, 1}, {"secs", fieldSeconds, 1},
		// helek
		{"helek", fieldNanos, 3333333333}, {"halakim", fieldNanos, 3333333333}, {"helakim", fieldNanos, 3333333333},
		// microfortnight
		{"microfortnight", fieldNanos, 1209600000}, {"microfortnights", fieldNanos, 1209600000},
		// jiffy
		{"jiffy", fieldNanos, 10000000}, {"jiffies", fieldNanos, 10000000},
		// shake
		{"shake", fieldNanos, 10}, {"shakes", fieldNanos, 10},
	}

	for _, tt := range tests {
		got, ok := unitTable[tt.key]
		if !ok {
			t.Errorf("unitTable[%q]: missing", tt.key)
			continue
		}
		if got.field != tt.wantField {
			t.Errorf("unitTable[%q].field = %d, want %d", tt.key, got.field, tt.wantField)
		}
		if got.scale != tt.wantScale {
			t.Errorf("unitTable[%q].scale = %d, want %d", tt.key, got.scale, tt.wantScale)
		}
	}
}

func TestOrdinalModifiers(t *testing.T) {
	if got := len(ordinalModifiers); got != 15 {
		t.Fatalf("ordinalModifiers has %d entries, want 15", got)
	}

	tests := []struct {
		key  string
		want int
	}{
		{"last", -1},
		{"this", 0},
		{"next", 1},
		{"first", 1},
		{"second", 2},
		{"third", 3},
		{"fourth", 4},
		{"fifth", 5},
		{"sixth", 6},
		{"seventh", 7},
		{"eighth", 8},
		{"ninth", 9},
		{"tenth", 10},
		{"eleventh", 11},
		{"twelfth", 12},
	}
	for _, tt := range tests {
		got, ok := ordinalModifiers[tt.key]
		if !ok {
			t.Errorf("ordinalModifiers[%q]: missing", tt.key)
		} else if got != tt.want {
			t.Errorf("ordinalModifiers[%q] = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestMonthNames(t *testing.T) {
	if got := len(monthNames); got != 36 {
		t.Fatalf("monthNames has %d entries, want 36", got)
	}

	tests := []struct {
		key  string
		want int
	}{
		// Full names
		{"january", 1}, {"february", 2}, {"march", 3}, {"april", 4},
		{"may", 5}, {"june", 6}, {"july", 7}, {"august", 8},
		{"september", 9}, {"october", 10}, {"november", 11}, {"december", 12},
		// 3-letter abbreviations
		{"jan", 1}, {"feb", 2}, {"mar", 3}, {"apr", 4},
		{"jun", 6}, {"jul", 7}, {"aug", 8}, {"sep", 9},
		{"oct", 10}, {"nov", 11}, {"dec", 12},
		// Abbreviations with period
		{"jan.", 1}, {"feb.", 2}, {"mar.", 3}, {"apr.", 4},
		{"jun.", 6}, {"jul.", 7}, {"aug.", 8}, {"sep.", 9},
		{"oct.", 10}, {"nov.", 11}, {"dec.", 12},
		// Special "sept" forms
		{"sept", 9}, {"sept.", 9},
	}
	for _, tt := range tests {
		got, ok := monthNames[tt.key]
		if !ok {
			t.Errorf("monthNames[%q]: missing", tt.key)
		} else if got != tt.want {
			t.Errorf("monthNames[%q] = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestWeekdayNames(t *testing.T) {
	if got := len(weekdayNames); got != 29 {
		t.Fatalf("weekdayNames has %d entries, want 29", got)
	}

	tests := []struct {
		key  string
		want time.Weekday
	}{
		// Full names
		{"sunday", time.Sunday}, {"monday", time.Monday},
		{"tuesday", time.Tuesday}, {"wednesday", time.Wednesday},
		{"thursday", time.Thursday}, {"friday", time.Friday},
		{"saturday", time.Saturday},
		// 3-letter abbreviations
		{"sun", time.Sunday}, {"mon", time.Monday}, {"tue", time.Tuesday},
		{"wed", time.Wednesday}, {"thu", time.Thursday},
		{"fri", time.Friday}, {"sat", time.Saturday},
		// Abbreviations with period
		{"sun.", time.Sunday}, {"mon.", time.Monday}, {"tue.", time.Tuesday},
		{"wed.", time.Wednesday}, {"thu.", time.Thursday},
		{"fri.", time.Friday}, {"sat.", time.Saturday},
		// Special forms
		{"tues", time.Tuesday}, {"tues.", time.Tuesday},
		{"wednes", time.Wednesday}, {"wednes.", time.Wednesday},
		{"thur", time.Thursday}, {"thur.", time.Thursday},
		{"thurs", time.Thursday}, {"thurs.", time.Thursday},
	}
	for _, tt := range tests {
		got, ok := weekdayNames[tt.key]
		if !ok {
			t.Errorf("weekdayNames[%q]: missing", tt.key)
		} else if got != tt.want {
			t.Errorf("weekdayNames[%q] = %v, want %v", tt.key, got, tt.want)
		}
	}
}
