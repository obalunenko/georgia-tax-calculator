# Feature Specification: Income Type and Declaration Field Numbers

**Feature Directory**: `specs/007-income-type-declaration-fields`

**Feature Branch**: `feat/income-type-declaration-fields` (from `develop`; directory numbering does not create a branch)

**Created**: 2026-11-15

**Status**: Implemented (2026-11-15); quickstart result scenarios run through the live service, interactive TUI and test-bot runs not done by hand (covered by model and fake-Telegram tests)

**Input**: User description: "добавь возможность указать тип дохода чтобы потом
в отчете при подсчете налогов выводить правильный номер поля в декларации для
заполнения. скриншот формы декларации: [Image #1]" (Add a way to specify the
income type, so the tax calculation report shows which declaration field to
fill in. The screenshot shows the "Income tax calculation (GEL)" section of the
Small Business monthly declaration on decl.rs.ge, fields (15)–(22).)

## Project Scope and Compatibility

- **Affected interfaces**: Both. CLI `ge-tax-calc run` and bot `/calculate`
  ask for the income type and show the same declaration fields (constitution
  II). Currency conversion (`convert`, `/convert`) is not affected.
- **Affected boundaries**:
  - `internal/service/`: the income type in the calculation request, its
    validation, the one-month check for Small Business, and the declaration
    field values in the calculation result.
  - `internal/taxes/`: the list of income types and their field numbers, and
    which categories use the Small Business declaration.
  - `pkg/moneyutils/`: formatting of declaration values with two decimals.
  - `cmd/ge-tax-calc/`: the year-income prompt text for Small Business, the
    income type step, month options after the first income, the type in income
    lists and the review, the declaration block in the result.
  - `cmd/bot/`: the same for the bot flow and its messages.
  - `README.md`: the income types, the one-month rule, the year-income prompt,
    and the declaration block (FR-011).
  - `internal/converter/`, `pkg/` packages: not changed.
- **Tax categories and rates**: Small Business (1%) and Small Business (over
  limit) (3%) get the income type and the declaration fields. Employment (20%)
  is unchanged, because it is declared on a different form. Rates are
  unchanged.
- **Currencies and rate dates**: Each income still uses the NBG rate for its
  own date. All income dates of one Small Business calculation fall in one
  calendar month; Employment keeps the one-year rule (006 FR-005).
- **User-facing contract**: For the Small Business categories, a year-income
  prompt that asks for field (15) of the previous month's declaration and an
  income type choice for each income; month choices limited to the first income's month for
  later incomes in those categories; the type in the income list, the final
  review, and the result; a "Declaration fields" block in the result; a
  recalculation line in the Small Business limit warning; a new service error
  for Small Business incomes from several months (interfaces prevent it). No
  new commands, flags, or environment variables.
- **Release impact**: Both binaries and the bot image. No generated code
  changes (the income type does not use stringer, plan research R1).

## Declaration Form Reference

Fields of the section "საშემოსავლო გადასახადის გაანგარიშება (ლარი)" (income
tax calculation, GEL) of the Small Business monthly declaration, transcribed
from the user's screenshot of decl.rs.ge (2026-11-14):

