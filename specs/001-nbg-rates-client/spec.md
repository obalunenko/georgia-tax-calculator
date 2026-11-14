# Feature Specification: NBG Rates Client with Caching

**Feature Directory**: `specs/001-nbg-rates-client`

**Feature Branch**: N/A (migrated from `master`; implemented before Spec Kit adoption)

**Created**: 2026-11-14

**Status**: migrated

**Input**: Reverse-engineered from `pkg/nbggovge/` and its tests. Main history:
`cbb73ce` (currencies generator, #173), `df895dc` (caching, #174), `99f1403`
(GEL in responses, #82).

## Project Scope and Compatibility

- **Affected interfaces**: Library only (`pkg/nbggovge`). Both the CLI and the
  bot consume it indirectly through `internal/service` → `internal/converter`.
- **Affected boundaries**: `pkg/nbggovge/` (client, cache, response model),
  `pkg/nbggovge/option/` (request options), `pkg/nbggovge/internal/` (request
  parameters), `pkg/nbggovge/currencies/` (generated currency codes and the
  generator tool), `pkg/nbggovge/mock/` (test double).
- **Tax categories and rates**: N/A.
- **Currencies and rate dates**: All currencies published by NBG for a date;
  44 codes (including GEL) are generated as constants. The request date
  defaults to the current day.
- **User-facing contract**: Go API: `Client`, `HTTPClient`, `New*`
  constructors, `CachedClient` with `ClearCache`, `ClearExpired`,
  `GetCacheStats`, `Rates`/`Currency` model, `ErrCodeNotFound`,
  `option.WithDate`, `option.WithCurrency`, `currencies.All()`,
  `mock.NewClient`.
- **Release impact**: None directly; the package ships inside both binaries.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Fetch official rates for a date (Priority: P1)

A caller (the converter, the currencies generator, or another program) asks
for NBG official rates for a date and an optional list of currency codes and
gets one rate set with every requested currency.

**Why this priority**: Every conversion and tax calculation depends on the
official rate for the income date.

**Independent Test**: Call `Rates` with `option.WithDate` and
`option.WithCurrency` against the real API (`INTEGRATION_TESTS=true`) and
compare with fixtures in `pkg/nbggovge/testdata/`.

**Acceptance Scenarios**:

1. **Given** the date 2024-02-10 and currency EUR, **When** `Rates` is called,
   **Then** the result equals `testdata/2024-02-10-eur.json`.
2. **Given** the date 2024-02-10 and currencies USD and GBP, **When** `Rates`
   is called, **Then** the result equals `testdata/2024-02-10-usd-gbp.json`.
3. **Given** a date and no currencies, **When** `Rates` is called, **Then** all
   currencies for that date are returned, GEL is included with rate 1, and the
   list is sorted by code (`testdata/2024-02-10-all.json`).
4. **Given** a currency code NBG does not know, **When** `Rates` is called,
   **Then** an empty rate set is returned without error (`testdata/empty.json`).
5. **Given** a rate set, **When** `CurrencyByCode` is called with a code in any
   letter case, **Then** the matching currency is returned, or
   `ErrCodeNotFound` if it is absent.

---

### User Story 2 - Reuse rates through a cache (Priority: P2)

A long-running caller (the bot) or a repeated calculation reuses rates for the
same date and currency set instead of calling NBG again.

**Why this priority**: Reduces HTTP requests and latency; the service layer
always uses the cached client.

**Independent Test**: Wrap a counting fake `Client` with `NewCachedClient` and
check call counts (`pkg/nbggovge/cache_test.go`).

**Acceptance Scenarios**:

1. **Given** an empty cache, **When** rates are requested, **Then** the
   underlying client is called once and the result is stored.
2. **Given** a cached entry younger than the TTL, **When** the same date and
   currency set is requested (in any order of codes), **Then** the cached rates
   are returned without calling the underlying client.
3. **Given** a cached entry older than the TTL, **When** it is requested,
   **Then** the underlying client is called again.
4. **Given** TTL 0, **When** entries are requested at any time, **Then** they
   never expire until the process restarts.
5. **Given** the underlying client fails, **When** rates are requested,
   **Then** the error is returned and nothing is cached.
6. **Given** cached entries, **When** `ClearCache`, `ClearExpired`, or
   `GetCacheStats` is called, **Then** all entries are removed, expired entries
   are removed, or total and expired counts are reported.

---

### User Story 3 - Use supported currency codes as constants (Priority: P3)

Interfaces and tests reference currency codes as constants and list all
supported codes.

**Why this priority**: The CLI and the bot build currency choices from
`currencies.All()`; constants prevent typos.

**Independent Test**: Run the generator against a saved NBG response and
compare the output with `pkg/nbggovge/currencies/codes.gen.go`.

**Acceptance Scenarios**:

1. **Given** an NBG rate set (from the live API with `-client` or a JSON file
   with `-input`), **When** the generator runs, **Then** it writes one constant
   per code with the currency name as a comment, adds GEL if missing, sorts
   codes, and generates `All()`.
2. **Given** the generated file, **When** it is inspected, **Then** its header
   records the source (`client` or `file`), the rate date, and the generator
   and Go versions.

---

### User Story 4 - Test without the network (Priority: P3)

Package consumers test conversion logic without calling NBG.

**Why this priority**: The constitution requires default tests to run without
network access.

**Independent Test**: `pkg/nbggovge/mock/mock_test.go`; `internal/converter`
tests use `mock.NewClient` with fixtures from `internal/converter/testdata/`.

**Acceptance Scenarios**:

1. **Given** fixture rate sets, **When** a client is built with
   `mock.NewClient`, **Then** `Rates` returns, through the real client code
   path, the set whose date matches the requested date, limited to the
   requested currencies (all when none are requested), and the fixtures are
   not modified.
2. **Given** a date without a fixture set, or currencies missing from the
   set, **When** `Rates` is called, **Then** an empty rate set is returned
   without error, as NBG does.
3. **Given** a custom `HTTPClient`, **When** it is passed to
   `NewWithHTTPClient` or a `NewCached*WithHTTPClient` constructor, **Then**
   all requests go through it.

---

### Edge Cases

- No date option: the current day is used (both by the client and for the cache
  key).
- Duplicate `WithCurrency` codes: the code is added once.
- Unknown currency: empty rate set, no error.
- NBG returns a non-200 status: error `invalid response status: <status>`.
- NBG returns more than one rate set: error.
- NBG returns an empty array: empty rate set, no error.
- GEL requested explicitly or no codes requested: a synthetic GEL entry
  (rate 1, quantity 1, name "Georgian Lari") is added if NBG omits it.
- Rates are per `Quantity` units of the currency; the client does not
  normalize them.
- Concurrent cache use: protected by a read-write mutex; concurrent misses for
  the same key may both call NBG.
- A failure while closing the response body is logged, not returned.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The client MUST request
  `https://nbg.gov.ge/gw/api/ct/monetarypolicy/currencies/en/json` with a
  `date` query parameter (`YYYY-MM-DD`) and one `currencies` parameter per
  requested code.
- **FR-002**: The request date MUST default to the current day when no date
  option is given.
- **FR-003**: With no currency option, the client MUST return all currencies
  NBG publishes for the date.
- **FR-004**: The client MUST add a GEL entry with rate 1 when no codes are
  requested or GEL is requested and NBG does not return it.
- **FR-005**: Returned currencies MUST be sorted by code.
- **FR-006**: The client MUST return an error for transport failures, non-200
  statuses, unreadable bodies, invalid JSON, and responses with more than one
  rate set; errors are wrapped with context.
- **FR-007**: Requests MUST honor the caller's `context.Context`.
- **FR-008**: `Rates.CurrencyByCode` MUST match codes case-insensitively and
  return `ErrCodeNotFound` when absent.
- **FR-009**: `CachedClient` MUST implement `Client` and cache successful
  results keyed by date and the sorted set of currency codes.
- **FR-010**: Cache entries MUST expire after the TTL; TTL 0 MUST mean no
  expiration; the default TTL MUST be 1 hour.
- **FR-011**: Failed requests MUST NOT be cached.
- **FR-012**: The cache MUST be safe for concurrent use.
- **FR-013**: `CachedClient` MUST expose `ClearCache`, `ClearExpired`, and
  `GetCacheStats` (total and expired entries).
- **FR-014**: The package MUST provide currency code constants and `All()`,
  generated from NBG data and including GEL.
- **FR-015**: Callers MUST be able to inject an `HTTPClient`, and
  `mock.NewClient` MUST serve fixture rate sets through the real client,
  selected by the requested date and limited to the requested currencies,
  returning an empty set where NBG would.
- **FR-016**: `New()` MUST use an HTTP client with a 30-second timeout, so a
  request without a context deadline cannot hang. (Added 2026-11-15, T021.)
- **FR-017**: The cache key MUST be the date and the sorted codes themselves,
  not a hash, so different requests never share a key. (Added 2026-11-15,
  T022.)
- **FR-018**: `CachedClient` MUST store and return copies of `Rates`, so a
  caller changing a result does not change the cache. (Added 2026-11-15,
  T022.)
- **FR-019**: Expired entries MUST be removed when a new entry is stored.
  (Added 2026-11-15, T022.)
- **FR-020**: The currencies generator MUST fail when the live response has
  no currency other than GEL. (Added 2026-11-15, T023.)

### Key Entities

- **Rates**: Rate set for one date: `Date` and a list of `Currency`.
- **Currency**: `Code`, `Name`, `Quantity` (units the rate applies to),
  `Rate`, `Diff`, formatted rate and diff, `Date`, `ValidFromDate`.
- **Rates request parameters**: Date and currency codes, set through
  `option.RatesOption`.
- **Cache entry**: Rates plus the time they were stored, keyed by date and
  sorted codes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For the fixture dates, live responses match the stored fixtures
  (`make test-integration`).
- **SC-002**: A repeated request for the same date and currency set within the
  TTL makes no additional HTTP calls.
- **SC-003**: Default test runs (`go test -race ./...`) need no network access.
- **SC-004**: The CLI (`cmd/ge-tax-calc/prompts.go`) and the bot
  (`cmd/bot/keyboards.go`) offer exactly the codes from `currencies.All()` as
  currency choices.

## Assumptions

- NBG keeps the current endpoint and JSON shape: an array with at most one
  rate set per date.
- NBG returns rates for the requested date; how weekends and holidays are
  answered is decided by NBG and is not handled specially by the client.
- The currency list changes rarely; regenerating `codes.gen.go` manually is
  acceptable (last generated 2025-08-02 from the live API).
- Network access to nbg.gov.ge is available at runtime.
