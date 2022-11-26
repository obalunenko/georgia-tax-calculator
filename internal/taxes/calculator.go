package taxes

import (
	"errors"
	"fmt"

	"github.com/obalunenko/georgia-tax-calculator/internal/moneyutils"
)

//go:generate stringer --type=TaxType --trimprefix=true --linecomment=true

// TaxType represents tax type for calculation.
type TaxType uint

// Valid checks if value of TaxType is in valid borders.
func (i TaxType) Valid() bool {
	return i > taxTypeUnknown && i < taxTypeSentinel
}

const (
	taxTypeUnknown TaxType = iota

	TaxTypeIndividualEntrepreneur // Individual Entrepreneur
	TaxTypeSmallBusiness          // Small Business
	TaxTypeEmployment             // Employment

	taxTypeSentinel
)

var (
	// ErrTaxRateNotFound returned when no tax rate found in taxrates.
	ErrTaxRateNotFound = errors.New("tax rate not found")
	// ErrTaxTypeNotSupported returned when TaxType has invalid value.
	ErrTaxTypeNotSupported = errors.New("tax type not supported")
)

var taxrates = map[TaxType]float64{
	TaxTypeSmallBusiness:          0.01,
	TaxTypeIndividualEntrepreneur: 0.03,
	TaxTypeEmployment:             0.2,
}

// Calc returns sum of tax for income according to TaxType.
func Calc(income float64, taxType TaxType) (float64, error) {
	if !taxType.Valid() {
		return 0, fmt.Errorf("%s: %w", taxType.String(), ErrTaxTypeNotSupported)
	}

	tr, ok := taxrates[taxType]
	if !ok {
		return 0, ErrTaxRateNotFound
	}

	sum := moneyutils.Multiply(income, tr)

	return moneyutils.Round(sum, 2), nil
}