| Field | Georgian label (abridged) | Meaning | Value in the calculator |
|-------|---------------------------|---------|-------------------------|
| (15) | მცირე ბიზნესის ერთობლივი შემოსავალი კალენდარული წლის დასაწყისიდან ნაზარდი ჯამით | Small Business gross income since the start of the calendar year, cumulative | Reported year income: the user's income since the start of the year plus the converted incomes |
| (16) | საშემოსავლო გადასახადის განაკვეთი საანგარიშო თვეში | Income tax rate in the reporting month, % | Rate of the selected category: 1 or 3 |
| (17) | დასაბეგრი შემოსავალი საანგარიშო თვის მიხედვით, მათ შორის: | Taxable income for the reporting month; the sum of (18)–(22) | Total converted income |
| (18) | ნაღდი ფულით ანგარიშსწორებით მიღებული შემოსავალი (საკონტროლო-სალარო აპარატის ბრუნვა) | Income received in cash (cash register turnover) | Converted incomes of type Cash |
| (19) | პოს-ტერმინალით მიღებული შემოსავალი | Income received through a POS terminal | Converted incomes of type POS terminal |
| (20) | უნაღდო ანგარიშსწორებით მიღებული შემოსავალი (პოს-ტერმინალების გარდა) | Non-cash income (bank transfers and similar), except POS terminals | Converted incomes of type Non-cash |
| (21) | სხვა შემოსავალი (გაწეული მომსახურება, რომელიც არ იბეგრება გადახდის წყაროსთან, ბარტერი და სხვა) | Other income: services not taxed at source, barter, and others | Converted incomes of type Other |
| (22) | საქონლის მიწოდებით/მომსახურების გაწევით მიღებული დასაბეგრი შემოსავალი, რომლის საკომპენსაციო თანხა არ არის ანაზღაურებული მიმდინარე საანგარიშო პერიოდში | Taxable income from goods supplied or services provided that were not paid for in the reporting period | Converted incomes of type Unpaid supply |

The screenshot's values: (15) 35942.00, (16) 1.00 %, (17) 4150.00,
(18) 3950.00, (19) 0.00, (20) 200.00, (21) 0.00, (22) 0.00, where
(17) = (18) + (19) + (20) + (21) + (22).

## Clarifications

### Session 2026-11-15

- Q: How should a Small Business calculation handle incomes from more than one
  month, given that fields (16) and (17) refer to one reporting month? → A:
  Limit the calculation to the month of its first income, as the year is
  limited today (006 FR-005); one calculation matches one monthly
  declaration.
- Q: When a Small Business (1%) calculation exceeds the 500 000 GEL limit in
  the reporting month, what should the result show, given that Article 90(2)
  taxes that whole month at 3% while the block shows (16) 1.00? → A: The limit
  warning adds a line telling the user to recalculate this month's
  declaration with Small Business (over limit), where field (16) is 3.00; the
  block and the tax stay at the selected rate (006 FR-002).
- Q: What should a Small Business user enter as the income since the start of
  the year, and how should the prompt say it, so that this month's incomes are
  not counted twice in field (15)? → A: For the Small Business categories the
  prompt asks for field (15) of the previous month's declaration (0 for the
  first declaration of the year); the Employment prompt is unchanged.
- Q: Should the bot show the declaration values in monospace so that Telegram
  copies each one with a tap? → A: Yes, only the values of the block
  (15)–(22) in the bot; the rest of the message and the CLI output stay plain
  text, with the same labels and numbers in both interfaces.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - See which declaration field each amount goes to (Priority: P1)

A taxpayer with Small Business status enters the month's incomes and chooses
how each one was received. The result lists the values for declaration fields
(15)–(22), so the taxpayer copies them into the form without adding up incomes
by type.

**Why this priority**: This is the purpose of the feature: the report says
what to enter in each field of the declaration.

**Independent Test**: In either interface calculate with Small Business, year
income `31792`, an income of `3950 GEL` of type Cash and an income of
`200 GEL` of type Non-cash in the same month, and compare the declaration block
with the screenshot.

**Acceptance Scenarios**:

1. **Given** category Small Business, year income 31 792 GEL, an income of
   3 950 GEL of type Cash, and an income of 200 GEL of type Non-cash in the
   same month, **When** the tax is calculated, **Then** the declaration block
   shows (15) 35942.00, (16) 1.00, (17) 4150.00, (18) 3950.00, (19) 0.00,
   (20) 200.00, (21) 0.00, (22) 0.00, and the tax is 41.50 GEL.
