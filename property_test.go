package dateparse

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"time"

	"pgregory.net/rapid"
)

// propRef is the fixed reference time used across all property tests.
var propRef = time.Date(2024, 7, 15, 12, 0, 0, 0, time.UTC)

// ---------------------------------------------------------------------------
// Shared helpers and generators
// ---------------------------------------------------------------------------

// unambiguousUnits are unit keywords with simple, fixed-field semantics.
var unambiguousUnits = []string{
	"day", "days", "hour", "hours", "week", "weeks",
	"minute", "minutes", "second", "seconds",
}

// monthAbbrevs for calendar date format equivalence.
var monthAbbrevs = []string{
	"", "jan", "feb", "mar", "apr", "may", "jun",
	"jul", "aug", "sep", "oct", "nov", "dec",
}

// weekdayNamesList for day-of-week tests.
var weekdayNamesList = []string{
	"sunday", "monday", "tuesday", "wednesday",
	"thursday", "friday", "saturday",
}

// ordinalNamesList for day-of-week ordinal tests.
var ordinalNamesList = []string{
	"first", "second", "third", "fourth", "fifth",
}

// applyUnitDelta applies N units of the given keyword to t with the given sign.
func applyUnitDelta(t time.Time, unit string, n int, sign int) time.Time {
	entry := unitTable[unit]
	amount := n * entry.scale * sign
	switch entry.field {
	case fieldYears:
		return t.AddDate(amount, 0, 0)
	case fieldMonths:
		return t.AddDate(0, amount, 0)
	case fieldDays:
		return t.AddDate(0, 0, amount)
	case fieldHours:
		return t.Add(time.Duration(amount) * time.Hour)
	case fieldMinutes:
		return t.Add(time.Duration(amount) * time.Minute)
	case fieldSeconds:
		return t.Add(time.Duration(amount) * time.Second)
	case fieldNanos:
		return t.Add(time.Duration(amount))
	}
	return t
}

