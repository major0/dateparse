# Parse

```go
func Parse(s string, ref time.Time) (time.Time, error)
```

Resolves a human-readable date/time expression to an absolute `time.Time`. The `ref` parameter is the reference time used as the implicit anchor when no explicit date is given.

## Absolute Formats

### RFC 3339 / ISO 8601

```
2024-01-15T14:30:00Z
2024-01-15T20:02:00-05:00
2024-01-15 14:30:00Z          (space instead of T)
2024-01-15T14:30:00.123Z      (fractional seconds)
```

### Calendar Dates

```
2024-01-15                    ISO format
99-01-15                      two-digit year (69-99 → 19xx, 00-68 → 20xx)
1/15/2024                     US format
1/15                          US format, ref year
15 Jan 2024                   day month year
Jan 15 2024                   month day year
Jan 15, 2024                  month day, year (comma accepted)
15-Jan-2024                   day-month-year with hyphens
15 Jan                        day month, ref year
```

Month names: full (January), 3-letter (Jan), with period (Jan.), and Sept/Sept.

### Epoch Seconds

```
@1705276800                   Unix epoch
@-86400                       negative (pre-1970)
@1078100502.5                 fractional seconds
```

### Named References

```
now                           reference time
today                         reference time
yesterday                     ref minus 1 day
tomorrow                      ref plus 1 day
```

### Day-of-Week

```
monday                        next Monday after ref
last friday                   most recent Friday before ref
this saturday                 Saturday of current week
next tuesday                  first Tuesday after current week
third monday                  3rd Monday after ref
```

Ordinals: last, this, next, first through twelfth.
Day names: full (Monday), 3-letter (Mon), with period (Mon.), and Tues, Wednes, Thur, Thurs.

## Relative Expressions

```
3 days ago                    ref minus 3 days
2 hours hence                 ref plus 2 hours
1 year 2 months 3 days        accumulated offset applied to ref
fortnight                     implicit multiplier of 1
+3 days                       signed multiplier
-2 hours                      negative offset
```

## Direction Operators

```
3 days before Jan 15, 2025    Jan 15 minus 3 days → Jan 12
2 weeks after July 13         July 13 plus 14 days → July 27
3 days ago                    sugar for "3 days before now"
2 weeks hence                 sugar for "2 weeks after now"
```

### Chaining

Direction operators chain left-to-right:

```
7 hours before 2 weeks after July 13
→ July 13 + 2 weeks = July 27
→ July 27 - 7 hours
```

## Composite Expressions

Items can be combined. Time-of-day overrides the time component of the resolved anchor.

```
last monday 3pm               last Monday at 15:00
yesterday at 10am             yesterday at 10:00
20 Jul 2020 14:30             specific date and time
2pm 3 days hence              3 days from now at 14:00
```

Noise tokens `at` and `and` are ignored:
```
yesterday at 3pm              same as "yesterday 3pm"
2 weeks and 3 days ago        same as "2 weeks 3 days ago"
```

Parenthetical comments are ignored:
```
Jan 15 2024 (deadline)        same as "Jan 15 2024"
```

## Error Handling

- Unrecognized input → error with position
- Invalid values (month 13, hour 25) → descriptive error
- Duplicate anchor without direction operator → conflict error
- `before`/`after` without preceding delta → error
- am/pm combined with timezone correction → error
- Returns `(time.Time{}, err)` on failure — never panics