2. **Given** category Small Business and two Non-cash incomes, in USD and in
   EUR, that convert to 1 000.00 GEL and 500.55 GEL, **When** the tax is
   calculated, **Then** fields (17) and (20) are 1500.55 and fields (18),
   (19), (21), and (22) are 0.00.
3. **Given** category Small Business (over limit), **When** the tax is
   calculated, **Then** field (16) is 3.00.
4. **Given** the same input, **When** it is calculated in the CLI and in the
   bot, **Then** both show the same declaration block labels and values; in
   the bot each value is monospace and copied with a tap.
5. **Given** category Small Business, year income 499 000 GEL, and a Non-cash
   income of 2 000 GEL, **When** the tax is calculated, **Then** the block
   shows (16) 1.00 and (20) 2000.00, the tax is 20 GEL, and the limit warning
   adds a line telling the user to recalculate this month's declaration with
   Small Business (over limit), where field (16) is 3.00.
6. **Given** a Small Business category was chosen, **When** the income since
   the start of the year is asked, **Then** the prompt asks for field (15) of
   the previous month's declaration and says to enter 0 for the first
   declaration of the year; with Employment the prompt is the same as before.

---

### User Story 2 - Choose the income type for each income (Priority: P1)

After choosing the currency of an income, the taxpayer chooses its type from
the five types of the declaration. The type is shown wherever the income is
listed.

**Why this priority**: The declaration block cannot be built without the
type; together with User Story 1 it is the minimum useful feature.

**Independent Test**: In either interface start a Small Business calculation,
enter an income up to the currency, and check the type choice; finish with two
incomes of different types and check the review and the result.

**Acceptance Scenarios**:

1. **Given** category Small Business or Small Business (over limit), **When**
   the user has chosen the currency of an income, **Then** the interface
   offers Cash (18), POS terminal (19), Non-cash (20), Other (21), and Unpaid
   supply (22), in this order, each with its field number.
2. **Given** captured incomes with types, **When** the income list or the
   final review is shown, **Then** each income line includes its type.
3. **Given** the CLI and a second or later income, **When** the type choice
   is shown, **Then** the type of the previous income is preselected.
4. **Given** the bot, **When** the user presses a type button from an earlier
   message, **Then** the bot replies "Unexpected input" and the step does not
   change (004 FR-015).

---

### User Story 3 - Keep one Small Business calculation within one month (Priority: P1)

A Small Business taxpayer adds several incomes. After the first income, only
its month can be chosen, so fields (16) and (17) describe one reporting month,
as the declaration does.

**Why this priority**: Without it, field (17) can add up incomes from several
months and the block no longer matches any declaration.

**Independent Test**: In either interface choose Small Business, add an income
dated 2026-10-05, start a second income and check that only 2026 and October
are offered; call the service with Small Business incomes from September and
October and check the error.

**Acceptance Scenarios**:

1. **Given** a Small Business category and a first income dated 2026-10-05,
   **When** the user adds another income, **Then** the year choice offers only
   2026 and the month choice offers only October (preselected in the CLI, the
   only button in the bot), and days follow the existing rules for that month.
2. **Given** the user answers "No" at the final review and the flow restarts
   income entry, **When** the first income is entered again, **Then** any year
   and month allowed today can be chosen again.
3. **Given** a Small Business calculation request with incomes dated in
   September and October of one year, **When** the service calculates,
   **Then** it fails with an error that the incomes span more than one
   calendar month, and no result is returned.
4. **Given** category Employment and a first income dated in October 2026,
   **When** the user adds another income, **Then** any month of 2026 allowed
   today can be chosen (006 FR-005 unchanged).

---

### User Story 4 - Employment calculation is unchanged (Priority: P2)

A taxpayer who chooses Employment goes through the same steps and gets the
same result as before: no income type, no month limit, no declaration block.

**Why this priority**: The fields belong to the Small Business declaration;
showing them for Employment would point the user to the wrong form.

**Independent Test**: Calculate with Employment in either interface and
compare the steps and the result with the current version.

