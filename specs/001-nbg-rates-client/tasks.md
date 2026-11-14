---

description: "Migrated task list for the NBG rates client with caching"
---

# Tasks: NBG Rates Client with Caching

**Input**: Design documents from `/specs/001-nbg-rates-client/`

**Prerequisites**: plan.md, spec.md

**Status**: migrated — all tasks reflect code that already exists on `master`.
Gaps found during migration are listed at the end without checkboxes; they are
candidates for follow-up features or bug reports, not pending work for this
feature.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Phase 1: Feature Preparation

- [x] T001 Define request parameters and query names in `pkg/nbggovge/internal/params.go` and `pkg/nbggovge/internal/constants.go`
- [x] T002 [P] Add fixtures `2024-02-10-all.json`, `2024-02-10-eur.json`, `2024-02-10-usd-gbp.json`, `empty.json` in `pkg/nbggovge/testdata/`

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T003 Define `Rates`, `Currency`, `ErrCodeNotFound`, and response parsing (single rate set, empty array) in `pkg/nbggovge/response.go`
- [x] T004 [P] Implement `WithDate` and `WithCurrency` (deduplicated codes) in `pkg/nbggovge/option/option.go`

---

## Phase 3: User Story 1 - Fetch official rates for a date (Priority: P1) 🎯 MVP

### Tests for User Story 1

- [x] T005 [P] [US1] Table-driven `CurrencyByCode` tests (USD, RUB found; unknown code returns `ErrCodeNotFound`) in `pkg/nbggovge/response_test.go`; case-insensitive matching is implemented but not tested
- [x] T006 [P] [US1] Integration tests against the live API plus a mocked unknown-currency case, gated by `INTEGRATION_TESTS`, in `pkg/nbggovge/client_test.go`

### Implementation for User Story 1

- [x] T007 [US1] Implement `Client`, `HTTPClient`, `New`, and `NewWithHTTPClient` in `pkg/nbggovge/client.go`
- [x] T008 [US1] Implement `Rates`: build the URL, default the date to today, send with context, check status, parse body (`pkg/nbggovge/client.go`)
- [x] T009 [US1] Add the synthetic GEL entry and sort currencies by code (`maybeAddGELCodeToResponse`, `sortRates` in `pkg/nbggovge/client.go`)

**Checkpoint**: Rates for a date and set of currencies are available.

---

## Phase 4: User Story 2 - Reuse rates through a cache (Priority: P2)

### Tests for User Story 2

- [x] T010 [P] [US2] Cache miss, hit, expiry, key generation, `ClearCache`, `ClearExpired`, `GetCacheStats`, and TTL 0 tests in `pkg/nbggovge/cache_test.go`
- [x] T011 [P] [US2] Cached constructor tests in `pkg/nbggovge/client_cached_test.go`

### Implementation for User Story 2

- [x] T012 [US2] Implement `CachedClient` with TTL, CRC32 key of date and sorted codes, and `RWMutex` in `pkg/nbggovge/cache.go`
- [x] T013 [US2] Implement `ClearCache`, `ClearExpired`, `GetCacheStats`, and `CacheStats` in `pkg/nbggovge/cache.go`
- [x] T014 [US2] Add `NewCached`, `NewCachedWithTTL`, `NewCachedWithHTTPClient`, `NewCachedWithHTTPClientAndTTL` in `pkg/nbggovge/client.go`
- [x] T015 [US2] Use `nbggovge.NewCached()` in `service.New()` (`internal/service/service.go`)

**Checkpoint**: Repeated requests within the TTL do not call NBG.

---

## Phase 5: User Story 3 - Use supported currency codes as constants (Priority: P3)

- [x] T016 [US3] Implement the generator (`-client`, `-input`, `-output`, `-pkg`; add GEL; sort; render; format) in `pkg/nbggovge/currencies/generator/main.go` and `codes.go.tmpl`
- [x] T017 [US3] Generate `pkg/nbggovge/currencies/codes.gen.go` (44 codes and `All()`) and add `pkg/nbggovge/currencies/doc.go`

---

## Phase 6: User Story 4 - Test without the network (Priority: P3)

- [x] T018 [US4] Implement `mock.NewClient` serving fixture rate sets through the real client in `pkg/nbggovge/mock/mock.go`
- [x] T020 [US4] Select the fixture set by the requested date, filter by requested currencies, return an empty set like NBG, and leave fixtures unmodified in `pkg/nbggovge/mock/mock.go`; tests in `pkg/nbggovge/mock/mock_test.go`; run the `mocked:` case of `TestClient_Rates` without `INTEGRATION_TESTS` (fixes gaps 1 and 2, 2026-11-15)

---

## Phase N: Polish & Cross-Cutting Concerns

- [x] T019 [P] Document cache usage, TTL choices, and thread safety in `pkg/nbggovge/README_CACHE.md` and the "For Developers" section of `README.md`

---

## Gaps Found During Migration

Not tasks of this feature. Track them with `/speckit-bugfix-report` or a new
`/speckit-specify` feature.

1. **`client.Rates` has no default-run test.** Partly fixed by T020: the
   `mocked:` case of `TestClient_Rates` and `mock/mock_test.go` now run by
   default and cover the query parameters, GEL insertion, sorting, and empty
   responses. Non-200 statuses, transport errors, and responses with more than
   one rate set are still untested.
2. **Fixed by T020 (2026-11-15).** `mock.NewClient` defects
   (`pkg/nbggovge/mock/mock.go`, before the fix):
   - reads the date from the `currencies` query parameter instead of `date`
     (`mock.go:32`), so the response date becomes the first requested
     currency code (e.g., `"EUR"`); today's date is used only when no
     currency is requested, which already returns 404;
   - does not filter by requested currencies: it always returns the whole
     fixture and checks only the first code (`q.Get`) to decide on 404;
   - returns 404 for an unknown code or no code, while NBG returns an empty
     rate set with status 200 (`testdata/empty.json`);
   - mutates the fixture slice passed to `NewClient`.
3. **Currency codes are regenerated manually.** No `//go:generate` directive
   or Makefile target runs the generator; `make codegen` only runs stringer.
   `codes.gen.go` was last generated on 2025-08-02. Constitution VII (1.0.2)
   documents the manual command; automating it remains open.
4. **No automatic cache eviction.** `ClearExpired` and `GetCacheStats` are not
   called anywhere; in the long-running bot, entries accumulate (one per
   date and currency set).
5. **No HTTP timeout.** `New()` uses `http.DefaultClient`; only the caller's
   context bounds a request.
6. **CRC32 cache key.** A hash collision would return rates for a different
   date or currency set.
7. **Shared cached data.** Cached `Rates` are returned by value, but the
   `Currencies` slice is shared; a caller that mutates it changes the cache.
8. **Generator empty-response guard never fires.** `getInput`
   (`pkg/nbggovge/currencies/generator/main.go:143`) checks
   `len(rates.Currencies) == 0`, but with no requested codes the client always
   appends GEL (FR-004), so an empty NBG response yields one currency and
   `codes.gen.go` would be regenerated with GEL only, without an error. If
   the guard did fire, it would wrap a nil error and produce an unclear
   message.
