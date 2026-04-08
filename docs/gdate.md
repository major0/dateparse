# gdate

GNU `date`-compatible CLI tool powered by the `dateparse` library.

## Install

```sh
go install github.com/major0/dateparse/cmd/gdate@latest
```

## Usage

```
gdate [OPTIONS] [+FORMAT]
```

## Options

| Flag | Short | Description |
|------|-------|-------------|
| `--date <string>` | `-d` | Parse a date/time expression and print the result |
| `--offset <string>` | `-o` | Parse a duration expression and print the offset |
| `--help` | | Print usage information |

`--date` and `--offset` are mutually exclusive.

## Date Mode (--date / -d)

Parse a date/time expression relative to now and print the resolved time.

```sh
# Absolute dates
gdate -d "2024-01-15"
gdate -d "Jan 15, 2024"
gdate -d "2024-01-15T14:30:00Z"
gdate -d "@1705276800"

# Named references
gdate -d "now"
gdate -d "yesterday"
gdate -d "tomorrow"

# Relative expressions
gdate -d "3 days ago"
gdate -d "2 weeks hence"
gdate -d "1 year 2 months 3 days ago"

# Direction operators
gdate -d "3 days before Jan 15, 2025"
gdate -d "2 weeks after July 13"
gdate -d "7 hours before 2 weeks after July 13"

# Day-of-week
gdate -d "last monday"
gdate -d "next friday"
gdate -d "third tuesday"

# Composite
gdate -d "last monday 3pm"
gdate -d "yesterday at 10am"
gdate -d "2pm 3 days hence"
```

## Offset Mode (--offset / -o)

Parse a duration expression and print the offset. Default output is seconds.

```sh
gdate -o "3 days"
# 259200

gdate -o "1 year 2 months"
# (prints total seconds equivalent)

gdate -o "5 hours ago"
# -18000

gdate -o "2 weeks and 3 days"
# 1468800
```

## Output Format (+FORMAT)

Use `+FORMAT` to control output. In date mode, this uses GNU strftime-compatible format specifiers translated to Go's `time.Format` layout.

```sh
gdate -d "2024-01-15" +"%Y-%m-%d"
# 2024-01-15

gdate -d "yesterday" +"%A, %B %d %Y"
# Saturday, June 14 2024

gdate -d "now" +"%s"
# 1718452800
```

In offset mode, `+FORMAT` selects the output unit or composite format:

```sh
gdate -o "3 days" +seconds     # 259200 (default)
gdate -o "3 days" +minutes     # 4320
gdate -o "3 days" +hours       # 72
gdate -o "3 days" +days        # 3
```

## No Arguments

With no arguments, prints the current time in the default format (equivalent to GNU `date` with no args).

```sh
gdate
# Wed Apr  8 14:30:00 UTC 2026
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Parse error or invalid arguments (error printed to stderr) |

## Examples

```sh
# When is 3 weeks from now?
gdate -d "3 weeks hence"

# What date was 100 days ago?
gdate -d "100 days ago" +"%Y-%m-%d"

# How many seconds in a fortnight?
gdate -o "1 fortnight"

# What's 2 months before Christmas 2025?
gdate -d "2 months before Dec 25, 2025" +"%B %d, %Y"

# Combine relative and absolute
gdate -d "3pm 2 days after next monday"
```