**Acceptance Scenarios**:

1. **Given** category Employment, **When** the user enters an income,
   **Then** no income type is asked.
2. **Given** category Employment, **When** the tax is calculated, **Then** the
   result has no declaration block and is otherwise unchanged.

---

### Edge Cases

- All incomes of one type: the fields of the other types show 0.00.
- Field (17) equals the sum of (18)–(22) to the tetri: each converted amount
  is already rounded to 2 places (002 FR-003), and totals add those amounts.
- Year income 0: field (15) equals field (17).
- Year income with more than two decimals (e.g. 31792.005): it is rounded to
  the tetri before use, so the Year Income line, the limit check, and field
  (15) show the same amount (FR-005; code review, 2026-11-15).
- The Small Business limit warning (006 FR-001) is still shown when it
  applies, with the recalculation line of FR-012; field (16) still shows the
  selected category's rate (006 FR-002).
- Small Business (over limit) with a year income at or below the limit: no
  warning and no recalculation line (006 FR-003 unchanged).
- Several currencies in one type: their converted amounts are added into the
  one field.
- A first Small Business income in the current month: later incomes are
  limited to the current month, with days up to today as before.
- Bot: a year or month button from an older keyboard (another month) is
  rejected as unexpected input (004 FR-015), because the month step now offers
  one month.
- Bot: typing text at the type step is handled like other button steps of the
  flow; `/cancel` at the type step ends the flow and clears the fixed month;
  "No" at the final review restarts income entry, clears the fixed month, and
  types are entered again with the incomes.
- Bot result with monospace values: the rest of the message (tax rate, incomes,
  totals, the limit warning and its emoji) is shown with the same text as
  before, with no characters lost or turned into formatting.
- Service called for a Small Business category with an income that has no
  type or an unknown type: rejected before any currency conversion.
- Service called for Employment with incomes that carry a type: the types are
  ignored and the result has no declaration block.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: For Small Business and Small Business (over limit), both
  interfaces MUST ask for the type of each income after its currency,
  offering exactly five types in field order: Cash (18), POS terminal (19),
  Non-cash (20), Other (21), Unpaid supply (22). Each option shows its field
  number.
- **FR-002**: For Employment, the interfaces MUST NOT ask for the income type,
  and the result MUST NOT include the declaration block.
- **FR-003**: For a Small Business category, the service MUST reject a
  calculation in which any income has no type or an unknown type, with an
  error that names the position of the income, before any currency conversion.
- **FR-004**: For a Small Business category, the calculation result MUST
  include the declaration field values: (15) the reported year income, (16)
  the category rate in percent, (17) the total converted income, and
  (18)–(22) the sum of the converted amounts of the incomes of the matching
  type, or 0 when there are none.
- **FR-005**: Field (17) MUST equal the sum of fields (18)–(22) and the total
  converted income shown in the result; field (15) MUST equal the reported
  year income shown in the result.
- **FR-006**: Both interfaces MUST show the declaration block with all eight
  fields in order (15) to (22), including fields with 0.00. Each line shows
  the field number, a short English label, and the value with exactly two
  decimal places (for example `4150.00`, rate `1.00`).
- **FR-007**: Each income in the income list, the final review, and the result
  MUST show its type when the category is a Small Business category.
- **FR-008**: The income type MUST NOT change the tax, the totals, the
  rounding, or the limit warning; it affects only how incomes are grouped into
  fields.
- **FR-009**: The CLI and the bot MUST offer the same type and month choices
  and show the same declaration block for the same input.
- **FR-010**: For Small Business categories, after the first income both
  interfaces MUST offer only the year and month of that income for later
  incomes; restarting income entry or the flow MUST clear this restriction.
  The service MUST reject a Small Business calculation whose incomes are
  dated in more than one calendar month, with an error that says so, before
  any currency conversion. Employment keeps the one-year rule (006 FR-005).
