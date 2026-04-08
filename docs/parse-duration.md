# ParseDuration

```go
func ParseDuration(s string) (Duration, error)
```

Resolves a relative expression to a `dateparse.Duration` — a multi-field offset with separate calendar and sub-day components.

## Duration Type

```go
type Duration struct {
    Years, Months, Days         int
    Hours, Minutes, Seconds     int
    Nanos                       int
}

func (d Duration) Apply(t time.Time) time.Time
```

`Apply` applies calendar fields via `time.AddDate(Years, Months, Days)` then sub-day fields via `time.Add`. This preserves human-intuitive calendar arithmetic — "1 month" from Jan 31 gives Feb 28, not a fixed number of seconds.

## Examples

```go
d, _ := dateparse.ParseDuration("3 days")
// Duration{Days: 3}

d, _ = dateparse.ParseDuration("1 year 2 months")
// Duration{Years: 1, Months: 2}

d, _ = dateparse.ParseDuration("5 hours ago")
// Duration{Hours: -5}

d, _ = dateparse.ParseDuration("2 weeks and 3 days")
// Duration{Days: 17}

d, _ = dateparse.ParseDuration("1 fortnight")
// Duration{Days: 14}

d, _ = dateparse.ParseDuration("-3 days")
// Duration{Days: -3}
```

## Direction Operators

`ago` and `before` negate all fields:

```
5 hours ago           → Duration{Hours: -5}
3 days before         → Duration{Days: -3}
```

`hence` and `after` keep fields positive:

```
5 hours hence         → Duration{Hours: 5}
3 days after          → Duration{Days: 3}
```

## Restrictions

ParseDuration errors if the expression contains an anchor (calendar date, epoch, named reference, or day-of-week). It's for pure delta expressions only.

```go
// These work:
ParseDuration("3 days")
ParseDuration("1 year 2 months ago")

// These error:
ParseDuration("3 days before Jan 15")   // contains anchor
ParseDuration("yesterday")              // named reference is an anchor
ParseDuration("@0")                     // epoch is an anchor
```

## Consistency with Parse

For any relative-only expression `s` and reference time `ref`:

```go
d, _ := dateparse.ParseDuration(s)
d.Apply(ref) == dateparse.Parse(s, ref)  // always true
```
