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
		fmt.Println(formatDate(t, format))
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

// totalSeconds computes the signed total seconds from a Duration.
func totalSeconds(d dateparse.Duration) int64 {
	// Calendar fields: approximate using 365 days/year, 30 days/month.
	days := int64(d.Years)*365 + int64(d.Months)*30 + int64(d.Days)
	secs := days*86400 + int64(d.Hours)*3600 + int64(d.Minutes)*60 + int64(d.Seconds)
	secs += int64(d.Nanos) / 1e9
	return secs
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

// formatDuration expands format tokens for a Duration.
// Short tokens: %Y, %M, %D, %h, %m, %s, %n.
// Long tokens: %{name} where name is a unit from unitTable.
func formatDuration(d dateparse.Duration, format string) string {
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
		if ch == '{' {
			// Long form: %{name}
			end := strings.IndexByte(format[i:], '}')
			if end < 0 {
				b.WriteString("%{")
				continue
			}
			name := format[i+1 : i+end]
			i += end
			b.WriteString(strconv.Itoa(lookupDurationUnit(d, name)))
			continue
		}
		if idx, ok := durationFieldMap[ch]; ok {
			b.WriteString(strconv.Itoa(durationField(d, idx)))
			continue
		}
		// Unknown token: pass through.
		b.WriteByte('%')
		b.WriteByte(ch)
	}
	return b.String()
}

// lookupDurationUnit resolves a unit name to the corresponding Duration field
// divided by the unit's scale factor, using the dateparse unitTable.
func lookupDurationUnit(d dateparse.Duration, name string) int {
	name = strings.ToLower(name)
	entry, ok := dateparse.LookupUnit(name)
	if !ok {
		return 0
	}
	var fieldVal int
	switch entry.Field {
	case dateparse.FieldYears:
		fieldVal = d.Years
	case dateparse.FieldMonths:
		fieldVal = d.Months
	case dateparse.FieldDays:
		fieldVal = d.Days
	case dateparse.FieldHours:
		fieldVal = d.Hours
	case dateparse.FieldMinutes:
		fieldVal = d.Minutes
	case dateparse.FieldSeconds:
		fieldVal = d.Seconds
	case dateparse.FieldNanos:
		fieldVal = d.Nanos
	}
	if entry.Scale == 0 {
		return 0
	}
	return fieldVal / entry.Scale
}