// randomCase applies random case transformation to each character.
func randomCase(s string, r *rand.Rand) string {
	var b strings.Builder
	for _, c := range s {
		if r.Intn(2) == 0 {
			b.WriteString(strings.ToUpper(string(c)))
		} else {
			b.WriteString(strings.ToLower(string(c)))
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 1: RFC 3339 Round-Trip
// ---------------------------------------------------------------------------

func TestProperty1_RFC3339RoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		year := rapid.IntRange(1970, 2100).Draw(t, "year")
		month := rapid.IntRange(1, 12).Draw(t, "month")
		maxDay := daysInMonth(year, month)
		day := rapid.IntRange(1, maxDay).Draw(t, "day")
		hour := rapid.IntRange(0, 23).Draw(t, "hour")
		minute := rapid.IntRange(0, 59).Draw(t, "minute")
		second := rapid.IntRange(0, 59).Draw(t, "second")
		nano := rapid.IntRange(0, 999999999).Draw(t, "nano")

		original := time.Date(year, time.Month(month), day, hour, minute, second, nano, time.UTC)
		formatted := original.Format(time.RFC3339Nano)

		got, err := Parse(formatted, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", formatted, err)
		}
		if !got.Equal(original) {
			t.Fatalf("Parse(%q) = %v, want %v", formatted, got, original)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 2: Epoch Round-Trip
// ---------------------------------------------------------------------------

func TestProperty2_EpochRoundTrip(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sec := rapid.Int64Range(-1e10, 1e10).Draw(t, "seconds")
		nano := rapid.IntRange(0, 999999999).Draw(t, "nano")

		var formatted string
		if nano > 0 {
			frac := fmt.Sprintf("%09d", nano)
			frac = strings.TrimRight(frac, "0")
			formatted = fmt.Sprintf("@%d.%s", sec, frac)
		} else {
			formatted = fmt.Sprintf("@%d", sec)
		}

		got, err := Parse(formatted, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", formatted, err)
		}
		if got.Unix() != sec {
			t.Fatalf("Parse(%q).Unix() = %d, want %d", formatted, got.Unix(), sec)
		}
		if got.Nanosecond() != nano {
			t.Fatalf("Parse(%q).Nanosecond() = %d, want %d", formatted, got.Nanosecond(), nano)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 3: Relative Direction Against Reference
// ---------------------------------------------------------------------------

func TestProperty3_RelativeDirectionAgainstReference(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 100).Draw(t, "n")
		unit := rapid.SampledFrom(unambiguousUnits).Draw(t, "unit")

		// "N unit ago" = ref - N*delta
		agoExpr := fmt.Sprintf("%d %s ago", n, unit)
		gotAgo, err := Parse(agoExpr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", agoExpr, err)
		}
		wantAgo := applyUnitDelta(propRef, unit, n, -1)
		if !gotAgo.Equal(wantAgo) {
			t.Fatalf("Parse(%q) = %v, want %v", agoExpr, gotAgo, wantAgo)
		}

		// "N unit hence" = ref + N*delta
		henceExpr := fmt.Sprintf("%d %s hence", n, unit)
		gotHence, err := Parse(henceExpr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", henceExpr, err)
		}
		wantHence := applyUnitDelta(propRef, unit, n, 1)
		if !gotHence.Equal(wantHence) {
			t.Fatalf("Parse(%q) = %v, want %v", henceExpr, gotHence, wantHence)
		}

		// "N unit" (bare) = ref + N*delta
		bareExpr := fmt.Sprintf("%d %s", n, unit)
		gotBare, err := Parse(bareExpr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", bareExpr, err)
		}
		if !gotBare.Equal(wantHence) {
			t.Fatalf("Parse(%q) = %v, want %v", bareExpr, gotBare, wantHence)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 4: Relative Direction Against Explicit Anchor
// ---------------------------------------------------------------------------

func TestProperty4_RelativeDirectionAgainstExplicitAnchor(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 100).Draw(t, "n")
		unit := rapid.SampledFrom([]string{"day", "days", "week", "weeks", "hour", "hours"}).Draw(t, "unit")
		aYear := rapid.IntRange(2000, 2030).Draw(t, "anchorYear")
		aMonth := rapid.IntRange(1, 12).Draw(t, "anchorMonth")
		maxDay := daysInMonth(aYear, aMonth)
		aDay := rapid.IntRange(1, maxDay).Draw(t, "anchorDay")
		anchor := fmt.Sprintf("%04d-%02d-%02d", aYear, aMonth, aDay)

		anchorTime, err := Parse(anchor, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", anchor, err)
		}

		// "N unit before YYYY-MM-DD"
		beforeExpr := fmt.Sprintf("%d %s before %s", n, unit, anchor)
		gotBefore, err := Parse(beforeExpr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", beforeExpr, err)
		}
		wantBefore := applyUnitDelta(anchorTime, unit, n, -1)
		if !gotBefore.Equal(wantBefore) {
			t.Fatalf("Parse(%q) = %v, want %v", beforeExpr, gotBefore, wantBefore)
		}

		// "N unit after YYYY-MM-DD"
		afterExpr := fmt.Sprintf("%d %s after %s", n, unit, anchor)
		gotAfter, err := Parse(afterExpr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", afterExpr, err)
		}
		wantAfter := applyUnitDelta(anchorTime, unit, n, 1)
		if !gotAfter.Equal(wantAfter) {
			t.Fatalf("Parse(%q) = %v, want %v", afterExpr, gotAfter, wantAfter)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 5: Multi-Relative Accumulation
// ---------------------------------------------------------------------------

func TestProperty5_MultiRelativeAccumulation(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Use units that map to distinct delta fields to avoid ambiguity.
		accumUnits := []string{"day", "hour", "minute"}
		count := rapid.IntRange(2, 4).Draw(t, "count")

		var parts []string
		expected := propRef
		for i := 0; i < count; i++ {
			n := rapid.IntRange(1, 50).Draw(t, fmt.Sprintf("n%d", i))
			u := rapid.SampledFrom(accumUnits).Draw(t, fmt.Sprintf("unit%d", i))
			parts = append(parts, fmt.Sprintf("%d %s", n, u))
			expected = applyUnitDelta(expected, u, n, 1)
		}

		expr := strings.Join(parts, " ")
		got, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}
		if !got.Equal(expected) {
			t.Fatalf("Parse(%q) = %v, want %v", expr, got, expected)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 6: Case Insensitivity
// ---------------------------------------------------------------------------

func TestProperty6_CaseInsensitivity(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 100).Draw(t, "n")
		unit := rapid.SampledFrom(unambiguousUnits).Draw(t, "unit")
		expr := fmt.Sprintf("%d %s ago", n, unit)

		seed := rapid.Int64().Draw(t, "seed")
		r := rand.New(rand.NewSource(seed)) //nolint:gosec // test-only, not security-sensitive
		transformed := randomCase(expr, r)

		gotOrig, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}
		gotTransformed, err := Parse(transformed, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", transformed, err)
		}
		if !gotOrig.Equal(gotTransformed) {
			t.Fatalf("Parse(%q) = %v, Parse(%q) = %v — should be equal",
				expr, gotOrig, transformed, gotTransformed)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 7: No-Panic Guarantee
// ---------------------------------------------------------------------------

func TestProperty7_NoPanicGuarantee(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		s := rapid.String().Draw(t, "input")
		// Just call Parse — if it panics, the test fails.
		_, _ = Parse(s, propRef)
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 8: Whitespace Invariance
// ---------------------------------------------------------------------------

func TestProperty8_WhitespaceInvariance(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 100).Draw(t, "n")
		unit := rapid.SampledFrom(unambiguousUnits).Draw(t, "unit")
		expr := fmt.Sprintf("%d %s ago", n, unit)

		// Add random leading/trailing whitespace.
		leading := strings.Repeat(" ", rapid.IntRange(0, 5).Draw(t, "leading"))
		trailing := strings.Repeat(" ", rapid.IntRange(0, 5).Draw(t, "trailing"))
		// Replace single spaces with multiple spaces.
		extraSpaces := strings.Repeat(" ", rapid.IntRange(2, 5).Draw(t, "extraSpaces"))
		padded := leading + strings.ReplaceAll(expr, " ", extraSpaces) + trailing

		gotOrig, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}
		gotPadded, err := Parse(padded, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", padded, err)
		}
		if !gotOrig.Equal(gotPadded) {
			t.Fatalf("Parse(%q) = %v, Parse(%q) = %v — should be equal",
				expr, gotOrig, padded, gotPadded)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 9: Calendar Date Format Equivalence
// ---------------------------------------------------------------------------

func TestProperty9_CalendarDateFormatEquivalence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		year := rapid.IntRange(2000, 2030).Draw(t, "year")
		month := rapid.IntRange(1, 12).Draw(t, "month")
		maxDay := daysInMonth(year, month)
		day := rapid.IntRange(1, maxDay).Draw(t, "day")

		// ISO format: YYYY-MM-DD
		iso := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
		// Literal: D Mon YYYY
		litDMY := fmt.Sprintf("%d %s %d", day, monthAbbrevs[month], year)
		// Literal: Mon D, YYYY
		litMDY := fmt.Sprintf("%s %d, %d", monthAbbrevs[month], day, year)

		gotISO, err := Parse(iso, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", iso, err)
		}
		gotDMY, err := Parse(litDMY, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", litDMY, err)
		}
		gotMDY, err := Parse(litMDY, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", litMDY, err)
		}

		want := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if !gotISO.Equal(want) {
			t.Fatalf("ISO Parse(%q) = %v, want %v", iso, gotISO, want)
		}
		if !gotDMY.Equal(want) {
			t.Fatalf("D Mon YYYY Parse(%q) = %v, want %v", litDMY, gotDMY, want)
		}
		if !gotMDY.Equal(want) {
			t.Fatalf("Mon D, YYYY Parse(%q) = %v, want %v", litMDY, gotMDY, want)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 10: Two-Digit Year Century Mapping
// ---------------------------------------------------------------------------

func TestProperty10_TwoDigitYearCenturyMapping(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		yy := rapid.IntRange(0, 99).Draw(t, "yy")
		expr := fmt.Sprintf("%02d-01-15", yy)

		got, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}

		var wantYear int
		if yy >= 69 {
			wantYear = 1900 + yy
		} else {
			wantYear = 2000 + yy
		}

		if got.Year() != wantYear {
			t.Fatalf("Parse(%q).Year() = %d, want %d", expr, got.Year(), wantYear)
		}
		if got.Month() != time.January || got.Day() != 15 {
			t.Fatalf("Parse(%q) = %v, want January 15", expr, got)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 11: Comment Invariance
// ---------------------------------------------------------------------------

func TestProperty11_CommentInvariance(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 100).Draw(t, "n")
		unit := rapid.SampledFrom(unambiguousUnits).Draw(t, "unit")
		expr := fmt.Sprintf("%d %s ago", n, unit)

		// Generate comment text without parentheses.
		commentChars := "abcdefghijklmnopqrstuvwxyz0123456789 "
		commentLen := rapid.IntRange(1, 20).Draw(t, "commentLen")
		var cb strings.Builder
		for i := 0; i < commentLen; i++ {
			idx := rapid.IntRange(0, len(commentChars)-1).Draw(t, fmt.Sprintf("c%d", i))
			cb.WriteByte(commentChars[idx])
		}
		comment := "(" + cb.String() + ")"

		// Insert comment at a whitespace boundary (before the expression).
		withComment := comment + " " + expr

		gotOrig, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}
		gotComment, err := Parse(withComment, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", withComment, err)
		}
		if !gotOrig.Equal(gotComment) {
			t.Fatalf("Parse(%q) = %v, Parse(%q) = %v — should be equal",
				expr, gotOrig, withComment, gotComment)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 12: "and" Noise Invariance
// ---------------------------------------------------------------------------

func TestProperty12_AndNoiseInvariance(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n1 := rapid.IntRange(1, 50).Draw(t, "n1")
		u1 := rapid.SampledFrom([]string{"day", "days"}).Draw(t, "u1")
		n2 := rapid.IntRange(1, 50).Draw(t, "n2")
		u2 := rapid.SampledFrom([]string{"hour", "hours"}).Draw(t, "u2")

		without := fmt.Sprintf("%d %s %d %s", n1, u1, n2, u2)
		with := fmt.Sprintf("%d %s and %d %s", n1, u1, n2, u2)

		gotWithout, err := Parse(without, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", without, err)
		}
		gotWith, err := Parse(with, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", with, err)
		}
		if !gotWithout.Equal(gotWith) {
			t.Fatalf("Parse(%q) = %v, Parse(%q) = %v — should be equal",
				without, gotWithout, with, gotWith)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 13: ParseDuration + Apply Consistency
// ---------------------------------------------------------------------------

func TestProperty13_ParseDurationApplyConsistency(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate a relative-only expression (no anchors).
		count := rapid.IntRange(1, 2).Draw(t, "count")
		accumUnits := []string{"day", "hour", "minute"}
		var parts []string
		for i := 0; i < count; i++ {
			n := rapid.IntRange(1, 50).Draw(t, fmt.Sprintf("n%d", i))
			u := rapid.SampledFrom(accumUnits).Draw(t, fmt.Sprintf("unit%d", i))
			parts = append(parts, fmt.Sprintf("%d %s", n, u))
		}
		expr := strings.Join(parts, " ")

		dur, err := ParseDuration(expr)
		if err != nil {
			t.Fatalf("ParseDuration(%q) error: %v", expr, err)
		}
		fromDuration := dur.Apply(propRef)

		fromParse, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}

		if !fromDuration.Equal(fromParse) {
			t.Fatalf("ParseDuration(%q).Apply(ref) = %v, Parse(%q, ref) = %v — should be equal",
				expr, fromDuration, expr, fromParse)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 14: Error Returns Zero Time
// ---------------------------------------------------------------------------

func TestProperty14_ErrorReturnsZeroTime(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// Generate known-invalid inputs.
		invalidInputs := []string{
			"2024-13-01",
			"2024-00-15",
			"2024-01-32",
			"xyzzy",
			"25:00",
			"2024-06-31",
			"gobbledygook",
			"!!!???",
		}
		input := rapid.SampledFrom(invalidInputs).Draw(t, "input")

		got, err := Parse(input, propRef)
		if err == nil {
			t.Fatalf("Parse(%q) expected error, got nil (result: %v)", input, got)
		}
		if !got.IsZero() {
			t.Fatalf("Parse(%q) returned non-zero time %v on error", input, got)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 15: Day-of-Week Nth Occurrence
// ---------------------------------------------------------------------------

func TestProperty15_DayOfWeekNthOccurrence(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		wdIdx := rapid.IntRange(0, 6).Draw(t, "weekday")
		ordIdx := rapid.IntRange(0, 4).Draw(t, "ordinal") // 0-4 → first through fifth
		wdName := weekdayNamesList[wdIdx]
		ordName := ordinalNamesList[ordIdx]
		targetWeekday := time.Weekday(wdIdx)
		ordinal := ordIdx + 1 // 1-5

		expr := fmt.Sprintf("%s %s", ordName, wdName)
		got, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}

		// Verify it falls on the correct weekday.
		if got.Weekday() != targetWeekday {
			t.Fatalf("Parse(%q).Weekday() = %v, want %v", expr, got.Weekday(), targetWeekday)
		}

		// Verify it's after ref (ordinals 1+ are future occurrences).
		if !got.After(propRef) {
			t.Fatalf("Parse(%q) = %v, expected after ref %v", expr, got, propRef)
		}

		// Verify it's the Nth occurrence after ref.
		// Count occurrences of this weekday between ref and got.
		refDate := time.Date(propRef.Year(), propRef.Month(), propRef.Day(), 0, 0, 0, 0, time.UTC)
		// Find the first occurrence of targetWeekday after ref.
		diff := int(targetWeekday) - int(refDate.Weekday())
		if diff <= 0 {
			diff += 7
		}
		firstOccurrence := refDate.AddDate(0, 0, diff)
		nthOccurrence := firstOccurrence.AddDate(0, 0, (ordinal-1)*7)

		if !got.Equal(nthOccurrence) {
			t.Fatalf("Parse(%q) = %v, want %v (ordinal %d)", expr, got, nthOccurrence, ordinal)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 16: Direction Chaining
// ---------------------------------------------------------------------------

func TestProperty16_DirectionChaining(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		d1 := rapid.IntRange(1, 10).Draw(t, "d1")
		d2 := rapid.IntRange(1, 10).Draw(t, "d2")
		aYear := rapid.IntRange(2000, 2030).Draw(t, "anchorYear")
		aMonth := rapid.IntRange(1, 12).Draw(t, "anchorMonth")
		maxDay := daysInMonth(aYear, aMonth)
		aDay := rapid.IntRange(1, maxDay).Draw(t, "anchorDay")
		anchor := fmt.Sprintf("%04d-%02d-%02d", aYear, aMonth, aDay)

		// "D1 days before D2 days after YYYY-MM-DD"
		expr := fmt.Sprintf("%d days before %d days after %s", d1, d2, anchor)
		got, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}

		anchorTime := time.Date(aYear, time.Month(aMonth), aDay, 0, 0, 0, 0, time.UTC)
		want := anchorTime.AddDate(0, 0, d2).AddDate(0, 0, -d1)
		if !got.Equal(want) {
			t.Fatalf("Parse(%q) = %v, want %v", expr, got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 17: Composite Anchor + Time-of-Day Override
// ---------------------------------------------------------------------------

func TestProperty17_CompositeAnchorTimeOfDayOverride(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		aYear := rapid.IntRange(2000, 2030).Draw(t, "anchorYear")
		aMonth := rapid.IntRange(1, 12).Draw(t, "anchorMonth")
		maxDay := daysInMonth(aYear, aMonth)
		aDay := rapid.IntRange(1, maxDay).Draw(t, "anchorDay")
		hour := rapid.IntRange(0, 23).Draw(t, "hour")
		minute := rapid.IntRange(0, 59).Draw(t, "minute")

		expr := fmt.Sprintf("%04d-%02d-%02d %02d:%02d", aYear, aMonth, aDay, hour, minute)
		got, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}

		want := time.Date(aYear, time.Month(aMonth), aDay, hour, minute, 0, 0, time.UTC)
		if !got.Equal(want) {
			t.Fatalf("Parse(%q) = %v, want %v", expr, got, want)
		}
	})
}

// ---------------------------------------------------------------------------
// Feature: gnu-dateparse, Property 18: Calendar-Field Delta via AddDate
// ---------------------------------------------------------------------------

func TestProperty18_CalendarFieldDeltaViaAddDate(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		n := rapid.IntRange(1, 12).Draw(t, "months")
		expr := fmt.Sprintf("%d months", n)

		got, err := Parse(expr, propRef)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", expr, err)
		}

		want := propRef.AddDate(0, n, 0)
		if !got.Equal(want) {
			t.Fatalf("Parse(%q) = %v, want %v", expr, got, want)
		}
	})
}
