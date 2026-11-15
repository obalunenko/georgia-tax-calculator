package taxes

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidIncomeType returned when income type is invalid.
var ErrInvalidIncomeType = errors.New("invalid income type")

// IncomeType is how Small Business income was received; it selects the field of the
// Small Business monthly declaration the income is reported in.
type IncomeType uint

const (
	incomeTypeUnknown IncomeType = iota

	// IncomeTypeCash is income received in cash (declaration field 18).
	IncomeTypeCash
	// IncomeTypePOSTerminal is income received through a POS terminal (declaration field 19).
	IncomeTypePOSTerminal
	// IncomeTypeNonCash is non-cash income except POS terminals (declaration field 20).
	IncomeTypeNonCash
	// IncomeTypeOther is other income, such as services not taxed at source or barter (declaration field 21).
	IncomeTypeOther
	// IncomeTypeUnpaidSupply is taxable income from goods or services not paid for in the month (declaration field 22).
	IncomeTypeUnpaidSupply

	// incomeTypeSentinel should be always last - used as a border of valid values.
	incomeTypeSentinel
)

// Declaration field numbers of the income types.
const (
	fieldCash         = 18
	fieldPOSTerminal  = 19
	fieldNonCash      = 20
	fieldOther        = 21
	fieldUnpaidSupply = 22
)

type incomeTypeInfo struct {
	name        string
	field       int
	description string
}

// incomeTypes follow the section "income tax calculation" of the Small Business monthly declaration
// on decl.rs.ge, fields (18)–(22).
var incomeTypes = map[IncomeType]incomeTypeInfo{
	IncomeTypeCash: {
		name:        "Cash",
		field:       fieldCash,
		description: "cash payments (cash register turnover)",
	},
	IncomeTypePOSTerminal: {
		name:        "POS terminal",
		field:       fieldPOSTerminal,
		description: "card payments through a POS terminal",
	},
	IncomeTypeNonCash: {
		name:        "Non-cash",
		field:       fieldNonCash,
		description: "bank transfers and other non-cash payments (not POS)",
	},
	IncomeTypeOther: {
		name:        "Other",
		field:       fieldOther,
		description: "services not taxed at source, barter, and other income",
	},
	IncomeTypeUnpaidSupply: {
		name:        "Unpaid supply",
		field:       fieldUnpaidSupply,
		description: "goods or services supplied, not paid for this month",
	},
}

// AllIncomeTypes returns all valid IncomeType in declaration field order.
func AllIncomeTypes() []IncomeType {
	types := make([]IncomeType, 0, len(incomeTypes))

	for t := incomeTypeUnknown + 1; t < incomeTypeSentinel; t++ {
		types = append(types, t)
	}

	return types
}

// ParseIncomeType parses IncomeType from its name, case-insensitively.
func ParseIncomeType(raw string) (IncomeType, error) {
	name := strings.TrimSpace(raw)

	for t := incomeTypeUnknown + 1; t < incomeTypeSentinel; t++ {
		if strings.EqualFold(incomeTypes[t].name, name) {
			return t, nil
		}
	}

	return incomeTypeUnknown, fmt.Errorf("%q: %w", raw, ErrInvalidIncomeType)
}

// IncomeTypeLabel returns the label of the income type with this name, e.g. "Cash (18)",
// or "" when the name is not an income type.
func IncomeTypeLabel(name string) string {
	t, err := ParseIncomeType(name)
	if err != nil {
		return ""
	}

	return t.Label()
}

// Valid checks if value of IncomeType is in valid borders.
func (i IncomeType) Valid() bool {
	return i > incomeTypeUnknown && i < incomeTypeSentinel
}

// String returns the income type name, e.g. "Non-cash", or "IncomeType(<n>)" when it is invalid.
func (i IncomeType) String() string {
	if !i.Valid() {
		return fmt.Sprintf("IncomeType(%d)", uint(i))
	}

	return incomeTypes[i].name
}

// Field returns the declaration field number of the income type, or 0 when it is invalid.
func (i IncomeType) Field() int {
	return incomeTypes[i].field
}

// Label returns the name with the declaration field number, e.g. "Cash (18)", or "" when it is invalid.
func (i IncomeType) Label() string {
	if !i.Valid() {
		return ""
	}

	return fmt.Sprintf("%s (%d)", i.String(), i.Field())
}

// Description explains which income the type covers, or returns "" when it is invalid.
func (i IncomeType) Description() string {
	return incomeTypes[i].description
}
