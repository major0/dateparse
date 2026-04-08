// Command gdate is a GNU date-compatible CLI tool.
package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/major0/dateparse"
	"github.com/major0/optargs/pflags"
)

func main() {
	dateStr := pflags.StringP("date", "d", "", "parse date string")
	offsetStr := pflags.StringP("offset", "o", "", "parse duration string")
	help := pflags.BoolP("help", "h", false, "show help")
	pflags.Parse()

	if *help {
		printUsage()
		os.Exit(0)
	}

	if *dateStr != "" && *offsetStr != "" {
		fmt.Fprintln(os.Stderr, "gdate: --date and --offset are mutually exclusive")
		os.Exit(1)
	}

	// Find +format in remaining args.
	var format string
	for _, arg := range pflags.Args() {
		if strings.HasPrefix(arg, "+") {
			format = arg[1:]
			break
		}
	}

	switch {
	case *offsetStr != "":
		runOffset(*offsetStr, format)
	case *dateStr != "":
		runDate(*dateStr, format)
	default:
		runNow(format)
	}
}

func runDate(s, format string) {
	t, err := dateparse.Parse(s, time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "gdate: %v\n", err)
		os.Exit(1)
	}
	if format != "" {
		if format == "@" {
			fmt.Println(t.Unix())
		} else {
			fmt.Println(formatDate(t, format))
		}
	} else {
		fmt.Println(t.Format(time.UnixDate))
	}
}

func runOffset(s, format string) {
	d, err := dateparse.ParseDuration(s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "gdate: %v\n", err)
		os.Exit(1)
	}
	if format != "" {
		// Check if format is a bare unit name (no % tokens).
		if !strings.Contains(format, "%") {
			name := strings.ToLower(format)
			if entry, ok := dateparse.LookupUnit(name); ok {
				val := convertToUnitExact(d, entry.Field, entry.Scale)
				fmt.Println(strconv.FormatFloat(val, 'f', -1, 64))
				return
			}
		}
		fmt.Println(formatDuration(d, format))
	} else {
		fmt.Println(totalSeconds(d))
	}
}

func runNow(format string) {
	t := time.Now()
	if format != "" {
		fmt.Println(formatDate(t, format))
	} else {
		fmt.Println(t.Format(time.UnixDate))
	}
}

func printUsage() {
	fmt.Println("Usage: gdate [OPTION]... [+FORMAT]")
	fmt.Println("Display the current time or parse a date/duration string.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -d, --date STRING    parse date STRING")
	fmt.Println("  -o, --offset STRING  parse duration STRING (print seconds)")
	fmt.Println("  -h, --help           display this help and exit")
	fmt.Println()
	fmt.Println("FORMAT controls the output (prefix with +).")
}

// totalNanos collapses all Duration fields into a single nanosecond count.
// Calendar fields use approximate conversions (365 days/year, 30 days/month).
func totalNanos(d dateparse.Duration) int64 {
	days := int64(d.Years)*365 + int64(d.Months)*30 + int64(d.Days)
	ns := days * 86400 * 1e9
	ns += int64(d.Hours) * 3600 * 1e9
	ns += int64(d.Minutes) * 60 * 1e9
	ns += int64(d.Seconds) * 1e9
	ns += int64(d.Nanos)
	return ns
}

// totalSeconds computes the signed total seconds from a Duration.
func totalSeconds(d dateparse.Duration) int64 {
	return totalNanos(d) / 1e9
}

// convertToUnitExact converts a Duration to the target unit using field-aware
// arithmetic. Within the same field group (calendar: years/months, clock:
// days/hours/minutes/seconds/nanos) the conversion is exact. Cross-group
// conversion uses approximations (30 days/month, 365 days/year) only when
// the other group has nonzero fields.
func convertToUnitExact(d dateparse.Duration, field dateparse.UnitField, scale int) float64 {
	if scale == 0 {
		return 0
	}

	// Calendar group: total months.
	calMonths := float64(int64(d.Years)*12 + int64(d.Months))

	// Clock group: total nanoseconds.
	clockNs := float64(int64(d.Days))*86400e9 +
		float64(int64(d.Hours))*3600e9 +
		float64(int64(d.Minutes))*60e9 +
		float64(int64(d.Seconds))*1e9 +
		float64(int64(d.Nanos))

	switch field {
	case dateparse.FieldYears, dateparse.FieldMonths:
		// Target is calendar. Start with exact calendar months.
		totalMonths := calMonths
		// Cross-group: add clock contribution only if nonzero.
		if clockNs != 0 {
			totalMonths += clockNs / (30 * 86400e9) // approximate
		}
		// Convert months to target unit.
		var monthsPerUnit float64
		if field == dateparse.FieldYears {
			monthsPerUnit = float64(scale) * 12
		} else {
			monthsPerUnit = float64(scale)
		}
		return totalMonths / monthsPerUnit

	default:
		// Target is clock (days, hours, minutes, seconds, nanos).
		totalNs := clockNs
		// Cross-group: add calendar contribution only if nonzero.
		if calMonths != 0 {
			totalNs += calMonths * 30 * 86400e9 // approximate
		}
		nsPerUnit := float64(nanosPerField[field]) * float64(scale)
		if nsPerUnit == 0 {
			return 0
		}
		return totalNs / nsPerUnit
	}
}

