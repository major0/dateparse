package dateparse

import (
	"fmt"
	"strconv"
)

// scanner holds the scanning state.
type scanner struct {
	input string // lowercased
	pos   int    // current byte offset
}

// scan consumes the entire input and returns a slice of items.
func (sc *scanner) scan() ([]item, error) {
	var items []item
	for sc.pos < len(sc.input) {
		sc.skipWhitespace()
		if sc.pos >= len(sc.input) {
			break
		}

		its, n, err := sc.matchNext()
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, fmt.Errorf("unexpected input at position %d: %q", sc.pos, sc.remaining())
		}
		items = append(items, its...)
		sc.pos += n
	}
	return items, nil
}

// matchNext tries each matcher in priority order and returns the first match.
func (sc *scanner) matchNext() ([]item, int, error) {
	type matchFunc func() ([]item, int, error)
	for _, m := range []matchFunc{
		sc.matchComment,
		sc.matchEpoch,
		sc.matchRFC3339,
		sc.matchTimeOfDay,
	} {
		its, n, err := m()
		if err != nil {
			return nil, 0, err
		}
		if n > 0 {
			return its, n, nil
		}
	}
	return nil, 0, nil
}

// remaining returns the unconsumed portion of the input.
func (sc *scanner) remaining() string {
	return sc.input[sc.pos:]
}

// skipWhitespace advances past spaces, tabs, and newlines.
func (sc *scanner) skipWhitespace() {
	for sc.pos < len(sc.input) {
		b := sc.input[sc.pos]
		if b == ' ' || b == '\t' || b == '\n' || b == '\r' {
			sc.pos++
		} else {
			break
		}
	}
}

// ---------------------------------------------------------------------------
// Helper functions
// ---------------------------------------------------------------------------

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlpha(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// parseDigits parses a run of ASCII digits from s into an int.
// Callers must ensure s contains only digits (validated via countDigits).
func parseDigits(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// countDigits returns the number of leading ASCII digits in s.
func countDigits(s string) int {
	i := 0
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	return i
}

// daysInMonth returns the number of days in the given month of the given year.
func daysInMonth(year, month int) int {
	switch month {
	case 1, 3, 5, 7, 8, 10, 12:
		return 31
	case 4, 6, 9, 11:
		return 30
	case 2:
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			return 29
		}
		return 28
	}
	return 0
}

// validateDate checks that year/month/day form a valid date.
func validateDate(year, month, day int) error {
	if month < 1 || month > 12 {
		return fmt.Errorf("invalid month: %d (expected 1-12)", month)
	}
	maxDay := daysInMonth(year, month)
	if day < 1 || day > maxDay {
		return fmt.Errorf("invalid day: %d (expected 1-%d for month %d)", day, maxDay, month)
	}
	return nil
}

// validateTime checks that hour/minute/second are in valid ranges.
func validateTime(hour, minute, second int) error {
	if hour > 23 {
		return fmt.Errorf("invalid hour: %d (expected 0-23)", hour)
	}
	if minute > 59 {
		return fmt.Errorf("invalid minute: %d (expected 0-59)", minute)
	}
	if second > 59 {
		return fmt.Errorf("invalid second: %d (expected 0-59)", second)
	}
	return nil
}

// parseFraction parses a fractional seconds string (digits after '.' or ',')
// and returns nanoseconds. Truncates to 9 digits of precision.
func parseFraction(s string) int {
	if len(s) == 0 {
		return 0
	}
	// Truncate to 9 digits.
	if len(s) > 9 {
		s = s[:9]
	}
	n, _ := strconv.Atoi(s)
	// Scale up: if s has fewer than 9 digits, multiply by 10^(9-len(s)).
	for i := len(s); i < 9; i++ {
		n *= 10
	}
	return n
}

// consumeSeconds tries to consume ":SS" at position i in s.
// Returns (second, newPos).
func consumeSeconds(s string, i int) (int, int) {
	if i < len(s) && s[i] == ':' {
		if i+2 < len(s) && isDigit(s[i+1]) && isDigit(s[i+2]) {
			sec := parseDigits(s[i+1 : i+3])
			return sec, i + 3
		}
	}
	return 0, i
}

// consumeFraction tries to consume ".fraction" or ",fraction" at position i in s.
// Returns (nanoseconds, newPos).
func consumeFraction(s string, i int) (int, int) {
	if i < len(s) && (s[i] == '.' || s[i] == ',') {
		i++ // skip separator
		fracStart := i
		for i < len(s) && isDigit(s[i]) {
			i++
		}
		if i > fracStart {
			return parseFraction(s[fracStart:i]), i
		}
	}
	return 0, i
}

// ---------------------------------------------------------------------------
// matchComment — priority 1: parenthetical comments
// ---------------------------------------------------------------------------

// matchComment matches a parenthetical comment with nested parenthesis support.
func (sc *scanner) matchComment() ([]item, int, error) {
	s := sc.remaining()
	if len(s) == 0 || s[0] != '(' {
		return nil, 0, nil
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return []item{{
					typ:   itemComment,
					value: s[1:i], // content without outer parens
					pos:   sc.pos,
				}}, i + 1, nil
			}
		}
	}
	// Unmatched opening paren — not a valid comment.
	return nil, 0, nil
}

