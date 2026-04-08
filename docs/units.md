# Supported Time Units

All units accept singular, plural, and abbreviated forms. Matching is case-insensitive. Each unit accumulates into a specific delta field for calendar-correct arithmetic.

## Calendar Fields (applied via `time.AddDate`)

| Unit | Aliases | Field × Scale |
|------|---------|---------------|
| millennium | millennia, millenniums | years × 1000 |
| saeculum | saecula | years × 100 |
| century | centuries | years × 100 |
| indiction | indictions | years × 15 |
| decade | decades | years × 10 |
| lustre | lustres, lustrum, lustra | years × 5 |
| olympiad | olympiads, quadrennium, quadrennia | years × 4 |
| biennial | biennials | years × 2 |
| annus | anni | years × 1 |
| year | years, yr, yrs | years × 1 |
| semester | semesters, friedman, friedmans | months × 6 |
| season | seasons | months × 3 |
| trimester | trimesters | months × 3 |
| month | months, mo, mos, mon | months × 1 |
| fortnight | fortnights | days × 14 |
| nundine | nundines | days × 8 |
| week | weeks, wk, wks | days × 7 |
| nychthemeron | nychthemera | days × 1 |
| day | days | days × 1 |

## Sub-Day Fields (applied via `time.Add`)

| Unit | Aliases | Field × Scale |
|------|---------|---------------|
| pahar | pahars | hours × 3 |
| hour | hours, hr, hrs | hours × 1 |
| ghurry | ghurries | minutes × 24 |
| mileway | mileways | minutes × 20 |
| minute | minutes, min, mins | minutes × 1 |
| moment | moments | seconds × 90 |
| scruple | scruples | seconds × 60 |
| second | seconds, sec, secs | seconds × 1 |
| helek | halakim, helakim | nanos × 3,333,333,333 |
| microfortnight | microfortnights | nanos × 1,209,600,000 |
| jiffy | jiffies | nanos × 10,000,000 |
| shake | shakes | nanos × 10 |

## Resolution Order

When a delta is applied to an anchor time:

1. Calendar fields first: `time.AddDate(years, months, days)`
2. Sub-day fields second: `time.Add(hours + minutes + seconds + nanos)`

This ensures human-intuitive calendar arithmetic. For example, "1 month" from January 31 gives February 28 (or 29 in a leap year), not a fixed number of seconds.

## Examples

```
3 days ago                    days field: -3
1 year 2 months               years: 1, months: 2
2 fortnights                  days: 28
1 millennium                  years: 1000
3 ghurries                    minutes: 72
5 shakes                      nanos: 50
1 helek                       nanos: 3,333,333,333 (~3.33 seconds)
```
