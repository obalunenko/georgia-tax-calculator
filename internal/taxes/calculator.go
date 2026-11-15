// Package taxes provides functionality for calculating taxes.
package taxes

import (
	"errors"
	"fmt"
	"strings"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
)

var (
	// ErrTaxRateNotFound returned when no tax rate found in taxrates.
	ErrTaxRateNotFound = errors.New("tax rate not found")
	// ErrTaxTypeNotSupported returned when TaxType has invalid value.
	ErrTaxTypeNotSupported = errors.New("tax type not supported")
)

//go:generate stringer --type=TaxType --trimprefix=true --linecomment=true

// TaxType represents tax type for calculation.
type TaxType uint

const (
	taxTypeUnknown TaxType = iota

	// TaxTypeSmallBusinessOverLimit is the Small Business rate for turnover above SmallBusinessTurnoverLimit.
	TaxTypeSmallBusinessOverLimit // Small Business (over limit)
	// TaxTypeSmallBusiness is Small Business tax type.
	TaxTypeSmallBusiness // Small Business
	// TaxTypeEmployment is Employment tax type.
	TaxTypeEmployment // Employment

	// taxTypeSentinel should be always last - used as a border of valid values.
	taxTypeSentinel
)

// ErrInvalidTaxType returned when tax type is invalid.
var ErrInvalidTaxType = errors.New("invalid tax type")

var stringToTaxType = map[string]TaxType{
	strings.ToLower(TaxTypeSmallBusiness.String()):          TaxTypeSmallBusiness,
	strings.ToLower(TaxTypeSmallBusinessOverLimit.String()): TaxTypeSmallBusinessOverLimit,
	strings.ToLower(TaxTypeEmployment.String()):             TaxTypeEmployment,
}

// ParseTaxType parses TaxType from string.
func ParseTaxType(raw string) (TaxType, error) {
	tt, ok := stringToTaxType[strings.TrimSpace(strings.ToLower(raw))]
	if !ok {
		return taxTypeUnknown, fmt.Errorf("%s: %w", raw, ErrInvalidTaxType)
	}

	return tt, nil
}

// Valid checks if value of TaxType is in valid borders.
func (i TaxType) Valid() bool {
	return i > taxTypeUnknown && i < taxTypeSentinel
}

// IsSmallBusiness reports whether income of this tax type is declared on the Small Business monthly declaration.
func (i TaxType) IsSmallBusiness() bool {
	return i == TaxTypeSmallBusiness || i == TaxTypeSmallBusinessOverLimit
}

const (
	onePercent     = 0.01
	threePercents  = onePercent * 3
	twentyPercents = onePercent * 20
)

// SmallBusinessTurnoverLimit is the Small Business annual turnover limit in GEL per calendar year
// (Tax Code of Georgia, Articles 88(3), 89(2)(a), 90(2); 700 000 GEL for wine tourism and agrotourism,
// which the calculator does not support).
const SmallBusinessTurnoverLimit float64 = 500_000

// taxrates follow the Tax Code of Georgia (consolidated version of 30.10.2026):
// Small Business 1% (Article 90(1)), Small Business over the limit 3% (Article 90(2)),
// Employment 20% (Article 81(1)).
var taxrates = map[TaxType]TaxRate{
	TaxTypeSmallBusiness:          newTaxRate(TaxTypeSmallBusiness, onePercent),
	TaxTypeSmallBusinessOverLimit: newTaxRate(TaxTypeSmallBusinessOverLimit, threePercents),
	TaxTypeEmployment:             newTaxRate(TaxTypeEmployment, twentyPercents),
}

// TaxRate represents tuple TaxType - rate.
type TaxRate struct {
	Type TaxType
	Rate float64
}

func (t TaxRate) String() string {
	return fmt.Sprintf("%s %s %%", t.Type.String(), moneyutils.ToString(t.Percent()))
}

// Percent returns the rate in percent, e.g. 1 for the 1% rate.
func (t TaxRate) Percent() float64 {
	const toPercentage float64 = 100

	return moneyutils.Multiply(t.Rate, toPercentage)
}

func newTaxRate(tt TaxType, rate float64) TaxRate {
	return TaxRate{
		Type: tt,
		Rate: rate,
	}
}

// AllTaxRates returns all supported TaxRate.
func AllTaxRates() ([]TaxRate, error) {
	taxes := AllTaxTypes()

	resp := make([]TaxRate, 0, len(taxes))

	for _, tax := range taxes {
		tr, err := tax.Rate()
		if err != nil {
			return nil, err
		}

		resp = append(resp, tr)
	}

	return resp, nil
}

// AllTaxTypes returns all supported TaxType.
func AllTaxTypes() []TaxType {
	taxes := make([]TaxType, 0, len(taxrates))

	for tt := range taxrates {
		taxes = append(taxes, tt)
	}

	return taxes
}

// Rate convert TaxType to TaxRate.
func (i TaxType) Rate() (TaxRate, error) {
	tr, ok := taxrates[i]
	if !ok {
		return TaxRate{}, ErrTaxRateNotFound
	}

	return tr, nil
}

// Response represents result of Calc.
type Response struct {
	Money models.Money
	Rate  TaxRate
}

// Calc returns sum of tax for income according to TaxType.
func Calc(income models.Money, taxType TaxType) (Response, error) {
	const roundPlaces int32 = 2

	if !taxType.Valid() {
		return Response{}, fmt.Errorf("%s: %w", taxType.String(), ErrTaxTypeNotSupported)
	}

	tr, err := taxType.Rate()
	if err != nil {
		return Response{}, fmt.Errorf("get tax rate: %w", err)
	}

	sum := moneyutils.Multiply(income.Amount, tr.Rate)

	rounded := moneyutils.Round(sum, roundPlaces)

	return Response{
		Money: models.NewMoney(rounded, income.Currency),
		Rate:  tr,
	}, nil
}