// ---------------------------------------------------------------------------
// matchEpoch — priority 2: @<seconds>[.fraction]
// ---------------------------------------------------------------------------

// matchEpoch matches '@' followed by optional sign and digits, optional
// fractional part ('.' or ','). Returns ([]item, bytesConsumed, error).
func (sc *scanner) matchEpoch() ([]item, int, error) {
	s := sc.remaining()
	if len(s) == 0 || s[0] != '@' {
		return nil, 0, nil
	}

	i := 1 // skip '@'
	if i >= len(s) {
		return nil, 0, nil
	}

	// Optional sign.
	negative := false
	switch s[i] {
	case '-':
		negative = true
		i++
	case '+':
		i++
	}

	// Must have at least one digit.
	digitStart := i
	for i < len(s) && isDigit(s[i]) {
		i++
	}
	if i == digitStart {
		return nil, 0, nil
	}

	sec, err := strconv.ParseInt(s[digitStart:i], 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid epoch seconds: %w", err)
	}
	if negative {
		sec = -sec
	}

	// Optional fractional part.
	ns := 0
	ns, i = consumeFraction(s, i)

	return []item{{
		typ:   itemEpoch,
		value: epochSeconds{seconds: sec, nanosecond: ns},
		pos:   sc.pos,
	}}, i, nil
}

// ---------------------------------------------------------------------------
// matchRFC3339 — priority 3: YYYY-MM-DDTHH:MM:SS[.frac][tz]
// ---------------------------------------------------------------------------

// matchRFC3339 matches YYYY-MM-DDTHH:MM:SS with optional fractional seconds
// and timezone offset. Accepts space as equivalent to 'T'. Returns two items
// (calendarDate + timeOfDay), bytes consumed, and error.
func (sc *scanner) matchRFC3339() ([]item, int, error) {
	s := sc.remaining()

	// Need at least YYYY-MM-DDxHH:MM = 16 chars.
	if len(s) < 16 {
		return nil, 0, nil
	}

	// Parse YYYY-MM-DD.
	if countDigits(s) != 4 {
		return nil, 0, nil
	}
	if s[4] != '-' {
		return nil, 0, nil
	}
	if countDigits(s[5:]) < 2 {
		return nil, 0, nil
	}
	if s[7] != '-' {
		return nil, 0, nil
	}
	if countDigits(s[8:]) < 2 {
		return nil, 0, nil
	}

	// Separator must be 't' (lowercased 'T') or space.
	sep := s[10]
	if sep != 't' && sep != ' ' {
		return nil, 0, nil
	}

	// If separator is space, the part after must look like a time (digit + digit + colon).
	if sep == ' ' {
		rest := s[11:]
		if len(rest) < 5 || !isDigit(rest[0]) || !isDigit(rest[1]) || rest[2] != ':' {
			return nil, 0, nil
		}
	}

	// Parse time part: HH:MM
	if len(s) < 16 || !isDigit(s[11]) || !isDigit(s[12]) || s[13] != ':' || !isDigit(s[14]) || !isDigit(s[15]) {
		return nil, 0, nil
	}

	year := parseDigits(s[0:4])
	month := parseDigits(s[5:7])
	day := parseDigits(s[8:10])
	hour := parseDigits(s[11:13])
	minute := parseDigits(s[14:16])

	i := 16
	second := 0
	ns := 0

	// Optional :SS
	second, i = consumeSeconds(s, i)

	// Optional fractional seconds.
	ns, i = consumeFraction(s, i)

	// Optional timezone: Z, +HH:MM, -HH:MM, +HHMM, -HHMM, +HH, -HH
	var tzOff *int
	if i < len(s) {
		off, tzN := parseTZSuffix(s[i:])
		if tzN > 0 {
			tzOff = &off
			i += tzN
		}
	}

	// Validate date components.
	if err := validateDate(year, month, day); err != nil {
		return nil, 0, err
	}

	// Validate time components.
	if err := validateTime(hour, minute, second); err != nil {
		return nil, 0, err
	}

	dateItem := item{
		typ:   itemCalendarDate,
		value: calendarDate{year: year, month: month, day: day},
		pos:   sc.pos,
	}
	timeItem := item{
		typ:   itemTimeOfDay,
		value: timeOfDay{hour: hour, minute: minute, second: second, nanosecond: ns, tzOffset: tzOff},
		pos:   sc.pos,
	}

	return []item{dateItem, timeItem}, i, nil
}

