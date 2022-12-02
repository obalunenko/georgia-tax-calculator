package taxes

import (
	"errors"
	"fmt"
	"strings"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/moneyutils"
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

	// TaxTypeIndividualEntrepreneur is Individual Entrepreneur tax type.
	TaxTypeIndividualEntrepreneur // Individual Entrepreneur
	// TaxTypeSmallBusiness is Small Business tax type.
	TaxTypeSmallBusiness // Small Business
	// TaxTypeEmployment is Employment tax type.
	TaxTypeEmployment // Employment

	// taxTypeSentinel should be always last - used as a border of valid values.
	taxTypeSentinel
)

var ErrInvalidTaxType = errors.New("invalid tax type")

var stringToTaxType = map[string]TaxType{
	strings.ToLower(TaxTypeSmallBusiness.String()):          TaxTypeSmallBusiness,
	strings.ToLower(TaxTypeIndividualEntrepreneur.String()): TaxTypeIndividualEntrepreneur,
	strings.ToLower(TaxTypeEmployment.String()):             TaxTypeEmployment,
}

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

var taxrates = map[TaxType]TaxRate{
	TaxTypeSmallBusiness:          newTaxRate(TaxTypeSmallBusiness, 0.01),
	TaxTypeIndividualEntrepreneur: newTaxRate(TaxTypeIndividualEntrepreneur, 0.03),
	TaxTypeEmployment:             newTaxRate(TaxTypeEmployment, 0.2),
}

type TaxRate struct {
	Type TaxType
	Rate float64
}

func (t TaxRate) String() string {
	return fmt.Sprintf("%s %s %%", t.Type.String(), moneyutils.ToString(moneyutils.Multiply(t.Rate, 100)))
}

func newTaxRate(tt TaxType, rate float64) TaxRate {
	return TaxRate{
		Type: tt,
		Rate: rate,
	}
}

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

func AllTaxTypes() []TaxType {
	taxes := make([]TaxType, 0, len(taxrates))

	for tt := range taxrates {
		taxes = append(taxes, tt)
	}

	return taxes
}

func (i TaxType) Rate() (TaxRate, error) {
	tr, ok := taxrates[i]
	if !ok {
		return TaxRate{}, ErrTaxRateNotFound
	}

	return tr, nil
}

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