// strftimeToGo maps GNU strftime tokens to Go time.Format reference time components.
var strftimeToGo = map[byte]string{
	'Y': "2006",
	'm': "01",
	'd': "02",
	'H': "15",
	'M': "04",
	'S': "05",
	'A': "Monday",
	'B': "January",
	'a': "Mon",
	'b': "Jan",
	'p': "PM",
	'Z': "MST",
	'z': "-0700",
}

// formatDate translates a GNU strftime format string to Go time.Format output.
func formatDate(t time.Time, format string) string {
	var b strings.Builder
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b.WriteByte(format[i])
			continue
		}
		i++
		if i >= len(format) {
			b.WriteByte('%')
			break
		}
		ch := format[i]
		if ch == '%' {
			b.WriteByte('%')
			continue
		}
		if ch == 's' {
			b.WriteString(strconv.FormatInt(t.Unix(), 10))
			continue
		}
		if goRef, ok := strftimeToGo[ch]; ok {
			b.WriteString(t.Format(goRef))
			continue
		}
		// Unknown token: pass through literally.
		b.WriteByte('%')
		b.WriteByte(ch)
	}
	return b.String()
}

// durationUnitTable maps unit names to {field index, scale} for Duration formatting.
// Field indices: 0=Years, 1=Months, 2=Days, 3=Hours, 4=Minutes, 5=Seconds, 6=Nanos.
var durationFieldMap = map[byte]int{
	'Y': 0, // Years
	'M': 1, // Months
	'D': 2, // Days
	'h': 3, // Hours
	'm': 4, // Minutes
	's': 5, // Seconds
	'n': 6, // Nanos
}

func durationField(d dateparse.Duration, idx int) int {
	switch idx {
	case 0:
		return d.Years
	case 1:
		return d.Months
	case 2:
		return d.Days
	case 3:
		return d.Hours
	case 4:
		return d.Minutes
	case 5:
		return d.Seconds
	case 6:
		return d.Nanos
	}
	return 0
}

// fieldPriority defines the reduction order: largest unit first.
// Lower number = reduced first.
var fieldPriority = map[dateparse.UnitField]int{
	dateparse.FieldYears:   0,
	dateparse.FieldMonths:  1,
	dateparse.FieldDays:    2,
	dateparse.FieldHours:   3,
	dateparse.FieldMinutes: 4,
	dateparse.FieldSeconds: 5,
	dateparse.FieldNanos:   6,
}

// nanosPerField gives the approximate nanoseconds per one unit of each field.
// Used for remainder reduction in composite format strings.
var nanosPerField = map[dateparse.UnitField]int64{
	dateparse.FieldYears:   365 * 86400 * 1e9,
	dateparse.FieldMonths:  30 * 86400 * 1e9,
	dateparse.FieldDays:    86400 * 1e9,
	dateparse.FieldHours:   3600 * 1e9,
	dateparse.FieldMinutes: 60 * 1e9,
	dateparse.FieldSeconds: 1e9,
	dateparse.FieldNanos:   1,
}

// tokenInfo holds a parsed %{name} or %X token from the format string.
type tokenInfo struct {
	name     string // unit name (for %{name}) or short key
	field    dateparse.UnitField
	scale    int
	priority int
}

// formatPlaceholder holds a parsed %{name} token position and metadata.
type formatPlaceholder struct {
	start, end int // byte range in format string (including %{ and })
	info       tokenInfo
}

