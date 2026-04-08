package dateparse

import (
	"testing"
	"time"
)

func TestUnitKeywords(t *testing.T) {
	// Verify total entry count.
	if got := len(unitKeywords); got != 58 {
		t.Fatalf("unitKeywords has %d entries, want 58", got)
	}

	// Representative sample: singular + at least one alias per unit.
	tests := []struct {
		key  string
		want relativeUnit
	}{
		{"second", unitSecond}, {"secs", unitSecond},
		{"minute", unitMinute}, {"mins", unitMinute},
		{"hour", unitHour}, {"hrs", unitHour},
		{"day", unitDay}, {"days", unitDay},
		{"week", unitWeek}, {"wks", unitWeek},
		{"fortnight", unitFortnight}, {"fortnights", unitFortnight},
		{"month", unitMonth}, {"mos", unitMonth}, {"mon", unitMonth},
		{"year", unitYear}, {"yrs", unitYear},
		{"ghurry", unitGhurry}, {"ghurries", unitGhurry},
		{"scruple", unitScruple}, {"scruples", unitScruple},
		{"mileway", unitMileway}, {"mileways", unitMileway},
		{"microfortnight", unitMicrofortnight}, {"microfortnights", unitMicrofortnight},
		{"nundine", unitNundine}, {"nundines", unitNundine},
		{"decade", unitDecade}, {"decades", unitDecade},
		{"century", unitCentury}, {"centuries", unitCentury},
		{"millennium", unitMillennium}, {"millennia", unitMillennium}, {"millenniums", unitMillennium},
		{"olympiad", unitOlympiad}, {"olympiads", unitOlympiad},
		{"indiction", unitIndiction}, {"indictions", unitIndiction},
		{"lustre", unitLustre}, {"lustres", unitLustre}, {"lustrum", unitLustre}, {"lustra", unitLustre},
		{"jiffy", unitJiffy}, {"jiffies", unitJiffy},
		{"moment", unitMoment}, {"moments", unitMoment},
	}
	for _, tt := range tests {
		got, ok := unitKeywords[tt.key]
		if !ok {
			t.Errorf("unitKeywords[%q]: missing", tt.key)
		} else if got != tt.want {
			t.Errorf("unitKeywords[%q] = %d, want %d", tt.key, got, tt.want)
		}
	}
}

func TestUnitDuration(t *testing.T) {
	if got := len(unitDuration); got != 13 {
		t.Fatalf("unitDuration has %d entries, want 13", got)
	}

	tests := []struct {
		unit relativeUnit
		want time.Duration
	}{
		{unitSecond, time.Second},
		{unitMinute, time.Minute},
		{unitHour, time.Hour},
		{unitDay, 24 * time.Hour},
		{unitWeek, 7 * 24 * time.Hour},
		{unitFortnight, 14 * 24 * time.Hour},
		{unitGhurry, 24 * time.Minute},
		{unitScruple, 60 * time.Second},
		{unitMileway, 20 * time.Minute},
		{unitMicrofortnight, 1209600 * time.Microsecond},
		{unitNundine, 8 * 24 * time.Hour},
		{unitJiffy, 10 * time.Millisecond},
		{unitMoment, 90 * time.Second},
	}
	for _, tt := range tests {
		got, ok := unitDuration[tt.unit]
		if !ok {
			t.Errorf("unitDuration[%d]: missing", tt.unit)
		} else if got != tt.want {
			t.Errorf("unitDuration[%d] = %v, want %v", tt.unit, got, tt.want)
		}
	}
}

func TestCalendarUnits(t *testing.T) {
	if got := len(calendarUnits); got != 8 {
		t.Fatalf("calendarUnits has %d entries, want 8", got)
	}

	tests := []struct {
		unit       relativeUnit
		wantYears  int
		wantMonths int
	}{
		{unitMonth, 0, 1},
		{unitYear, 1, 0},
		{unitDecade, 10, 0},
		{unitCentury, 100, 0},
		{unitMillennium, 1000, 0},
		{unitOlympiad, 4, 0},
		{unitIndiction, 15, 0},
		{unitLustre, 5, 0},
	}
	for _, tt := range tests {
		got, ok := calendarUnits[tt.unit]
		if !ok {
			t.Errorf("calendarUnits[%d]: missing", tt.unit)
		} else if got[0] != tt.wantYears || got[1] != tt.wantMonths {
			t.Errorf("calendarUnits[%d] = {%d, %d}, want {%d, %d}",
				tt.unit, got[0], got[1], tt.wantYears, tt.wantMonths)
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
