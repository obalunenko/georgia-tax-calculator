package taxes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

func TestCalc(t *testing.T) {
	const taxTypeNotExist = TaxType(999)

	type args struct {
		income  models.Money
		taxType TaxType
	}

	tests := []struct {
		name    string
		args    args
		want    models.Money
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "small business",
			args: args{
				income:  models.NewMoney(100278.88, currencies.GEL),
				taxType: TaxTypeSmallBusiness,
			},
			want:    models.NewMoney(1002.79, currencies.GEL),
			wantErr: assert.NoError,
		},
		{
			name: "Individual Entrepreneur",
			args: args{
				income:  models.NewMoney(100278.88, currencies.GEL),
				taxType: TaxTypeIndividualEntrepreneur,
			},
			want:    models.NewMoney(3008.37, currencies.GEL),
			wantErr: assert.NoError,
		},
		{
			name: "Employment",
			args: args{
				income:  models.NewMoney(100278.88, currencies.GEL),
				taxType: TaxTypeEmployment,
			},
			want:    models.NewMoney(20055.78, currencies.GEL),
			wantErr: assert.NoError,
		},
		{
			name: "Not exist - error",
			args: args{
				income:  models.NewMoney(100278.88, currencies.GEL),
				taxType: taxTypeNotExist,
			},
			want:    models.Money{},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Calc(tt.args.income, tt.args.taxType)
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
