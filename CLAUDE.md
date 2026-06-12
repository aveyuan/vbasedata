# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`vbasedata` (module `github.com/aveyuan/vbasedata`) is a Go library of reusable infrastructure component wrappers. It has no `main` package and is not run directly — it is imported by other services. Everything lives in the single flat package `vbasedata` at the repo root.

## Commands

```bash
go build ./...        # build
go vet ./...          # static checks
go test ./...         # run all tests
go test -run TestNewGorm_SQLite -v   # run a single test
```

Note: `cat`/`echo`-style shell reads may be blocked by the permission mode — use the Read tool for files.

### Tests requiring external services

Some tests are gated on environment variables and `t.Skip()` when unset, so `go test ./...` passes with no infra:

- **MySQL / Postgres auto-create** (`gorm_auto_create_test.go`): set `VB_TEST_MYSQL_ADDR`/`VB_TEST_MYSQL_USER`/`VB_TEST_MYSQL_PASS` (or the `VB_TEST_PG_*` equivalents, plus optional `VB_TEST_PG_SSLMODE`/`VB_TEST_PG_TIMEZONE`). These create and drop a uniquely-named throwaway database.
- **Email** (`email_test.go`): hardcoded placeholder SMTP credentials — will fail unless edited with real values. Treat as a manual/example test, not CI.
- **SQLite** (`gorm_sqlite_test.go`) and **captcha** (`captcha_test.go`) run with no external dependencies (SQLite is pure-Go via `glebarez/sqlite`; no CGO).

## Architecture & conventions

Each file is one self-contained component exposing a `New*` constructor and a `*Config` struct. There is no central wiring, DI container, or registry — callers construct each piece independently.

Two recurring patterns to follow when adding or editing components:

1. **Connection-style constructors return `(client, cleanup func(), error)`.** `NewGorm` and `NewRedis` ping on startup and return a `func()` that closes the pool and logs. Callers are expected to `defer` the cleanup. Lightweight components (`NewCaptcha`, `NewPond`, `NewEmail`, `NewLruCache`, `NewIdgenerator`) return just the value (Pond exposes `Stop()` for shutdown).

2. **Config structs carry `yaml` + `json` tags and constructors fill in defaults in-place.** Zero-valued config fields are replaced with hardcoded defaults by mutating the passed `*Config` (e.g. `GormConfig.Conns`, `CaptchaConfig.Width`, `PondConfig.MaxWorkers`). Designed to be unmarshaled from a YAML/JSON config file.

Logging uses the stdlib `log/slog`. Constructors that take a `*slog.Logger` should tolerate `nil` (see `NewEmail`, which falls back to `slog.Default()`).

### Component map

- `gorm.go` — `NewGorm` is the most substantial component. Supports `mysql` / `pg`(`postgres`) / `sqlite` (default fallback) selected by `GormConfig.Type`. Key behavior: if the target database does not exist (MySQL error 1049 / PG SQLSTATE `3D000`), it connects to an admin DB and issues `CREATE DATABASE`, then reconnects. Postgres timezone is resolved from `TZ` / `/etc/timezone` / `/etc/localtime` when unset. Identifiers are manually quoted (`quoteMySQLIdent`/`quotePGIdent`).
- `redis.go` — `NewRedis` wraps `redis.NewUniversalClient`, so one config (`RedisConfig.Addr` as a string slice) transparently covers single-node, cluster, and sentinel modes.
- `lru.go` — `LruCache` over `hashicorp/golang-lru/v2/expirable`. Implements the `base64Captcha.Store` interface (`Set`/`Get`/`Verify`), so it can be passed directly as the captcha store (see `captcha_test.go`).
- `captcha.go` — `NewCaptcha` builds a `base64Captcha.DriverMath`; takes any `base64Captcha.Store` (LRU above, or a Redis-backed store).
- `pond.go` — `Pond` wraps `alitto/pond` worker pool with a panic handler that logs via slog.
- `email.go` — `Email` over `wneessen/go-mail`.
- `idgenerator.go` — snowflake-style IDs via `yitter/idgenerator-go`; `WorkerIdBitLength=4` caps the cluster at 16 nodes.
- `app.go` / `server.go` — plain config structs (`App`, `Http`, `Grpc`) with no logic, meant to be embedded in a service's top-level config.

## Notes for editing

- Comments and log messages are in Chinese — match the surrounding style.
- `gorm.go` logs successful connections via `slog.Error` (not `Info`); this is existing behavior, not necessarily a bug to silently "fix".
- The email subject is hardcoded to a placeholder string in `SendMsg` and ignores `Msg.Title` — keep this in mind if touching `email.go`.