// parseTZSuffix parses a timezone suffix from s. Returns (offsetSeconds, bytesConsumed).
// If no timezone is found, returns (0, 0).
func parseTZSuffix(s string) (int, int) {
	if len(s) == 0 {
		return 0, 0
	}

	// 'z' (lowercased 'Z')
	if s[0] == 'z' {
		// Make sure 'z' is not followed by a letter (would be a timezone name like "zulu").
		if len(s) > 1 && isAlpha(s[1]) {
			return 0, 0
		}
		return 0, 1
	}

	// +/- offset
	if s[0] != '+' && s[0] != '-' {
		return 0, 0
	}

	sign := 1
	if s[0] == '-' {
		sign = -1
	}

	rest := s[1:]
	nd := countDigits(rest)

	switch {
	case nd >= 2 && len(rest) > 2 && rest[2] == ':' && len(rest) > 4 && countDigits(rest[3:]) >= 2:
		// +HH:MM
		hh := parseDigits(rest[0:2])
		mm := parseDigits(rest[3:5])
		return sign * (hh*3600 + mm*60), 6 // sign + HH:MM
	case nd >= 4:
		// +HHMM
		hh := parseDigits(rest[0:2])
		mm := parseDigits(rest[2:4])
		return sign * (hh*3600 + mm*60), 5 // sign + HHMM
	case nd >= 2:
		// +HH
		hh := parseDigits(rest[0:2])
		return sign * (hh * 3600), 3 // sign + HH
	}

	return 0, 0
}

// ---------------------------------------------------------------------------
// matchTimeOfDay — priority 4
// ---------------------------------------------------------------------------

// matchTimeOfDay matches time-of-day in these formats:
//   - HH:MM, HH:MM:SS, HH:MM:SS.fraction
//   - 12-hour: Npm, Npm, N:MMam, N:MMpm (also a.m./p.m.)
//   - Optional trailing timezone correction
//   - Rejects am/pm combined with timezone correction
func (sc *scanner) matchTimeOfDay() ([]item, int, error) {
	s := sc.remaining()
	if len(s) == 0 {
		return nil, 0, nil
	}

	// Try 24-hour format first: HH:MM...
	if its, n, err := sc.matchTime24(s); n > 0 || err != nil {
		return its, n, err
	}

	// Try 12-hour format: N[N][:MM]am/pm
	if its, n, err := sc.matchTime12(s); n > 0 || err != nil {
		return its, n, err
	}

	return nil, 0, nil
}

// matchTime24 matches 24-hour time: HH:MM[:SS[.frac]][tz]
func (sc *scanner) matchTime24(s string) ([]item, int, error) {
	// Need at least H:MM (4 chars) or HH:MM (5 chars).
	nd := countDigits(s)
	if nd < 1 || nd > 2 {
		return nil, 0, nil
	}

	// Must have colon after the hour digits.
	if nd >= len(s) || s[nd] != ':' {
		return nil, 0, nil
	}

	// Must have two digits after the colon.
	if nd+3 > len(s) || !isDigit(s[nd+1]) || !isDigit(s[nd+2]) {
		return nil, 0, nil
	}

	hour := parseDigits(s[0:nd])
	minute := parseDigits(s[nd+1 : nd+3])
	i := nd + 3

	second := 0
	ns := 0

	// Optional :SS
	second, i = consumeSeconds(s, i)

	// Optional fractional seconds.
	ns, i = consumeFraction(s, i)

	// Check for am/pm after 24-hour digits — this makes it a 12-hour time,
	// so bail and let matchTime12 handle it.
	if i < len(s) {
		rest := s[i:]
		if hasAMPMPrefix(rest) {
			return nil, 0, nil
		}
	}

	// Optional timezone correction.
	var tzOff *int
	if i < len(s) {
		// Skip optional whitespace before tz.
		j := i
		for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
			j++
		}
		off, tzN := parseTZSuffix(s[j:])
		if tzN > 0 {
			tzOff = &off
			i = j + tzN
		}
	}

	// Validate.
	if err := validateTime(hour, minute, second); err != nil {
		return nil, 0, err
	}

	return []item{{
		typ:   itemTimeOfDay,
		value: timeOfDay{hour: hour, minute: minute, second: second, nanosecond: ns, tzOffset: tzOff},
		pos:   sc.pos,
	}}, i, nil
}

