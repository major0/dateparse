package dateparse

import "time"

// unitKeywords maps lowercase unit name forms to their relativeUnit constant.
// Includes singular, plural, and abbreviated forms for O(1) lookup.
var unitKeywords = map[string]relativeUnit{
	"second": unitSecond, "seconds": unitSecond, "sec": unitSecond, "secs": unitSecond,
	"minute": unitMinute, "minutes": unitMinute, "min": unitMinute, "mins": unitMinute,
	"hour": unitHour, "hours": unitHour, "hr": unitHour, "hrs": unitHour,
	"day": unitDay, "days": unitDay,
	"week": unitWeek, "weeks": unitWeek, "wk": unitWeek, "wks": unitWeek,
	"fortnight": unitFortnight, "fortnights": unitFortnight,
	"month": unitMonth, "months": unitMonth, "mo": unitMonth, "mos": unitMonth, "mon": unitMonth,
	"year": unitYear, "years": unitYear, "yr": unitYear, "yrs": unitYear,
	"ghurry": unitGhurry, "ghurries": unitGhurry,
	"scruple": unitScruple, "scruples": unitScruple,
	"mileway": unitMileway, "mileways": unitMileway,
	"microfortnight": unitMicrofortnight, "microfortnights": unitMicrofortnight,
	"nundine": unitNundine, "nundines": unitNundine,
	"decade": unitDecade, "decades": unitDecade,
	"century": unitCentury, "centuries": unitCentury,
	"millennium": unitMillennium, "millennia": unitMillennium, "millenniums": unitMillennium,
	"olympiad": unitOlympiad, "olympiads": unitOlympiad,
	"indiction": unitIndiction, "indictions": unitIndiction,
	"lustre": unitLustre, "lustres": unitLustre, "lustrum": unitLustre, "lustra": unitLustre,
	"jiffy": unitJiffy, "jiffies": unitJiffy,
	"moment": unitMoment, "moments": unitMoment,
}

// unitDuration maps fixed-duration relativeUnit values to their time.Duration.
var unitDuration = map[relativeUnit]time.Duration{
	unitSecond:         time.Second,
	unitMinute:         time.Minute,
	unitHour:           time.Hour,
	unitDay:            24 * time.Hour,
	unitWeek:           7 * 24 * time.Hour,
	unitFortnight:      14 * 24 * time.Hour,
	unitGhurry:         24 * time.Minute,
	unitScruple:        60 * time.Second,
	unitMileway:        20 * time.Minute,
	unitMicrofortnight: 1209600 * time.Microsecond,
	unitNundine:        8 * 24 * time.Hour,
	unitJiffy:          10 * time.Millisecond,
	unitMoment:         90 * time.Second,
}

// calendarUnits maps calendar-relative units to {years, months} for time.AddDate.
var calendarUnits = map[relativeUnit][2]int{
	unitMonth:      {0, 1},
	unitYear:       {1, 0},
	unitDecade:     {10, 0},
	unitCentury:    {100, 0},
	unitMillennium: {1000, 0},
	unitOlympiad:   {4, 0},
	unitIndiction:  {15, 0},
	unitLustre:     {5, 0},
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