// formatDuration expands format tokens for a Duration.
//
// For composite formats with multiple %{name} tokens, fields are reduced
// largest-to-smallest: each token gets the quotient, and the remainder
// carries to smaller fields. E.g. "3 days 4 hours" with format
// "%{days} days %{hours} hours" → "3 days 4 hours" (not "3 days 76 hours").
//
// Short tokens (%Y, %M, %D, %h, %m, %s, %n) read the raw field value.
// Long tokens (%{name}) participate in remainder reduction.
func formatDuration(d dateparse.Duration, format string) string {
	// Pass 1: collect all %{name} tokens and their positions.
	var placeholders []formatPlaceholder
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		if i+1 >= len(format) {
			break
		}
		if format[i+1] == '{' {
			end := strings.IndexByte(format[i+2:], '}')
			if end < 0 {
				break
			}
			name := strings.ToLower(format[i+2 : i+2+end])
			entry, ok := dateparse.LookupUnit(name)
			if ok {
				pri := fieldPriority[entry.Field]
				placeholders = append(placeholders, formatPlaceholder{
					start: i,
					end:   i + 2 + end + 1, // past the '}'
					info: tokenInfo{
						name:     name,
						field:    entry.Field,
						scale:    entry.Scale,
						priority: pri,
					},
				})
			}
			i = i + 2 + end // skip past }
		}
	}

	// Pass 2: sort placeholders by priority (largest field first) and reduce.
	// All units get integer values except the smallest (last in priority order),
	// which gets a decimal remainder.
	resolvedStr := make(map[int]string) // placeholder index → formatted value
	if len(placeholders) > 0 {
		// Sort indices by priority.
		order := make([]int, len(placeholders))
		for i := range order {
			order[i] = i
		}
		sortByPriority(order, placeholders)

		remaining := totalNanos(d)
		sign := int64(1)
		if remaining < 0 {
			sign = -1
			remaining = -remaining
		}
		for oi, idx := range order {
			p := placeholders[idx]
			nsPerUnit := nanosPerField[p.info.field] * int64(p.info.scale)
			if nsPerUnit <= 0 {
				resolvedStr[idx] = "0"
				continue
			}
			isLast := oi == len(order)-1
			if isLast {
				// Smallest unit gets decimal remainder.
				val := float64(sign) * float64(remaining) / float64(nsPerUnit)
				formatted := strconv.FormatFloat(val, 'f', -1, 64)
				resolvedStr[idx] = formatted
			} else {
				count := remaining / nsPerUnit
				remaining -= count * nsPerUnit
				resolvedStr[idx] = strconv.FormatInt(sign*count, 10)
			}
		}
	}

	// Pass 3: expand the format string.
	var b strings.Builder
	phIdx := 0
	for i := 0; i < len(format); i++ {
		// Check if we're at a placeholder position.
		if phIdx < len(placeholders) && i == placeholders[phIdx].start {
			b.WriteString(resolvedStr[phIdx])
			i = placeholders[phIdx].end - 1 // -1 because loop increments
			phIdx++
			continue
		}
		if format[i] != '%' {
			b.WriteByte(format[i])
			continue
		}
		i++
		if i >= len(format) {
			b.WriteByte('%')
			break
		}
		ch := format[i]
		if ch == '%' {
			b.WriteByte('%')
			continue
		}
		if ch == '{' {
			// Already handled by placeholder pass — skip to }.
			end := strings.IndexByte(format[i:], '}')
			if end >= 0 {
				i += end
			}
			continue
		}
		if idx, ok := durationFieldMap[ch]; ok {
			b.WriteString(strconv.Itoa(durationField(d, idx)))
			continue
		}
		b.WriteByte('%')
		b.WriteByte(ch)
	}
	return b.String()
}

// sortByPriority sorts indices by placeholder priority (insertion sort, small N).
func sortByPriority(order []int, phs []formatPlaceholder) {
	for i := 1; i < len(order); i++ {
		key := order[i]
		j := i - 1
		for j >= 0 && phs[order[j]].info.priority > phs[key].info.priority {
			order[j+1] = order[j]
			j--
		}
		order[j+1] = key
	}
}

// lookupDurationUnit resolves a unit name and converts the entire Duration
// into that unit using field-aware exact arithmetic.
func lookupDurationUnit(d dateparse.Duration, name string) float64 {
	name = strings.ToLower(name)
	entry, ok := dateparse.LookupUnit(name)
	if !ok {
		return 0
	}
	return convertToUnitExact(d, entry.Field, entry.Scale)
}
