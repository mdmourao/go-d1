# go-d1

[![Go Reference](https://pkg.go.dev/badge/github.com/mdmourao/go-d1.svg)](https://pkg.go.dev/github.com/mdmourao/go-d1)
[![Go Report Card](https://goreportcard.com/badge/github.com/mdmourao/go-d1)](https://goreportcard.com/report/github.com/mdmourao/go-d1)
[![Status: Beta](https://img.shields.io/badge/status-beta-orange.svg)](#project-status)

[![Code Quality & Security Check](https://github.com/mdmourao/go-d1/actions/workflows/code-quality.yml/badge.svg)](https://github.com/mdmourao/go-d1/actions/workflows/code-quality.yml)
[![Test Suite](https://github.com/mdmourao/go-d1/actions/workflows/test.yml/badge.svg)](https://github.com/mdmourao/go-d1/actions/workflows/test.yml)

A Go [`database/sql`](https://pkg.go.dev/database/sql) driver for [Cloudflare D1](https://developers.cloudflare.com/d1/), the serverless SQLite database on Cloudflare's edge.

Part of the [orangegopher.dev](https://orangegopher.dev) project — because JavaScript and TypeScript are cool, but Go has a mascot. 

> **Project status: Beta.** APIs and behavior may change. Please open issues with feedback before relying on it in production.

## Overview

Cloudflare D1 is exposed via an HTTP API rather than a wire protocol, so it cannot be reached directly with a standard SQL driver. `go-d1` bridges that gap by talking to a small Cloudflare Worker that proxies SQL traffic to your D1 database, allowing you to use D1 from any Go program through the standard `database/sql` interface — including ORMs like [GORM](https://gorm.io).

```
┌────────────┐   HTTPS    ┌──────────────────┐   D1 binding   ┌──────────┐
│  Go app    │ ─────────► │  go-d1-proxy     │ ─────────────► │   D1     │
│  (go-d1)   │ ◄───────── │ (Cloudflare WK)  │ ◄───────────── │ database │
└────────────┘            └──────────────────┘                └──────────┘
```

## Requirements

This driver **requires** the companion Cloudflare Worker proxy to be deployed:

- **Proxy worker:** [github.com/mdmourao/go-d1-proxy](https://github.com/mdmourao/go-d1-proxy)

Follow the instructions in that repository to deploy the worker and bind it to your D1 database. The proxy URL becomes the DSN you pass to this driver.

## Installation

```bash
go get github.com/mdmourao/go-d1
```

## Quick start

```go
package main

import (
	"database/sql"
	"log"

	_ "github.com/mdmourao/go-d1"
)

func main() {
	db, err := sql.Open("god1", "https://your-worker.workers.dev")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("SELECT id, name FROM users WHERE active = ?", true)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Fatal(err)
		}
		log.Printf("%d: %s", id, name)
	}
}
```

### DSN

The DSN is the HTTPS URL of your deployed proxy worker:

```
https://your-worker.your-account.workers.dev
```

Only `http` and `https` schemes are supported.

## Configuration

Use `NewConnector` together with `sql.OpenDB` for typed configuration:

```go
import (
	"database/sql"
	"log/slog"
	"os"
	"time"

	god1 "github.com/mdmourao/go-d1"
)

connector, err := god1.NewConnector(
	"https://your-worker.workers.dev",
	god1.WithLogger(slog.New(slog.NewTextHandler(os.Stdout, nil))),
	god1.WithRequestTimeout(30*time.Second),
	god1.WithCloudflareAccess("client-id", "client-secret"),
	god1.WithSQLiteVersion("3.53.0"),
)
if err != nil {
	log.Fatal(err)
}

db := sql.OpenDB(connector)
```

Available options:

| Option | Description |
| --- | --- |
| `WithLogger(*slog.Logger)` | Structured logger for driver-level events. |
| `WithRequestTimeout(time.Duration)` | HTTP client timeout per request (default `15s`). |
| `WithCloudflareAccess(id, secret)` | Sends `CF-Access-Client-Id` / `CF-Access-Client-Secret` headers when the worker is protected by Cloudflare Access. |
| `WithSQLiteVersion(string)` | Reported SQLite version (default `3.53.0`). |

## Using with GORM

A full example using both raw SQL and GORM is available here:

- **Example client:** [github.com/mdmourao/go-d1-client](https://github.com/mdmourao/go-d1-client)

When using GORM you must **disable the default transaction wrapping**, since D1 over HTTP does not support interactive transactions:

```go
gormDB, err := gorm.Open(dialector, &gorm.Config{
    SkipDefaultTransaction: true, // required
})
```

## Limitations

D1's HTTP API and JSON value model impose a few constraints. Please read these carefully before adopting the driver.

- **No explicitly named parameters.** Only positional `?` parameters are supported. Named bindings such as `:name`, `@name` or `$1` will not work.
- **No interactive transactions.** D1 over HTTP cannot hold a transaction open across multiple round-trips. When using GORM, set `SkipDefaultTransaction: true` in `gorm.Config`.
- **No `BLOB` support.** Binary columns are not supported by this driver.
- **Dates are stored as RFC3339 strings.** `time.Time` values are serialized as RFC3339 text. Round-tripping non-RFC3339 timestamps from existing tables may require manual handling.
- **Integer range is limited to ±2^53 (`9007199254740992`).** Values flow through JSON numbers, so integers larger than JavaScript's safe integer range will lose precision. Store such values as `TEXT` if you need the full `int64` range.


## Project status

This is a **beta** project, actively developed as part of [orangegopher.dev](https://orangegopher.dev). Bug reports, feature requests and pull requests are welcome.

## Related projects

- [go-d1-proxy](https://github.com/mdmourao/go-d1-proxy) — the required Cloudflare Worker proxy.
- [go-d1-client](https://github.com/mdmourao/go-d1-client) — example app using `go-d1` with raw SQL and GORM.

## License

See [LICENSE](LICENSE.md).
