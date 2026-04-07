# dateparse

GNU `date --date` compatible timestamp parser for Go.

## Install

```sh
go get github.com/major0/dateparse
```

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/major0/dateparse"
)

func main() {
	t, err := dateparse.Parse("3 days ago", time.Now())
	if err != nil {
		panic(err)
	}
	fmt.Println(t)
}
```

## gdate CLI

A GNU `date`-compatible CLI tool for testing and interactive use.

```sh
go install github.com/major0/dateparse/cmd/gdate@latest

gdate --date "last monday 3pm"
gdate --date "2024-01-15" +"%Y-%m-%d"
gdate --date "3 days ago"
```

## Supported Formats

- RFC 3339 / ISO 8601: `2024-01-15T14:30:00Z`
- Calendar dates: `Jan 15 2024`, `15 Jan 2024`, `2024-01-15`, `1/15/2024`
- Time of day: `14:30`, `3pm`, `3:30 a.m.`
- Named references: `now`, `today`, `yesterday`, `tomorrow`
- Relative items: `3 days ago`, `2 hours hence`, `1 year 2 months`
- Day-of-week: `last monday`, `next friday`, `third tuesday`
- Epoch seconds: `@1705276800`
- Composable: `last monday 3pm`, `yesterday at 10am`, `2pm 3 days hence`
- Historical units: ghurry, scruple, mileway, microfortnight, nundine, lustre, and more

## License

MIT — Copyright (c) 2026 Mark Ferrell
