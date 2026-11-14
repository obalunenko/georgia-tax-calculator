package service

import (
	"errors"
	"fmt"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/dateutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/moneyutils"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

// ErrIncomesSpanYears returned when incomes of one calculation are dated in more than one calendar year.
var ErrIncomesSpanYears = errors.New("incomes span more than one calendar year")

// LimitWarning reports that the year income exceeds the Small Business annual turnover limit.
// It does not change the tax, which is still calculated at the category rate.
type LimitWarning struct {
	Limit      models.Money
	YearIncome models.Money
	Excess     models.Money
}

func (w LimitWarning) String() string {
	return fmt.Sprintf(
		"Year income %s exceeds the Small Business limit of %s by %s. "+
			"Under Article 90(2) of the Tax Code, income from the start of the month the limit was exceeded "+
			"until the end of the year is taxed at %s: calculate it with %s. The tax above is calculated at %s.",
		w.YearIncome.String(), w.Limit.String(), w.Excess.String(),
		ratePercent(taxes.TaxTypeSmallBusinessOverLimit), taxes.TaxTypeSmallBusinessOverLimit.String(),
		ratePercent(taxes.TaxTypeSmallBusiness),
	)
}

// ratePercent formats the rate of tt as a percentage, e.g. "3%".
func ratePercent(tt taxes.TaxType) string {
	const toPercentage float64 = 100

	tr, err := tt.Rate()
	if err != nil {
		return tt.String()
	}

	return moneyutils.ToString(moneyutils.Multiply(tr.Rate, toPercentage)) + "%"
}

// newLimitWarning returns a warning when tt is Small Business and yearIncome (GEL) is above the limit.
func newLimitWarning(tt taxes.TaxType, yearIncome float64) *LimitWarning {
	if tt != taxes.TaxTypeSmallBusiness || yearIncome <= taxes.SmallBusinessTurnoverLimit {
		return nil
	}

	return &LimitWarning{
		Limit:      models.NewMoney(taxes.SmallBusinessTurnoverLimit, currencies.GEL),
		YearIncome: models.NewMoney(yearIncome, currencies.GEL),
		Excess:     models.NewMoney(moneyutils.Add(yearIncome, -taxes.SmallBusinessTurnoverLimit), currencies.GEL),
	}
}

// checkSingleYear checks that all incomes are dated in one calendar year.
func checkSingleYear(incomes []Income) error {
	var first int

	for i, inc := range incomes {
		year, err := dateutils.ParseYear(inc.Year)
		if err != nil {
			return fmt.Errorf("income %d: %w", i+1, err)
		}

		if i == 0 {
			first = year

			continue
		}

		if year != first {
			return fmt.Errorf("%d and %d: %w", first, year, ErrIncomesSpanYears)
		}
	}

	return nil
}
