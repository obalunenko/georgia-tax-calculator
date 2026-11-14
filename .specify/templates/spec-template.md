# Feature Specification: [FEATURE NAME]

**Feature Directory**: `specs/[NNN-feature-name]`

**Feature Branch**: [Existing branch, if used; directory numbering does not create a branch]

**Created**: [DATE]

**Status**: Draft

**Input**: User description: "$ARGUMENTS"

## Project Scope and Compatibility

- **Affected interfaces**: [CLI `ge-tax-calc`, Telegram bot, both, or library
  only (`pkg/`). If only one interface is affected, state why parity with the
  other is not required.]
- **Affected boundaries**: [Select applicable paths: `cmd/ge-tax-calc/`,
  `cmd/bot/`, `internal/service/`, `internal/converter/`, `internal/taxes/`,
  `internal/models/`, `pkg/nbggovge/`, `pkg/moneyutils/`, `pkg/dateutils/`,
  build/release config, or local Spec Kit tooling. Explain each one.]
- **Tax categories and rates**: [Affected `TaxType` categories (Individual
  Entrepreneur, Small Business, Employment) and rates, or N/A.]
- **Currencies and rate dates**: [Affected currencies, how the rate date is
  chosen, and NBG API or cache effects, or N/A.]
- **User-facing contract**: [CLI commands, flags, and menus; bot commands
  (`/start`, `/calculate`, `/convert`, `/cancel`, `/help`); output format;
  environment variables (`TELEGRAM_BOT_TOKEN`, `TELEGRAM_BOT_LOG_LEVEL`,
  `USER_STORE_PATH`), or N/A.]
- **Release impact**: [Effect on `ge-tax-calc`/bot binaries, the bot Docker
  image, or generated code; otherwise N/A.]

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.

  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - [Brief Title] (Priority: P1)

[Describe this user journey in plain language]

**Why this priority**: [Explain the value and why it has this priority level]

**Independent Test**: [Describe how this can be tested independently - e.g., "Can be fully tested by [specific action] and delivers [specific value]"]

**Acceptance Scenarios**:

1. **Given** [initial state], **When** [action], **Then** [expected outcome]
2. **Given** [initial state], **When** [action], **Then** [expected outcome]

---

### User Story 2 - [Brief Title] (Priority: P2)

[Describe this user journey in plain language]

**Why this priority**: [Explain the value and why it has this priority level]

**Independent Test**: [Describe how this can be tested independently]

**Acceptance Scenarios**:

1. **Given** [initial state], **When** [action], **Then** [expected outcome]

---

### User Story 3 - [Brief Title] (Priority: P3)

[Describe this user journey in plain language]

**Why this priority**: [Explain the value and why it has this priority level]

**Independent Test**: [Describe how this can be tested independently]

**Acceptance Scenarios**:

1. **Given** [initial state], **When** [action], **Then** [expected outcome]

---

[Add more user stories as needed, each with an assigned priority]

### Edge Cases

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right edge cases.
-->

- [For changed input: invalid or negative amounts, unknown currency codes,
  malformed or future dates, empty input. Mark unrelated cases N/A.]
- [For changed conversion: dates without an NBG rate, NBG API unavailable or
  slow, cached versus fresh rates, GEL-to-GEL conversion.]
- [For changed calculation: several incomes in different currencies, rounding
  of converted amounts and tax, unsupported tax category.]
- [For changed bot flows: `/cancel` mid-flow, unexpected messages during a step,
  concurrent users.]

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: [Observable capability and the CLI/bot flow that exposes it, e.g., "Users MUST be able to convert an amount to GEL at the NBG rate for a chosen date"]
- **FR-002**: [Validation, error, and parity behavior for both interfaces]
- **FR-003**: [Calculation, rounding, or output-format behavior]

Add only requirements relevant to this feature. Use
`[NEEDS CLARIFICATION: specific unanswered question]` for unresolved behavior,
e.g., "Rate for weekend dates [NEEDS CLARIFICATION: use the last published
rate or reject the date?]". Do not invent accounts, databases, or web
frontends.

### Key Entities *(include if feature involves data)*

- **[Entity 1]**: [What it represents, key attributes without implementation; e.g., Income (amount, currency, date), Tax category (name, rate), Conversion (amount, from, to, rate date)]
- **[Entity 2]**: [What it represents, relationships to other entities]

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: [Observable completion of the intended CLI and/or bot flow, e.g., "A user can calculate tax for three incomes in different currencies in one session"]
- **SC-002**: [Correctness outcome, e.g., "Converted amounts and tax match a reference calculation using the NBG rate for the income date, after rounding"]
- **SC-003**: [Parity or compatibility outcome, e.g., "The CLI and the bot produce the same result for the same input"]

## Assumptions

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right assumptions based on reasonable defaults
  chosen when the feature description did not specify certain details.
-->

- [Assumption about rates, e.g., "NBG publishes the rate for the requested date"]
- [Assumption about scope, e.g., "Only the existing tax categories are affected"]
- [Assumption about environment, e.g., "Network access to nbg.gov.ge is available at runtime"]
- [Dependency on an external service, e.g., "NBG currency rates API; Telegram Bot API for the bot"]
