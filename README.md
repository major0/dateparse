# dateparse

A GNU `date --date` compatible timestamp parser and time calculator for Go. Parses human-readable date/time expressions into `time.Time` values using a single-pass scan+accumulate architecture with zero dependencies beyond the Go standard library.

## Features

- Full GNU `date --date` input format compatibility
- Direction operators: `before`, `after`, `ago`, `hence` with chaining support
- 30+ time units including historical and humorous units (ghurry, scruple, helek, microfortnight, etc.)
- Single-pass left-to-right scanner — no intermediate token list, no recursion
- Multi-field calendar-correct delta arithmetic via `time.AddDate` + `time.Add`
- Case-insensitive, whitespace-tolerant, parenthetical comment support
- Never panics on any input

## Parse

Resolves a date/time expression to an absolute `time.Time`.

```go
t, err := dateparse.Parse("3 days before Jan 15, 2025", time.Now())
t, err := dateparse.Parse("last monday 3pm", time.Now())
t, err := dateparse.Parse("7 hours before 2 weeks after July 13", time.Now())
```

See [docs/parse.md](docs/parse.md) for the full format reference.

## ParseDuration

Resolves a relative expression to a `dateparse.Duration` with separate calendar and sub-day fields.

```go
d, err := dateparse.ParseDuration("1 year 2 months 3 days")
result := d.Apply(time.Now())
```

See [docs/parse-duration.md](docs/parse-duration.md) for details.

## gdate CLI

A GNU `date`-compatible command-line tool.

```sh
gdate -d "3 days before Jan 15, 2025"
gdate -d "last monday 3pm" +"%Y-%m-%d %H:%M"
gdate -o "2 weeks and 3 days"
gdate -o "3 weeks and 2 days" +"%{days} days %{hours} hours"
# 23 days 0 hours
gdate -o "1 year 2 months 5 hours ago" +"%{years} years %{months} months %{hours} hours"
# -1 years -2 months -5 hours
gdate -o "28 days" +"%{fortnights} fortnights"
# 2 fortnights
```

See [docs/gdate.md](docs/gdate.md) for the full CLI reference.

## Install

Library:

```sh
go get github.com/major0/dateparse
```

CLI:

```sh
go install github.com/major0/dateparse/cmd/gdate@latest
```

## Documentation

- [Parse — format reference and examples](docs/parse.md)
- [ParseDuration — duration expressions and the Duration type](docs/parse-duration.md)
- [gdate — CLI reference and examples](docs/gdate.md)
- [Units — full conversion table with all supported time units](docs/units.md)

## License

MIT — Copyright (c) 2026 Mark Ferrell
