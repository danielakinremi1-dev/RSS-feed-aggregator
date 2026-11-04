# Gator

[![CI](https://github.com/danielakinremi1-dev/RSS-feed-aggregator/actions/workflows/ci.yml/badge.svg)](https://github.com/danielakinremi1-dev/RSS-feed-aggregator/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A multi-user command-line RSS and Atom aggregator written in Go and backed by PostgreSQL. Its scraper fetches feeds concurrently, and you can run several copies of it against one database without them fetching the same feed twice.

![gator demo: register, add feeds, aggregate concurrently, browse](docs/demo.gif)

## Features

- **RSS 2.0 and Atom**, with lenient date parsing for the many ways real feeds format dates
- **Concurrent scraper**: a bounded worker pool that claims work with `FOR UPDATE SKIP LOCKED`, so multiple `agg` processes can share one database
- **Polite fetching**: conditional GET (`ETag` / `Last-Modified`), exponential backoff for failing feeds, and a 10 MB body cap
- **Graceful shutdown**: Ctrl-C lets in-flight fetches finish; press it again to force quit
- **Multi-user**: each user follows their own feeds and browses their own timeline

## Install

Download a prebuilt binary for Linux, macOS or Windows from the [Releases page](https://github.com/danielakinremi1-dev/RSS-feed-aggregator/releases), or install with Go:

```sh
go install github.com/danielakinremi1-dev/RSS-feed-aggregator/cmd/gator@latest
```

You need PostgreSQL 13 or later. Building from source needs Go 1.27+. The quick start below runs Postgres in Docker, so it also needs Docker.

## Quick start

The database migrations live in this repo, so clone it even if you installed a prebuilt binary:

```sh
git clone https://github.com/danielakinremi1-dev/RSS-feed-aggregator.git
cd RSS-feed-aggregator
make db-up                                               # Postgres on localhost:5432
go install github.com/pressly/goose/v3/cmd/goose@latest  # migration tool
make migrate-up
```

If port 5432 is taken, use another one: `GATOR_DB_PORT=5433 make db-up migrate-up`.

Point gator at the database in `~/.gatorconfig.json`:

```json
{ "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable" }
```

gator adds a `current_user_name` field to this file itself when you run `register` or `login`.

Then:

```sh
gator register alice
gator addfeed "Go Blog" https://go.dev/blog/feed.atom
gator agg --workers 4 1m      # fetch every minute; Ctrl-C to stop
gator browse                  # newest posts from the feeds you follow
```

## Commands

| Command | Description |
|---|---|
| `register <name>` | Create a user and log in as them |
| `login <name>` | Switch to an existing user |
| `users` | List all users |
| `addfeed <name> <url>` | Add a feed and follow it |
| `feeds` | List all feeds and their fetch status |
| `follow <url>` / `unfollow <url>` | Follow or unfollow an existing feed |
| `following` | List the feeds you follow |
| `browse [limit]` | Show your newest posts (default 10) |
| `agg [--workers N] <interval>` | Fetch due feeds every interval, N at a time (default 4) |
| `reset --force` | Delete **all** users, feeds, follows and posts |
| `version`, `help` | Print the version or the command list |

## How the scraper works

On every tick, `agg` claims a batch of feeds, scrapes them in parallel, and records each feed's outcome.

**Claiming work** takes a single SQL statement:

```sql
UPDATE feeds
SET last_fetched_at = now(), next_fetch_at = now() + interval '5 minutes'
WHERE id IN (
    SELECT id FROM feeds
    WHERE next_fetch_at IS NULL OR next_fetch_at <= now()
    ORDER BY last_fetched_at ASC NULLS FIRST
    LIMIT $1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;
```

- `FOR UPDATE SKIP LOCKED` makes concurrent claims skip rows that another process is holding, instead of waiting on them.
- The claim has to be **one statement**. A standalone `SELECT … FOR UPDATE` in autocommit mode releases its locks the moment it returns.
- Setting `next_fetch_at` acts as a **lease**. If another claim commits a row after this statement's snapshot, Postgres re-checks only the `WHERE` clause against the new row version, and the lease makes that check fail. Without the lease, the integration test caught two processes claiming the same feed. The lease also means a feed held by a process that crashed becomes due again after 5 minutes.

**Fetching** uses an `errgroup` with `SetLimit(workers)`. One failing feed never cancels the rest of its batch.

- Each request is conditional, so unchanged feeds return `304 Not Modified` and are skipped without being downloaded or parsed.
- When a feed fails, `consecutive_failures` and `last_error` are recorded, and the feed waits `interval × 2ⁿ` (capped at 6 hours) before it is retried. One success resets the count.

**Shutdown**: `signal.NotifyContext` cancels the context on SIGINT or SIGTERM.

- The ticker stops, but the batch already in flight runs on `context.WithoutCancel` with a 30-second timeout per feed, so its posts are saved.
- After the first signal, default signal handling is restored, so a second Ctrl-C exits immediately.

## Development

```sh
make test               # unit tests with the race detector
make test-integration   # also runs Postgres tests via testcontainers (needs Docker)
make lint               # gofmt, go vet, golangci-lint
make sqlc               # regenerate internal/database after editing sql/
make demo               # re-render docs/demo.gif (needs vhs) on a throwaway database
```

Every Makefile target that touches the database reads `DB_URL` from the environment.

```
cmd/gator/          entry point: signals, config, DB connection
internal/cli/       commands and the login middleware
internal/scraper/   claim → worker pool → record outcome, backoff
internal/feed/      HTTP fetch, RSS/Atom parsing, date parsing
internal/database/  sqlc-generated queries (from sql/queries)
sql/schema/         goose migrations
```

## License

[MIT](LICENSE)