- **FR-011**: `README.md` MUST describe the income types, their field numbers,
  the one-month rule for Small Business, the year-income prompt of FR-013, and
  the declaration block.
- **FR-012**: When the Small Business limit warning applies (006 FR-001), it
  MUST add a line saying that this month's declaration is filled at the
  over-limit rate: recalculate with Small Business (over limit), where field
  (16) is 3.00. The declaration block and the tax MUST stay at the selected
  category's rate (006 FR-002), and the line MUST be the same in both
  interfaces.
- **FR-013**: For the Small Business categories, both interfaces MUST ask for
  the income since the start of the year as field (15) of the previous
  month's declaration, and say to enter 0 for the first declaration of the
  year (January, or the first month with Small Business status). The prompt
  for Employment MUST stay unchanged. The value rules of the input (003, 006)
  do not change.
- **FR-014**: In the bot, each value of the declaration block MUST be shown in
  monospace, so that Telegram copies it with a tap. Field numbers, labels, and
  the rest of the result message MUST stay plain text, and the CLI output MUST
  stay plain text; the labels and numbers MUST be identical in both
  interfaces (FR-009).

### Key Entities

- **Income type**: One of Cash, POS terminal, Non-cash, Other, Unpaid supply;
  each has a declaration field number (18)–(22) and a short label. Used only
  for the Small Business categories.
- **Income**: Date, amount, currency, and, for the Small Business categories,
  income type.
- **Declaration field**: Field number, short English label, and value (GEL,
  or percent for field (16)).
- **Declaration block**: Fields (15)–(22) in form order; present in the result
  only for the Small Business categories.
- **Calculation month**: The calendar month of the first income of a Small
  Business calculation; all its incomes share it, and it is the reporting
  month of fields (16) and (17).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: For the input of US1.1, both interfaces show the same eight
  values as the screenshot of the declaration.
- **SC-002**: For any Small Business input, field (17) equals the sum of
  fields (18)–(22) and the total converted income, to the tetri.
- **SC-003**: The income type adds exactly one choice per income in the Small
  Business categories and none in Employment.
- **SC-004**: The CLI and the bot produce the same declaration block labels and
  values for the same input.
- **SC-005**: For the same incomes (with an income type added to each income
  in the Small Business categories, and dated in one month), the tax and the
  totals are the same as before the feature, the limit warning appears for
  exactly the same inputs (now with the FR-012 line), and Employment results
  are unchanged.
- **SC-006**: A Small Business calculation with incomes from two different
  months cannot be entered in either interface and is rejected by the
  service.
- **SC-007**: In the bot, each of the eight declaration values is copied with
  one tap, and the copied text is exactly the number shown (for example
  `4150.00`).

## Assumptions

- Field numbers and labels come from the user's screenshot of the Small
  Business monthly declaration on decl.rs.ge taken on 2026-11-14. If the
  Revenue Service renumbers the form, only the field table changes.
- Fields after (22), including the tax payable, are not in the screenshot and
  are out of scope; the result keeps the existing tax line without a field
  number. The form's "automatic filling" (ავტომატური შევსება) checkbox is
  out of scope.
- The type is chosen per income, because one month can include cash and
  non-cash incomes. There is no default type for the first income;
  foreign-currency bank transfers are Non-cash (20).
- Employment income is declared on a different form, so Employment gets no
  type and no fields.
- Field (15) of the previous month's declaration is the income since the start
  of the year up to the reporting month (FR-013), so field (15) of this month
  is that amount plus the converted incomes, which matches the cumulative
  meaning of the field.
- Labels in the output are short English names; the field numbers link them to
  the Georgian form.
- Income types are not stored between calculations.
- The date of an Unpaid supply (22) income is the date the goods were
  supplied or the service was provided; it is converted at the NBG rate of
  that date, like any other income.
- A Small Business taxpayer with incomes in several months runs one
  calculation per month, entering field (15) of the previous month's
  declaration each time.