// matchTime12 matches 12-hour time: N[N][:MM[:SS[.frac]]] am/pm/a.m./p.m.
func (sc *scanner) matchTime12(s string) ([]item, int, error) {
	nd := countDigits(s)
	if nd < 1 || nd > 2 {
		return nil, 0, nil
	}

	hour := parseDigits(s[0:nd])
	i := nd
	minute := 0
	second := 0
	ns := 0

	// Optional :MM
	if i < len(s) && s[i] == ':' {
		if i+2 < len(s) && isDigit(s[i+1]) && isDigit(s[i+2]) {
			minute = parseDigits(s[i+1 : i+3])
			i += 3

			// Optional :SS
			second, i = consumeSeconds(s, i)

			// Optional fractional seconds.
			ns, i = consumeFraction(s, i)
		}
	}

	// Skip optional whitespace before am/pm.
	j := i
	for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
		j++
	}

	// Must have am/pm suffix.
	isPM, ampmLen := parseAMPM(s[j:])
	if ampmLen == 0 {
		return nil, 0, nil
	}
	i = j + ampmLen

	// Validate 12-hour range.
	if hour < 1 || hour > 12 {
		return nil, 0, fmt.Errorf("invalid hour: %d (expected 1-12 for 12-hour format)", hour)
	}
	if minute > 59 {
		return nil, 0, fmt.Errorf("invalid minute: %d (expected 0-59)", minute)
	}
	if second > 59 {
		return nil, 0, fmt.Errorf("invalid second: %d (expected 0-59)", second)
	}

	// Convert to 24-hour.
	if hour == 12 {
		if !isPM {
			hour = 0 // 12am → 0
		}
		// 12pm stays 12
	} else if isPM {
		hour += 12
	}

	// Check for timezone correction — reject with am/pm.
	afterAMPM := i
	for afterAMPM < len(s) && (s[afterAMPM] == ' ' || s[afterAMPM] == '\t') {
		afterAMPM++
	}
	if afterAMPM < len(s) {
		_, tzN := parseTZSuffix(s[afterAMPM:])
		if tzN > 0 {
			return nil, 0, fmt.Errorf("am/pm cannot combine with timezone correction")
		}
	}

	return []item{{
		typ:   itemTimeOfDay,
		value: timeOfDay{hour: hour, minute: minute, second: second, nanosecond: ns, tzOffset: nil},
		pos:   sc.pos,
	}}, i, nil
}

// hasAMPMPrefix checks if s starts with an am/pm indicator.
func hasAMPMPrefix(s string) bool {
	_, n := parseAMPM(s)
	return n > 0
}

// parseAMPM checks if s starts with am/pm/a.m./p.m. (case-insensitive, input
// is already lowercased). Returns (isPM, bytesConsumed). If no match, returns
// (false, 0).
func parseAMPM(s string) (bool, int) {
	if len(s) == 0 {
		return false, 0
	}

	// Try "a.m." / "p.m." first (longer match).
	if len(s) >= 4 {
		if s[0:4] == "a.m." {
			return false, 4
		}
		if s[0:4] == "p.m." {
			return true, 4
		}
	}

	// Try "am" / "pm".
	if len(s) >= 2 {
		if s[0:2] == "am" {
			// Make sure it's not followed by more letters (e.g. "amend").
			if len(s) > 2 && isAlpha(s[2]) && s[2] != '.' {
				return false, 0
			}
			return false, 2
		}
		if s[0:2] == "pm" {
			if len(s) > 2 && isAlpha(s[2]) && s[2] != '.' {
				return false, 0
			}
			return true, 2
		}
	}

	return false, 0
}
