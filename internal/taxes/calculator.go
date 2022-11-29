package taxes

import (
	"errors"
	"fmt"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
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

	// TaxTypeIndividualEntrepreneur is Individual Entrepreneur tax type.
	TaxTypeIndividualEntrepreneur // Individual Entrepreneur
	// TaxTypeSmallBusiness is Small Business tax type.
	TaxTypeSmallBusiness // Small Business
	// TaxTypeEmployment is Employment tax type.
	TaxTypeEmployment // Employment

	// taxTypeSentinel should be always last - used as a border of valid values.
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
func Calc(income models.Money, taxType TaxType) (models.Money, error) {
	const roundPlaces int32 = 2

	if !taxType.Valid() {
		return models.Money{}, fmt.Errorf("%s: %w", taxType.String(), ErrTaxTypeNotSupported)
	}

	tr, ok := taxrates[taxType]
	if !ok {
		return models.Money{}, ErrTaxRateNotFound
	}

	sum := moneyutils.Multiply(income.Amount, tr)

	rounded := moneyutils.Round(sum, roundPlaces)

	return models.NewMoney(rounded, income.Currency), nil
}
