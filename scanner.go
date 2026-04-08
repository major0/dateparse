package dateparse

import (
	"fmt"
	"strconv"
	"time"
)

// scanner holds the scanning state.
type scanner struct {
	input             string    // lowercased
	pos               int       // current byte offset
	st                state     // accumulated state
	ref               time.Time // reference time for resolving anchors
	parseDurationMode bool      // when true, anchors and direction-to-anchor ops are errors
}

// scan consumes the entire input and returns the accumulated state.
func (sc *scanner) scan() (*state, error) {
	for sc.pos < len(sc.input) {
		sc.skipWhitespace()
		if sc.pos >= len(sc.input) {
			break
		}

		n, matched, err := sc.matchNext()
		if err != nil {
			return nil, err
		}
		if !matched {
			return nil, fmt.Errorf("unexpected input at position %d: %q", sc.pos, sc.remaining())
		}
		sc.pos += n
	}

	// End-of-scan validation: if direction is -1 or +1 (before/after was set)
	// but no anchor followed, return error.
	if sc.st.direction == -1 || sc.st.direction == 1 {
		return nil, fmt.Errorf("\"before\"/\"after\" without following anchor")
	}
	if len(sc.st.pendingOps) > 0 {
		return nil, fmt.Errorf("\"before\"/\"after\" without following anchor")
	}

	return &sc.st, nil
}

