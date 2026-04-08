package dateparse

import "time"

// unitTable maps lowercase unit name forms to their {field, scale} pair.
// All aliases (singular, plural, abbreviated) in one map for O(1) lookup.
var unitTable = map[string]unitEntry{
	// years × 1000
	"millennium": {fieldYears, 1000}, "millennia": {fieldYears, 1000}, "millenniums": {fieldYears, 1000},
	// years × 100
	"saeculum": {fieldYears, 100}, "saecula": {fieldYears, 100},
	"century": {fieldYears, 100}, "centuries": {fieldYears, 100},
	// years × 15
	"indiction": {fieldYears, 15}, "indictions": {fieldYears, 15},
	// years × 10
	"decade": {fieldYears, 10}, "decades": {fieldYears, 10},
	// years × 5
	"lustre": {fieldYears, 5}, "lustres": {fieldYears, 5}, "lustrum": {fieldYears, 5}, "lustra": {fieldYears, 5},
	// years × 4
	"olympiad": {fieldYears, 4}, "olympiads": {fieldYears, 4}, "quadrennium": {fieldYears, 4}, "quadrennia": {fieldYears, 4},
	// years × 2
	"biennial": {fieldYears, 2}, "biennials": {fieldYears, 2},
	// years × 1
	"annus": {fieldYears, 1}, "anni": {fieldYears, 1},
	"year": {fieldYears, 1}, "years": {fieldYears, 1}, "yr": {fieldYears, 1}, "yrs": {fieldYears, 1},
	// months × 6
	"semester": {fieldMonths, 6}, "semesters": {fieldMonths, 6}, "friedman": {fieldMonths, 6}, "friedmans": {fieldMonths, 6},
	// months × 3
	"season": {fieldMonths, 3}, "seasons": {fieldMonths, 3},
	"trimester": {fieldMonths, 3}, "trimesters": {fieldMonths, 3},
	// months × 1
	"month": {fieldMonths, 1}, "months": {fieldMonths, 1}, "mo": {fieldMonths, 1}, "mos": {fieldMonths, 1}, "mon": {fieldMonths, 1},
	// days × 14
	"fortnight": {fieldDays, 14}, "fortnights": {fieldDays, 14},
	// days × 8
	"nundine": {fieldDays, 8}, "nundines": {fieldDays, 8},
	// days × 7
	"week": {fieldDays, 7}, "weeks": {fieldDays, 7}, "wk": {fieldDays, 7}, "wks": {fieldDays, 7},
	// days × 1
	"nychthemeron": {fieldDays, 1}, "nychthemera": {fieldDays, 1},
	"day": {fieldDays, 1}, "days": {fieldDays, 1},
	// hours × 3
	"pahar": {fieldHours, 3}, "pahars": {fieldHours, 3},
	// hours × 1
	"hour": {fieldHours, 1}, "hours": {fieldHours, 1}, "hr": {fieldHours, 1}, "hrs": {fieldHours, 1},
	// minutes × 24
	"ghurry": {fieldMinutes, 24}, "ghurries": {fieldMinutes, 24},
	// minutes × 20
	"mileway": {fieldMinutes, 20}, "mileways": {fieldMinutes, 20},
	// minutes × 1
	"minute": {fieldMinutes, 1}, "minutes": {fieldMinutes, 1}, "min": {fieldMinutes, 1}, "mins": {fieldMinutes, 1},
	// seconds × 90
	"moment": {fieldSeconds, 90}, "moments": {fieldSeconds, 90},
	// seconds × 60
	"scruple": {fieldSeconds, 60}, "scruples": {fieldSeconds, 60},
	// seconds × 1
	"second": {fieldSeconds, 1}, "seconds": {fieldSeconds, 1}, "sec": {fieldSeconds, 1}, "secs": {fieldSeconds, 1},
	// nanos × 3333333333
	"helek": {fieldNanos, 3333333333}, "halakim": {fieldNanos, 3333333333}, "helakim": {fieldNanos, 3333333333},
	// nanos × 1209600000
	"microfortnight": {fieldNanos, 1209600000}, "microfortnights": {fieldNanos, 1209600000},
	// nanos × 10000000
	"jiffy": {fieldNanos, 10000000}, "jiffies": {fieldNanos, 10000000},
	// nanos × 10
	"shake": {fieldNanos, 10}, "shakes": {fieldNanos, 10},
}

// ordinalModifiers maps ordinal modifier keywords to their integer values.
var ordinalModifiers = map[string]int{
	"last": -1, "this": 0, "next": 1, "first": 1,
	"second": 2, "third": 3, "fourth": 4, "fifth": 5, "sixth": 6,
	"seventh": 7, "eighth": 8, "ninth": 9, "tenth": 10, "eleventh": 11, "twelfth": 12,
}

// monthNames maps lowercase month name forms to month number (1-12).
// Includes full names, 3-letter abbreviations, abbreviations with trailing
// period, and the special "sept"/"sept." forms.
var monthNames = map[string]int{
	"january": 1, "jan": 1, "jan.": 1,
	"february": 2, "feb": 2, "feb.": 2,
	"march": 3, "mar": 3, "mar.": 3,
	"april": 4, "apr": 4, "apr.": 4,
	"may":  5,
	"june": 6, "jun": 6, "jun.": 6,
	"july": 7, "jul": 7, "jul.": 7,
	"august": 8, "aug": 8, "aug.": 8,
	"september": 9, "sep": 9, "sep.": 9, "sept": 9, "sept.": 9,
	"october": 10, "oct": 10, "oct.": 10,
	"november": 11, "nov": 11, "nov.": 11,
	"december": 12, "dec": 12, "dec.": 12,
}

// weekdayNames maps lowercase day name forms to time.Weekday.
// Includes full names, 3-letter abbreviations, abbreviations with trailing
// period, and special forms (Tues, Wednes, Thur, Thurs).
var weekdayNames = map[string]time.Weekday{
	"sunday": time.Sunday, "sun": time.Sunday, "sun.": time.Sunday,
	"monday": time.Monday, "mon": time.Monday, "mon.": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday, "tue.": time.Tuesday,
	"tues": time.Tuesday, "tues.": time.Tuesday,
	"wednesday": time.Wednesday, "wed": time.Wednesday, "wed.": time.Wednesday,
	"wednes": time.Wednesday, "wednes.": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday, "thu.": time.Thursday,
	"thur": time.Thursday, "thur.": time.Thursday, "thurs": time.Thursday, "thurs.": time.Thursday,
	"friday": time.Friday, "fri": time.Friday, "fri.": time.Friday,
	"saturday": time.Saturday, "sat": time.Saturday, "sat.": time.Saturday,
}
