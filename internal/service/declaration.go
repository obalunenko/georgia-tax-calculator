package service

import (
	"fmt"
	"strings"

	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
)

// DeclarationHeader is the first line of Declaration.String().
const DeclarationHeader = "Declaration fields (Small Business monthly declaration):"

// Fields of the Small Business monthly declaration that are not income types.
const (
	fieldYearIncome    = 15
	fieldTaxRate       = 16
	fieldTaxableIncome = 17
)

// declarationPlaces is the number of decimal places the declaration form expects.
const declarationPlaces int32 = 2

// DeclarationField is one field of the Small Business monthly declaration with the value to enter.
type DeclarationField struct {
	// Number is the field number on the form, e.g. 17.
	Number int
	// Label is a short English name of the field.
	Label string
	// Value is the amount in GEL, or the rate in percent for field 16.
	Value float64
}

// Formatted returns Value with exactly two decimal places, as the form expects, e.g. "4150.00".
func (f DeclarationField) Formatted() string {
	return moneyutils.ToStringFixed(f.Value, declarationPlaces)
}

// Prefix returns the part of the line before the value, e.g. "(17) Taxable income for the month: ".
func (f DeclarationField) Prefix() string {
	return fmt.Sprintf("(%d) %s: ", f.Number, f.Label)
}

// String returns the declaration line, e.g. "(17) Taxable income for the month: 4150.00".
func (f DeclarationField) String() string {
	return f.Prefix() + f.Formatted()
}

// Declaration holds the values of fields (15)–(22) of the Small Business monthly declaration.
type Declaration struct {
	// Fields are in form order, all present.
	Fields []DeclarationField
}

// String returns DeclarationHeader followed by one line per field.
func (d Declaration) String() string {
	lines := make([]string, 0, len(d.Fields)+1)
	lines = append(lines, DeclarationHeader)

	for _, f := range d.Fields {
		lines = append(lines, f.String())
	}

	return strings.Join(lines, "\n")
}

// newDeclaration fills fields (15)–(22): the reported year income, the category rate,
// the total converted income, and its split by income type.
func newDeclaration(tr taxes.TaxRate, yearIncome, total float64, incomes []IncomeResponse) *Declaration {
	byType := make(map[taxes.IncomeType]float64, len(incomes))

	for _, inc := range incomes {
		byType[inc.Type] = moneyutils.Add(byType[inc.Type], inc.Converted.Amount)
	}

	fields := []DeclarationField{
		{Number: fieldYearIncome, Label: "Income since the start of the year", Value: yearIncome},
		{Number: fieldTaxRate, Label: "Tax rate, %", Value: tr.Percent()},
		{Number: fieldTaxableIncome, Label: "Taxable income for the month", Value: total},
	}

	for _, it := range taxes.AllIncomeTypes() {
		fields = append(fields, DeclarationField{Number: it.Field(), Label: it.String(), Value: byType[it]})
	}

	return &Declaration{Fields: fields}
}

// declarationRate formats the rate of tt as the declaration expects it in field 16, e.g. "3.00".
func declarationRate(tt taxes.TaxType) string {
	tr, err := tt.Rate()
	if err != nil {
		return tt.String()
	}

	return moneyutils.ToStringFixed(tr.Percent(), declarationPlaces)
}

// checkIncomeTypes parses the income type of every income for the Small Business categories.
// Other categories do not use income types, so their incomes get the zero type.
func checkIncomeTypes(tt taxes.TaxType, incomes []Income) ([]taxes.IncomeType, error) {
	types := make([]taxes.IncomeType, len(incomes))

	if !tt.IsSmallBusiness() {
		return types, nil
	}

	for i, inc := range incomes {
		it, err := taxes.ParseIncomeType(inc.Type)
		if err != nil {
			return nil, fmt.Errorf("income %d: %w", i+1, err)
		}

		types[i] = it
	}

	return types, nil
}
