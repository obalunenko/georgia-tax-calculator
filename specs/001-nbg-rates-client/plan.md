# Implementation Plan: NBG Rates Client with Caching

**Feature Directory**: `specs/001-nbg-rates-client` | **Branch**: N/A (migrated)

**Date**: 2026-11-14 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/001-nbg-rates-client/spec.md`

**Note**: Reverse-engineered from existing code (`status: migrated`). It
describes what is implemented, not a proposal.

## Summary

`pkg/nbggovge` is a small HTTP client for the NBG currency rates API. Requests
are configured through functional options (date, currency codes). Responses
are normalized: a synthetic GEL entry is added when relevant and currencies are
sorted by code. `CachedClient` decorates any `Client` with an in-memory TTL
cache. Currency codes are generated into constants by a standalone generator,
and `mock.NewClient` serves fixtures through the real client code path.

## Technical Context

**Language/Version**: Go 1.27.3, module
`github.com/obalunenko/georgia-tax-calculator`.

**Primary Dependencies**: Standard library (`net/http`, `encoding/json`,
`hash/crc32`, `sync`); `github.com/obalunenko/logger` for body-close errors;
generator also uses `golang.org/x/tools/imports`, `html/template`, and
`github.com/obalunenko/version`.

**Storage**: In-memory map in `CachedClient`; no persistence.

**Testing**: Go `testing` with testify; fixtures in `pkg/nbggovge/testdata/`;
counting fake `Client` in `cache_test.go`; `doerMock` `HTTPClient` in
`client_test.go`. Real-API tests are gated by `INTEGRATION_TESTS=true`.

**Target Platform**: Any platform the CLI and bot build for (CGO disabled).

**Project Type**: Reusable library inside `pkg/`.

**Performance Goals**: Not specified. The cache removes repeated requests for
the same date and currency set within the TTL (1 hour by default).

**Constraints**: No dependency on `internal/` (constitution V). Callers reach
the client through `internal/converter`, never from `cmd/` (constitution I).

**Scale/Scope**: One endpoint; 44 currency codes; ~420 lines of library code,
~400 lines of generated code and generator, ~700 lines of tests.

## Constitution Check

- [x] I. Interfaces do not import `pkg/nbggovge` directly; only
      `pkg/nbggovge/currencies` constants are used in `cmd/`.
- [x] II. N/A (library; no interface flows).
- [x] III. N/A (no money arithmetic; rates are returned as published).
- [x] IV. This package is the `nbggovge.Client`; production uses
      `NewCached`; the real-API test is gated by `INTEGRATION_TESTS`.
- [x] V. No imports of `internal/` packages of the module.
- [ ] VI. Cache, constructors, and response parsing are tested;
      `client.Rates` is only tested under `INTEGRATION_TESTS` (see tasks Gaps).
- [x] VII. `codes.gen.go` is generated and regenerated manually with the
      currencies generator, as constitution VII describes (no automation; see
      tasks Gaps).
- [x] Known deviations are not extended.

## Project Structure

### Documentation (this feature)

```text
specs/001-nbg-rates-client/
├── spec.md
├── plan.md
└── tasks.md
```

### Source Code (repository root)

```text
pkg/nbggovge/
├── client.go              # Client, HTTPClient, constructors, Rates, GEL, sorting
├── response.go            # Rates, Currency, CurrencyByCode, ErrCodeNotFound, parsing
├── cache.go               # CachedClient, TTL, key generation, cache management
├── README_CACHE.md        # Cache usage documentation
├── option/option.go       # WithDate, WithCurrency
├── internal/              # RatesParams, query parameter names, date layout
├── currencies/
│   ├── codes.gen.go       # Generated constants and All()
│   └── generator/         # main.go + codes.go.tmpl
├── mock/mock.go           # NewClient over fixtures
├── *_test.go              # cache, constructors, client (integration), response
└── testdata/              # 2024-02-10 fixtures and empty.json
```

**Structure Decision**: Single library package with option and internal
subpackages; the cache is a decorator in the same package so constructors can
return `*CachedClient`.

**Validation**: `go test -race ./pkg/nbggovge/...`;
`make test-integration` for live API checks; `make vet`, `make lint-pipeline`.

## Technical Decisions

1. **Functional options** (`option.RatesOption` applying to
   `internal.RatesParams`) keep `Rates` extensible without breaking callers;
   parameters stay unexported to consumers through `internal/`.
2. **`HTTPClient` interface** (`Do(*http.Request)`) allows injecting test
   doubles and custom transports; `New()` uses `http.DefaultClient`.
3. **Synthetic GEL entry** lets converters treat GEL like any other currency
   (rate 1, quantity 1) even though NBG does not publish it.
4. **Cache key**: CRC32 of `YYYY-MM-DD:` plus sorted, comma-joined codes;
   order of `WithCurrency` calls does not matter.
5. **Concurrency**: `sync.RWMutex`; reads under a read lock, writes under a
   write lock; no request coalescing.
6. **Code generation**: the generator reads one rate set (live API with
   `-client`, or a JSON file with `-input`), adds GEL, sorts, renders
   `codes.go.tmpl`, and formats with `imports.Process`.

## Complexity Tracking

No constitution violations requiring justification. Partially met gates (VI,
VII) are recorded as gaps in [tasks.md](tasks.md).

| Factor | Assessment |
|--------|------------|
| Files | 10 Go source (1 generated), 5 test, 4 fixtures, 1 template |
| Dependency depth | Low: standard library plus logger |
| Concurrency | Read-write mutex in the cache |
| External coupling | NBG endpoint URL and JSON shape |