// matchNext tries each matcher in priority order and returns the first match.
func (sc *scanner) matchNext() (int, bool, error) {
	if n, ok, err := sc.matchComment(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchEpoch(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchRFC3339(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchTimeOfDay(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchCalendarDate(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchTimezone(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchNamedRef(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchDayOfWeek(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchRelative(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchDirectionOp(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchPureNumber(); ok || err != nil {
		return n, ok, err
	}
	if n, ok, err := sc.matchNoise(); ok || err != nil {
		return n, ok, err
	}
	return 0, false, nil
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

// asciiLower returns s with all ASCII uppercase letters converted to lowercase.
// Avoids the allocation of strings.ToLower for ASCII-only inputs.
func asciiLower(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			// Found an uppercase letter — need to allocate.
			b := make([]byte, len(s))
			copy(b, s[:i])
			b[i] = s[i] + 32
			for j := i + 1; j < len(s); j++ {
				if s[j] >= 'A' && s[j] <= 'Z' {
					b[j] = s[j] + 32
				} else {
					b[j] = s[j]
				}
			}
			return string(b)
		}
	}
	// Already lowercase — return original string (no allocation).
	return s
}

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
// Skips the content without updating state.
//
//nolint:unparam // error return is part of the matcher interface contract
func (sc *scanner) matchComment() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 || s[0] != '(' {
		return 0, false, nil
	}
	depth := 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i + 1, true, nil
			}
		}
	}
	// Unmatched opening paren — not a valid comment.
	return 0, false, nil
}

// ---------------------------------------------------------------------------
// matchEpoch — priority 2: @<seconds>[.fraction]
// ---------------------------------------------------------------------------

// matchEpoch matches '@' followed by optional sign and digits, optional
// fractional part ('.' or ','). Sets sc.st.anchor to the resolved time.
func (sc *scanner) matchEpoch() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 || s[0] != '@' {
		return 0, false, nil
	}

	i := 1 // skip '@'
	if i >= len(s) {
		return 0, false, nil
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
		return 0, false, nil
	}

	sec, err := strconv.ParseInt(s[digitStart:i], 10, 64)
	if err != nil {
		return 0, false, fmt.Errorf("invalid epoch seconds: %w", err)
	}
	if negative {
		sec = -sec
	}

	// Optional fractional part.
	ns := 0
	ns, i = consumeFraction(s, i)

	t := time.Unix(sec, int64(ns))
	if err := sc.setAnchor(t); err != nil {
		return 0, false, err
	}

	return i, true, nil
}

// ---------------------------------------------------------------------------
// matchRFC3339 — priority 3: YYYY-MM-DDTHH:MM:SS[.frac][tz]
// ---------------------------------------------------------------------------

// matchRFC3339 matches YYYY-MM-DDTHH:MM:SS with optional fractional seconds
// and timezone offset. Accepts space as equivalent to 'T'. Sets sc.st.anchor
// to the resolved date and sc.st.timeOfDay to the parsed time.
func (sc *scanner) matchRFC3339() (int, bool, error) {
	s := sc.remaining()

	// Need at least YYYY-MM-DDxHH:MM = 16 chars.
	if len(s) < 16 {
		return 0, false, nil
	}

	// Parse YYYY-MM-DD.
	if countDigits(s) != 4 {
		return 0, false, nil
	}
	if s[4] != '-' {
		return 0, false, nil
	}
	if countDigits(s[5:]) < 2 {
		return 0, false, nil
	}
	if s[7] != '-' {
		return 0, false, nil
	}
	if countDigits(s[8:]) < 2 {
		return 0, false, nil
	}

	// Separator must be 't' (lowercased 'T') or space.
	sep := s[10]
	if sep != 't' && sep != ' ' {
		return 0, false, nil
	}

	// If separator is space, the part after must look like a time (digit + digit + colon).
	if sep == ' ' {
		rest := s[11:]
		if len(rest) < 5 || !isDigit(rest[0]) || !isDigit(rest[1]) || rest[2] != ':' {
			return 0, false, nil
		}
	}

	// Parse time part: HH:MM
	if len(s) < 16 || !isDigit(s[11]) || !isDigit(s[12]) || s[13] != ':' || !isDigit(s[14]) || !isDigit(s[15]) {
		return 0, false, nil
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
		return 0, false, err
	}

	// Validate time components.
	if err := validateTime(hour, minute, second); err != nil {
		return 0, false, err
	}

	// Set anchor to the resolved date at midnight UTC.
	anchor := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if err := sc.setAnchor(anchor); err != nil {
		return 0, false, err
	}

	// Set time-of-day.
	if err := sc.setTimeOfDay(timeOfDay{
		hour:       hour,
		minute:     minute,
		second:     second,
		nanosecond: ns,
		tzOffset:   tzOff,
	}); err != nil {
		return 0, false, err
	}

	return i, true, nil
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
//
// Sets sc.st.timeOfDay on match.
func (sc *scanner) matchTimeOfDay() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 {
		return 0, false, nil
	}

	// Try 24-hour format first: HH:MM...
	if n, matched, err := sc.matchTime24(s); matched || err != nil {
		return n, matched, err
	}

	// Try 12-hour format: N[N][:MM]am/pm
	if n, matched, err := sc.matchTime12(s); matched || err != nil {
		return n, matched, err
	}

	return 0, false, nil
}

// matchTime24 matches 24-hour time: HH:MM[:SS[.frac]][tz]
func (sc *scanner) matchTime24(s string) (int, bool, error) {
	// Need at least H:MM (4 chars) or HH:MM (5 chars).
	nd := countDigits(s)
	if nd < 1 || nd > 2 {
		return 0, false, nil
	}

	// Must have colon after the hour digits.
	if nd >= len(s) || s[nd] != ':' {
		return 0, false, nil
	}

	// Must have two digits after the colon.
	if nd+3 > len(s) || !isDigit(s[nd+1]) || !isDigit(s[nd+2]) {
		return 0, false, nil
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
			return 0, false, nil
		}
	}

	// Optional timezone correction.
	var tzOff *int
	if i < len(s) {
		// Skip optional whitespace before tz.
		j := i + skipSpacesAndTabs(s[i:])
		off, tzN := parseTZSuffix(s[j:])
		if tzN > 0 {
			tzOff = &off
			i = j + tzN
		}
	}

	// Validate.
	if err := validateTime(hour, minute, second); err != nil {
		return 0, false, err
	}

	if err := sc.setTimeOfDay(timeOfDay{
		hour:       hour,
		minute:     minute,
		second:     second,
		nanosecond: ns,
		tzOffset:   tzOff,
	}); err != nil {
		return 0, false, err
	}

	return i, true, nil
}

// matchTime12 matches 12-hour time: N[N][:MM[:SS[.frac]]] am/pm/a.m./p.m.
func (sc *scanner) matchTime12(s string) (int, bool, error) {
	nd := countDigits(s)
	if nd < 1 || nd > 2 {
		return 0, false, nil
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
	j := i + skipSpacesAndTabs(s[i:])

	// Must have am/pm suffix.
	isPM, ampmLen := parseAMPM(s[j:])
	if ampmLen == 0 {
		return 0, false, nil
	}
	i = j + ampmLen

	// Validate 12-hour range.
	if hour < 1 || hour > 12 {
		return 0, false, fmt.Errorf("invalid hour: %d (expected 1-12 for 12-hour format)", hour)
	}
	if minute > 59 {
		return 0, false, fmt.Errorf("invalid minute: %d (expected 0-59)", minute)
	}
	if second > 59 {
		return 0, false, fmt.Errorf("invalid second: %d (expected 0-59)", second)
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
	afterAMPM := i + skipSpacesAndTabs(s[i:])
	if afterAMPM < len(s) {
		_, tzN := parseTZSuffix(s[afterAMPM:])
		if tzN > 0 {
			return 0, false, fmt.Errorf("am/pm cannot combine with timezone correction")
		}
	}

	if err := sc.setTimeOfDay(timeOfDay{
		hour:       hour,
		minute:     minute,
		second:     second,
		nanosecond: ns,
		tzOffset:   nil,
	}); err != nil {
		return 0, false, err
	}

	return i, true, nil
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

// ---------------------------------------------------------------------------
// Helper functions for new matchers
// ---------------------------------------------------------------------------

// skipSpacesAndTabs returns the number of leading spaces and tabs in s.
func skipSpacesAndTabs(s string) int {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return i
}

// extractWord extracts the next run of ASCII letters from s[0:].
// Returns the lowercase word and its length.
func extractWord(s string) (string, int) {
	i := 0
	for i < len(s) && isAlpha(s[i]) {
		i++
	}
	return s[:i], i
}

// applyDeltaToTime applies delta d to time t with the given sign (+1 or -1).
func applyDeltaToTime(t time.Time, d delta, sign int) time.Time {
	t = t.AddDate(sign*d.years, sign*d.months, sign*d.days)
	dur := time.Duration(sign*d.hours)*time.Hour +
		time.Duration(sign*d.minutes)*time.Minute +
		time.Duration(sign*d.seconds)*time.Second +
		time.Duration(sign*d.nanos)
	return t.Add(dur)
}

// negateDelta negates all fields of d.
func negateDelta(d *delta) {
	d.years = -d.years
	d.months = -d.months
	d.days = -d.days
	d.hours = -d.hours
	d.minutes = -d.minutes
	d.seconds = -d.seconds
	d.nanos = -d.nanos
}

// resetDelta zeroes out all fields of d.
func resetDelta(d *delta) {
	*d = delta{}
}

// isDeltaZero returns true if all delta fields are zero.
func isDeltaZero(d delta) bool {
	return d == delta{}
}

// resolveWeekday resolves the Nth occurrence of a weekday relative to ref.
//   - ordinal -1 (last): most recent occurrence before ref
//   - ordinal 0 (this): occurrence in current week (week starts Monday)
//   - ordinal 1 (next/first): first occurrence after ref
//   - ordinal 2-12: Nth occurrence after ref
func resolveWeekday(ref time.Time, day time.Weekday, ordinal int) time.Time {
	refDay := ref.Weekday()
	// Truncate ref to midnight UTC.
	base := time.Date(ref.Year(), ref.Month(), ref.Day(), 0, 0, 0, 0, time.UTC)

	switch {
	case ordinal < 0:
		// Last: most recent occurrence before ref.
		diff := int(refDay) - int(day)
		if diff <= 0 {
			diff += 7
		}
		return base.AddDate(0, 0, -diff)

	case ordinal == 0:
		// This: occurrence in current week. Week starts Monday.
		// Find the Monday of the current week.
		mondayOffset := int(refDay) - int(time.Monday)
		if mondayOffset < 0 {
			mondayOffset += 7
		}
		monday := base.AddDate(0, 0, -mondayOffset)
		// Target day in this week.
		dayOffset := int(day) - int(time.Monday)
		if dayOffset < 0 {
			dayOffset += 7
		}
		return monday.AddDate(0, 0, dayOffset)

	default:
		// ordinal >= 1: Nth occurrence after ref.
		diff := int(day) - int(refDay)
		if diff <= 0 {
			diff += 7
		}
		// First occurrence after ref.
		first := base.AddDate(0, 0, diff)
		// Nth occurrence.
		return first.AddDate(0, 0, (ordinal-1)*7)
	}
}

// setAnchor sets the anchor on state, checking for conflicts.
func (sc *scanner) setAnchor(t time.Time) error {
	if sc.parseDurationMode {
		return fmt.Errorf("ParseDuration: expression contains an anchor (use Parse instead)")
	}
	// If there are pending direction ops, apply them in reverse order (innermost first).
	if len(sc.st.pendingOps) > 0 {
		// First apply any remaining delta with the current direction.
		if !isDeltaZero(sc.st.delta) {
			t = applyDeltaToTime(t, sc.st.delta, sc.st.direction)
			resetDelta(&sc.st.delta)
		}
		// Apply stacked ops in reverse order.
		for i := len(sc.st.pendingOps) - 1; i >= 0; i-- {
			op := sc.st.pendingOps[i]
			t = applyDeltaToTime(t, op.d, op.dir)
		}
		sc.st.pendingOps = nil
		sc.st.direction = 0
		sc.st.anchor = t
		sc.st.anchorSet = true
		return nil
	}
	// If there's a pending direction (before/after), apply the accumulated
	// delta to this anchor.
	if sc.st.direction == -1 || sc.st.direction == 1 {
		t = applyDeltaToTime(t, sc.st.delta, sc.st.direction)
		resetDelta(&sc.st.delta)
		sc.st.direction = 0
		sc.st.anchor = t
		sc.st.anchorSet = true
		return nil
	}
	if sc.st.anchorSet {
		return fmt.Errorf("conflicting anchor: anchor already set")
	}
	sc.st.anchor = t
	sc.st.anchorSet = true
	return nil
}

// setTimeOfDay sets the timeOfDay on state, checking for conflicts.
func (sc *scanner) setTimeOfDay(tod timeOfDay) error {
	if sc.st.todSet {
		return fmt.Errorf("conflicting time-of-day: time already set")
	}
	sc.st.tod = tod
	sc.st.todSet = true
	return nil
}

// addToDelta adds N * scale to the appropriate delta field.
func addToDelta(d *delta, field deltaField, n int) {
	switch field {
	case fieldYears:
		d.years += n
	case fieldMonths:
		d.months += n
	case fieldDays:
		d.days += n
	case fieldHours:
		d.hours += n
	case fieldMinutes:
		d.minutes += n
	case fieldSeconds:
		d.seconds += n
	case fieldNanos:
		d.nanos += n
	}
}

// cascadeScale maps each field to {next smaller field, multiplier to convert}.
// years→months(×12), months→days(×30), days→hours(×24),
// hours→minutes(×60), minutes→seconds(×60), seconds→nanos(×1e9).
var cascadeScale = [...]struct {
	next deltaField
	mult float64
}{
	fieldYears:   {fieldMonths, 12},
	fieldMonths:  {fieldDays, 30},
	fieldDays:    {fieldHours, 24},
	fieldHours:   {fieldMinutes, 60},
	fieldMinutes: {fieldSeconds, 60},
	fieldSeconds: {fieldNanos, 1e9},
	fieldNanos:   {fieldNanos, 0}, // terminal
}

// addFractionalToDelta adds a floating-point value to the delta, cascading
// the fractional remainder into progressively smaller fields.
// E.g. 2.5 days → 2 days + 12 hours.
func addFractionalToDelta(d *delta, field deltaField, val float64) {
	for field <= fieldNanos {
		whole := int(val)
		if val < 0 && float64(whole) != val {
			whole-- // floor toward negative infinity
		}
		addToDelta(d, field, whole)
		frac := val - float64(whole)
		if frac == 0 || field == fieldNanos {
			break
		}
		cs := cascadeScale[field]
		if cs.mult == 0 {
			break
		}
		field = cs.next
		val = frac * cs.mult
	}
}

// extractMonthWord tries to extract a month name (possibly with trailing period)
// from s. Returns (monthNumber, bytesConsumed) or (0, 0) if no match.
func extractMonthWord(s string) (int, int) {
	// Extract alpha run.
	word, wlen := extractWord(s)
	if wlen == 0 {
		return 0, 0
	}
	// Try with trailing period first.
	if wlen < len(s) && s[wlen] == '.' {
		withDot := word + "."
		if m, ok := monthNames[withDot]; ok {
			return m, wlen + 1
		}
	}
	if m, ok := monthNames[word]; ok {
		return m, wlen
	}
	return 0, 0
}

// extractWeekdayWord tries to extract a weekday name (possibly with trailing period)
// from s. Returns (weekday, bytesConsumed, found).
func extractWeekdayWord(s string) (time.Weekday, int, bool) {
	word, wlen := extractWord(s)
	if wlen == 0 {
		return 0, 0, false
	}
	// Try with trailing period first.
	if wlen < len(s) && s[wlen] == '.' {
		withDot := word + "."
		if wd, ok := weekdayNames[withDot]; ok {
			return wd, wlen + 1, true
		}
	}
	if wd, ok := weekdayNames[word]; ok {
		return wd, wlen, true
	}
	return 0, 0, false
}

// ---------------------------------------------------------------------------
// matchCalendarDate — priority 5
// ---------------------------------------------------------------------------

// matchCalendarDate matches calendar dates in various formats:
//   - ISO "YYYY-MM-DD" (must NOT be followed by T/space+digits — that's RFC 3339)
//   - Two-digit year "YY-MM-DD" (69-99→1969-1999, 00-68→2000-2068)
//   - US format "M/D/YYYY" and "M/D"
//   - Literal month forms: "D Mon YYYY", "D Mon", "Mon D YYYY", "Mon D, YYYY", "D-Mon-YYYY"
func (sc *scanner) matchCalendarDate() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 {
		return 0, false, nil
	}

	// Try literal month forms first (they start with alpha or digit).
	if n, matched, err := sc.matchLiteralMonthDate(s); matched || err != nil {
		return n, matched, err
	}

	// Try numeric forms.
	nd := countDigits(s)

	// ISO "YYYY-MM-DD" — 4 digits, dash, 2 digits, dash, 2 digits.
	if nd == 4 && len(s) > 4 && s[4] == '-' {
		return sc.matchISODate(s)
	}

	// Two-digit year "YY-MM-DD" — 2 digits, dash, 2 digits, dash, 2 digits.
	if nd == 2 && len(s) > 2 && s[2] == '-' {
		return sc.matchTwoDigitYearDate(s)
	}

	// US format with slash: M/D/YYYY or M/D.
	if nd >= 1 && nd <= 2 {
		afterDigits := s[nd:]
		if len(afterDigits) > 0 && afterDigits[0] == '/' {
			return sc.matchUSDate(s, nd)
		}
	}

	return 0, false, nil
}

// matchISODate matches "YYYY-MM-DD" (not followed by T or space+digit which is RFC 3339).
func (sc *scanner) matchISODate(s string) (int, bool, error) {
	// YYYY-MM-DD = 10 chars minimum.
	if len(s) < 10 {
		return 0, false, nil
	}
	if countDigits(s[5:]) < 2 || s[7] != '-' || countDigits(s[8:]) < 2 {
		return 0, false, nil
	}

	year := parseDigits(s[0:4])
	month := parseDigits(s[5:7])
	day := parseDigits(s[8:10])

	// Check if this is actually an RFC 3339 prefix: followed by 't' or space+digit:colon.
	// RFC 3339 matcher has higher priority, so if it didn't match, this is a standalone date.
	// But we still need to guard against "YYYY-MM-DD HH:MM" being split.
	if len(s) > 10 {
		next := s[10]
		if next == 't' {
			return 0, false, nil // let RFC 3339 handle it
		}
		if next == ' ' && len(s) > 11 && isDigit(s[11]) {
			// Check if it looks like a time: space + digit(s) + colon.
			rest := s[11:]
			digs := countDigits(rest)
			if digs >= 1 && digs+11 < len(s) && s[11+digs] == ':' {
				return 0, false, nil // let RFC 3339 handle it
			}
		}
	}

	if err := validateDate(year, month, day); err != nil {
		return 0, false, err
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if err := sc.setAnchor(t); err != nil {
		return 0, false, err
	}
	return 10, true, nil
}

// matchTwoDigitYearDate matches "YY-MM-DD".
func (sc *scanner) matchTwoDigitYearDate(s string) (int, bool, error) {
	if len(s) < 8 {
		return 0, false, nil
	}
	if countDigits(s[3:]) < 2 || s[5] != '-' || countDigits(s[6:]) < 2 {
		return 0, false, nil
	}

	yy := parseDigits(s[0:2])
	month := parseDigits(s[3:5])
	day := parseDigits(s[6:8])

	var year int
	if yy >= 69 {
		year = 1900 + yy
	} else {
		year = 2000 + yy
	}

	if err := validateDate(year, month, day); err != nil {
		return 0, false, err
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if err := sc.setAnchor(t); err != nil {
		return 0, false, err
	}
	return 8, true, nil
}

// matchUSDate matches "M/D/YYYY" or "M/D" (year omitted → ref year).
func (sc *scanner) matchUSDate(s string, monthDigits int) (int, bool, error) {
	i := monthDigits + 1 // skip past month digits and '/'
	dayStart := i
	dayDigits := countDigits(s[i:])
	if dayDigits < 1 || dayDigits > 2 {
		return 0, false, nil
	}
	i += dayDigits

	month := parseDigits(s[0:monthDigits])
	day := parseDigits(s[dayStart : dayStart+dayDigits])

	// Check for /YYYY.
	year := sc.ref.Year()
	consumed := i
	if i < len(s) && s[i] == '/' {
		rest := s[i+1:]
		yd := countDigits(rest)
		if yd == 4 {
			year = parseDigits(rest[0:4])
			consumed = i + 1 + 4
		} else if yd > 0 {
			// Slash followed by non-4-digit number — not a valid US date.
			return 0, false, nil
		}
	}

	if err := validateDate(year, month, day); err != nil {
		return 0, false, err
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if err := sc.setAnchor(t); err != nil {
		return 0, false, err
	}
	return consumed, true, nil
}

// matchLiteralMonthDate matches forms with month names:
//   - "D Mon YYYY", "D Mon" (number, space, month name, optional space+year)
//   - "Mon D YYYY", "Mon D, YYYY" (month name, space, number, optional comma+space+year)
//   - "D-Mon-YYYY" (number, dash, month name, dash, year)
func (sc *scanner) matchLiteralMonthDate(s string) (int, bool, error) {
	if len(s) == 0 {
		return 0, false, nil
	}

	// If starts with digit: try "D Mon YYYY", "D Mon", "D-Mon-YYYY".
	nd := countDigits(s)
	if nd >= 1 && nd <= 2 {
		day := parseDigits(s[0:nd])

		// "D-Mon-YYYY"
		if nd < len(s) && s[nd] == '-' {
			rest := s[nd+1:]
			month, mlen := extractMonthWord(rest)
			if month > 0 && mlen < len(rest) && rest[mlen] == '-' {
				yearRest := rest[mlen+1:]
				yd := countDigits(yearRest)
				if yd == 4 {
					year := parseDigits(yearRest[0:4])
					if err := validateDate(year, month, day); err != nil {
						return 0, false, err
					}
					t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
					if err := sc.setAnchor(t); err != nil {
						return 0, false, err
					}
					return nd + 1 + mlen + 1 + 4, true, nil
				}
			}
			// Not a D-Mon-YYYY form; fall through (don't consume the digits).
		}

		// "D Mon YYYY" or "D Mon"
		if nd < len(s) && s[nd] == ' ' {
			rest := s[nd+1:]
			month, mlen := extractMonthWord(rest)
			if month > 0 {
				i := nd + 1 + mlen
				year := sc.ref.Year()
				// Check for optional space + 4-digit year.
				if i < len(s) && s[i] == ' ' {
					yearRest := s[i+1:]
					yd := countDigits(yearRest)
					if yd == 4 {
						year = parseDigits(yearRest[0:4])
						i += 1 + 4
					}
				}
				if err := validateDate(year, month, day); err != nil {
					return 0, false, err
				}
				t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
				if err := sc.setAnchor(t); err != nil {
					return 0, false, err
				}
				return i, true, nil
			}
		}
	}

	// If starts with alpha: try "Mon D YYYY", "Mon D, YYYY".
	if isAlpha(s[0]) {
		month, mlen := extractMonthWord(s)
		if month > 0 {
			rest := s[mlen:]
			// Need space + digits.
			if len(rest) > 0 && rest[0] == ' ' {
				rest = rest[1:]
				dd := countDigits(rest)
				if dd >= 1 && dd <= 2 {
					day := parseDigits(rest[0:dd])
					i := mlen + 1 + dd
					year := sc.ref.Year()

					// Check for optional comma.
					afterDay := s[i:]
					if len(afterDay) > 0 && afterDay[0] == ',' {
						i++
						afterDay = s[i:]
					}

					// Check for optional space + 4-digit year.
					if len(afterDay) > 0 && afterDay[0] == ' ' {
						yearRest := afterDay[1:]
						yd := countDigits(yearRest)
						if yd == 4 {
							year = parseDigits(yearRest[0:4])
							i += 1 + 4
						}
					}

					if err := validateDate(year, month, day); err != nil {
						return 0, false, err
					}
					t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
					if err := sc.setAnchor(t); err != nil {
						return 0, false, err
					}
					return i, true, nil
				}
			}
		}
	}

	return 0, false, nil
}

// ---------------------------------------------------------------------------
// matchTimezone — priority 6
// ---------------------------------------------------------------------------

// matchTimezone matches timezone specifications:
//   - "utc" or "z" → offset 0
//   - "u.t.c." → strip periods, match as "utc"
//   - "utc+05:30", "utc-04" → UTC base + numeric correction
//   - Standalone numeric: "+0530", "-04:00" etc.
//   - "DST" suffix → add 3600 to offset
//
//nolint:unparam // error return is part of the matcher interface contract
func (sc *scanner) matchTimezone() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 {
		return 0, false, nil
	}

	offset := 0
	consumed := 0

	// Try "u.t.c." (6 chars).
	switch {
	case len(s) >= 6 && s[0:6] == "u.t.c.":
		consumed = 6
	case len(s) >= 3 && s[0:3] == "utc":
		if len(s) > 3 && isAlpha(s[3]) {
			return 0, false, nil
		}
		consumed = 3
	case s[0] == 'z':
		if len(s) > 1 && isAlpha(s[1]) {
			return 0, false, nil
		}
		consumed = 1
	case s[0] == '+' || s[0] == '-':
		// Standalone numeric offset.
		off, n := parseTZSuffix(s)
		if n > 0 {
			offset = off
			consumed = n
			sc.st.tod.tzOffset = &offset
			sc.st.todSet = true
			return consumed, true, nil
		}
		return 0, false, nil
	default:
		return 0, false, nil
	}

	// Check for numeric correction after utc/z: "+05:30", "-04", etc.
	if consumed < len(s) && (s[consumed] == '+' || s[consumed] == '-') {
		off, n := parseTZSuffix(s[consumed:])
		if n > 0 {
			offset += off
			consumed += n
		}
	}

	// Check for DST suffix.
	rest := s[consumed:]
	// Skip optional whitespace.
	ws := skipSpacesAndTabs(rest)
	if ws+3 <= len(rest) {
		word, wlen := extractWord(rest[ws:])
		if word == "dst" && wlen == 3 {
			offset += 3600
			consumed += ws + 3
		}
	}

	sc.st.tod.tzOffset = &offset
	sc.st.todSet = true
	return consumed, true, nil
}

// ---------------------------------------------------------------------------
// matchNamedRef — priority 7
// ---------------------------------------------------------------------------

// matchNamedRef matches named references: now, today, yesterday, tomorrow, this.
func (sc *scanner) matchNamedRef() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 || !isAlpha(s[0]) {
		return 0, false, nil
	}

	word, wlen := extractWord(s)
	if wlen == 0 {
		return 0, false, nil
	}

	switch word {
	case "now", "today":
		t := sc.ref
		if err := sc.setAnchor(t); err != nil {
			return 0, false, err
		}
		return wlen, true, nil

	case "yesterday":
		t := sc.ref.AddDate(0, 0, -1)
		if err := sc.setAnchor(t); err != nil {
			return 0, false, err
		}
		return wlen, true, nil

	case "tomorrow":
		t := sc.ref.AddDate(0, 0, 1)
		if err := sc.setAnchor(t); err != nil {
			return 0, false, err
		}
		return wlen, true, nil

	case "this":
		// "this" sets the thisModifier flag (zero displacement modifier).
		sc.st.thisModifier = true
		return wlen, true, nil
	}

	return 0, false, nil
}

// ---------------------------------------------------------------------------
// matchDayOfWeek — priority 8
// ---------------------------------------------------------------------------

// matchDayOfWeek matches optional ordinal modifier + day name.
// Resolves to the correct occurrence of that weekday relative to sc.ref.
func (sc *scanner) matchDayOfWeek() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 || !isAlpha(s[0]) {
		return 0, false, nil
	}

	word, wlen := extractWord(s)
	if wlen == 0 {
		return 0, false, nil
	}

	ordinal := 1 // default: next occurrence
	consumed := 0

	// Check if the word is an ordinal modifier.
	if ord, ok := ordinalModifiers[word]; ok {
		// Look ahead for a day name.
		rest := s[wlen:]
		// Skip whitespace.
		ws := skipSpacesAndTabs(rest)
		if ws > 0 || (len(rest) > 0 && isAlpha(rest[0])) {
			wd, dayLen, found := extractWeekdayWord(rest[ws:])
			if found {
				ordinal = ord
				consumed = wlen + ws + dayLen
				return sc.finishDayOfWeek(s, consumed, wd, ordinal)
			}
		}
		// Ordinal not followed by a day name — don't match (let other matchers handle).
		return 0, false, nil
	}

	// Check if the word itself is a day name.
	wd, dayLen, found := extractWeekdayWord(s)
	if found {
		consumed = dayLen
		return sc.finishDayOfWeek(s, consumed, wd, ordinal)
	}

	return 0, false, nil
}

// finishDayOfWeek resolves the weekday, handles trailing comma/period, and sets anchor.
func (sc *scanner) finishDayOfWeek(s string, consumed int, wd time.Weekday, ordinal int) (int, bool, error) {
	// Handle trailing comma.
	if consumed < len(s) && s[consumed] == ',' {
		consumed++
	}

	// Use thisModifier to override ordinal to 0 if "this" was seen.
	if sc.st.thisModifier {
		ordinal = 0
		sc.st.thisModifier = false
	}

	t := resolveWeekday(sc.ref, wd, ordinal)
	if err := sc.setAnchor(t); err != nil {
		return 0, false, err
	}
	return consumed, true, nil
}

// ---------------------------------------------------------------------------
// matchRelative — priority 9
// ---------------------------------------------------------------------------

// matchRelative matches [N] unit_keyword and accumulates into delta.
//
//nolint:unparam // error return is part of the matcher interface contract
func (sc *scanner) matchRelative() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 {
		return 0, false, nil
	}

	consumed := 0
	var multiplier float64 = 1
	hasNumber := false
	isFractional := false

	// Try to parse an optional signed number (integer or decimal).
	if s[0] == '+' || s[0] == '-' || isDigit(s[0]) {
		sign := float64(1)
		i := 0
		switch s[0] {
		case '+':
			i = 1
		case '-':
			sign = -1
			i = 1
		}
		nd := countDigits(s[i:])
		if nd > 0 {
			multiplier = sign * float64(parseDigits(s[i:i+nd]))
			consumed = i + nd
			hasNumber = true
			// Check for decimal point.
			if consumed < len(s) && s[consumed] == '.' {
				consumed++ // skip '.'
				fd := countDigits(s[consumed:])
				if fd > 0 {
					frac := float64(parseDigits(s[consumed : consumed+fd]))
					for k := 0; k < fd; k++ {
						frac /= 10
					}
					if sign < 0 {
						multiplier -= frac
					} else {
						multiplier += frac
					}
					consumed += fd
					isFractional = true
				}
			}
		} else if s[0] == '+' || s[0] == '-' {
			// Sign with no digits — not a relative item.
			return 0, false, nil
		}
	}

	// Skip whitespace between number and unit.
	rest := s[consumed:]
	ws := skipSpacesAndTabs(rest)

	// Extract the unit keyword.
	if ws+consumed >= len(s) || !isAlpha(s[consumed+ws]) {
		return 0, false, nil
	}

	word, wlen := extractWord(rest[ws:])
	if wlen == 0 {
		return 0, false, nil
	}

	entry, ok := unitTable[word]
	if !ok {
		// Not a unit keyword. If we consumed a number, we can't match.
		return 0, false, nil
	}

	// "mon" is ambiguous: it's both a unit (months) and a weekday (Monday).
	// In the relative matcher, "mon" should only match as a unit if preceded by a number.
	if word == "mon" && !hasNumber {
		return 0, false, nil
	}

	consumed += ws + wlen

	if isFractional {
		addFractionalToDelta(&sc.st.delta, entry.field, multiplier*float64(entry.scale))
	} else {
		addToDelta(&sc.st.delta, entry.field, int(multiplier)*entry.scale)
	}
	return consumed, true, nil
}

// ---------------------------------------------------------------------------
// matchDirectionOp — priority 10
// ---------------------------------------------------------------------------

// matchDirectionOp matches direction operators: ago, hence, before, after.
func (sc *scanner) matchDirectionOp() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 || !isAlpha(s[0]) {
		return 0, false, nil
	}

	word, wlen := extractWord(s)
	if wlen == 0 {
		return 0, false, nil
	}

	switch word {
	case "ago":
		if sc.parseDurationMode {
			// In duration mode, just negate the delta without setting an anchor.
			negateDelta(&sc.st.delta)
			return wlen, true, nil
		}
		// Negate delta, apply to ref as anchor, reset delta.
		t := applyDeltaToTime(sc.ref, sc.st.delta, -1)
		resetDelta(&sc.st.delta)
		if err := sc.setAnchor(t); err != nil {
			return 0, false, err
		}
		return wlen, true, nil

	case "hence":
		if sc.parseDurationMode {
			// In duration mode, delta stays positive — nothing to do.
			return wlen, true, nil
		}
		// Apply delta to ref as anchor (positive), reset delta.
		t := applyDeltaToTime(sc.ref, sc.st.delta, 1)
		resetDelta(&sc.st.delta)
		if err := sc.setAnchor(t); err != nil {
			return 0, false, err
		}
		return wlen, true, nil

	case "before":
		if sc.parseDurationMode {
			// In duration mode, "before" requires an anchor — error.
			return 0, false, fmt.Errorf("ParseDuration: \"before\" not allowed in duration expression (use Parse instead)")
		}
		if isDeltaZero(sc.st.delta) {
			return 0, false, fmt.Errorf("\"before\" without preceding delta")
		}
		// Push the current delta+direction onto the pending stack.
		// For the first direction op (direction==0), this captures the delta
		// that precedes "before". For chained ops, it captures the intermediate delta.
		sc.st.pendingOps = append(sc.st.pendingOps, pendingOp{d: sc.st.delta, dir: -1})
		resetDelta(&sc.st.delta)
		sc.st.direction = -1
		return wlen, true, nil

	case "after":
		if sc.parseDurationMode {
			// In duration mode, "after" requires an anchor — error.
			return 0, false, fmt.Errorf("ParseDuration: \"after\" not allowed in duration expression (use Parse instead)")
		}
		if isDeltaZero(sc.st.delta) {
			return 0, false, fmt.Errorf("\"after\" without preceding delta")
		}
		// Push the current delta+direction onto the pending stack.
		sc.st.pendingOps = append(sc.st.pendingOps, pendingOp{d: sc.st.delta, dir: 1})
		resetDelta(&sc.st.delta)
		sc.st.direction = 1
		return wlen, true, nil
	}

	return 0, false, nil
}

// ---------------------------------------------------------------------------
// matchPureNumber — priority 11
// ---------------------------------------------------------------------------

// matchPureNumber matches bare digit sequences:
//   - 8 digits with no prior anchor → yyyymmdd calendar date
//   - 4 digits with no prior timeOfDay → hhmm time
func (sc *scanner) matchPureNumber() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 || !isDigit(s[0]) {
		return 0, false, nil
	}

	nd := countDigits(s)

	// 8-digit number → yyyymmdd.
	if nd == 8 && !sc.st.anchorSet {
		year := parseDigits(s[0:4])
		month := parseDigits(s[4:6])
		day := parseDigits(s[6:8])
		if err := validateDate(year, month, day); err != nil {
			return 0, false, err
		}
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if err := sc.setAnchor(t); err != nil {
			return 0, false, err
		}
		return 8, true, nil
	}

	// 4-digit number → hhmm.
	if nd == 4 && !sc.st.todSet {
		hour := parseDigits(s[0:2])
		minute := parseDigits(s[2:4])
		if err := validateTime(hour, minute, 0); err != nil {
			return 0, false, err
		}
		if err := sc.setTimeOfDay(timeOfDay{hour: hour, minute: minute}); err != nil {
			return 0, false, err
		}
		return 4, true, nil
	}

	return 0, false, nil
}

// ---------------------------------------------------------------------------
// matchNoise — priority 12
// ---------------------------------------------------------------------------

// matchNoise matches noise tokens: "and", "at", bare comma, stray hyphen.
//
//nolint:unparam // error return is part of the matcher interface contract
func (sc *scanner) matchNoise() (int, bool, error) {
	s := sc.remaining()
	if len(s) == 0 {
		return 0, false, nil
	}

	// Bare comma.
	if s[0] == ',' {
		return 1, true, nil
	}

	// Stray hyphen not followed by digit.
	if s[0] == '-' {
		if len(s) == 1 || !isDigit(s[1]) {
			return 1, true, nil
		}
		return 0, false, nil
	}

	// "and" or "at" keywords.
	if isAlpha(s[0]) {
		word, wlen := extractWord(s)
		switch word {
		case "and", "at":
			return wlen, true, nil
		}
	}

	return 0, false, nil
}
