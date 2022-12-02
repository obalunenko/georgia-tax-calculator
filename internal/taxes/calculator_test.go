package taxes

import (
	"fmt"
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
		want    Response
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "small business",
			args: args{
				income:  models.NewMoney(100278.88, currencies.GEL),
				taxType: TaxTypeSmallBusiness,
			},
			want: Response{
				Money: models.NewMoney(1002.79, currencies.GEL),
				Rate: TaxRate{
					Type: TaxTypeSmallBusiness,
					Rate: 0.01,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Individual Entrepreneur",
			args: args{
				income:  models.NewMoney(100278.88, currencies.GEL),
				taxType: TaxTypeIndividualEntrepreneur,
			},
			want: Response{
				Money: models.NewMoney(3008.37, currencies.GEL),
				Rate: TaxRate{
					Type: TaxTypeIndividualEntrepreneur,
					Rate: 0.03,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Employment",
			args: args{
				income:  models.NewMoney(100278.88, currencies.GEL),
				taxType: TaxTypeEmployment,
			},
			want: Response{
				Money: models.NewMoney(20055.78, currencies.GEL),
				Rate: TaxRate{
					Type: TaxTypeEmployment,
					Rate: 0.2,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "Not exist - error",
			args: args{
				income:  models.NewMoney(100278.88, currencies.GEL),
				taxType: taxTypeNotExist,
			},
			want:    Response{},
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

func TestAllTaxTypes(t *testing.T) {
	expected := []TaxType{
		TaxTypeSmallBusiness, TaxTypeIndividualEntrepreneur, TaxTypeEmployment,
	}

	assert.ElementsMatchf(t, expected, AllTaxTypes(), "AllTaxTypes()")
}

func TestTaxType_Rate(t *testing.T) {
	tests := []struct {
		name    string
		i       TaxType
		want    TaxRate
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "Employment - 20%",
			i:       TaxTypeEmployment,
			want:    TaxRate{Type: TaxTypeEmployment, Rate: 0.2},
			wantErr: assert.NoError,
		},
		{
			name:    "Small business - 1%",
			i:       TaxTypeSmallBusiness,
			want:    TaxRate{Type: TaxTypeSmallBusiness, Rate: 0.01},
			wantErr: assert.NoError,
		},
		{
			name:    "Individual Entrepreneur - 3%",
			i:       TaxTypeIndividualEntrepreneur,
			want:    TaxRate{Type: TaxTypeIndividualEntrepreneur, Rate: 0.03},
			wantErr: assert.NoError,
		},
		{
			name:    "Not supported - error",
			i:       TaxType(0),
			want:    TaxRate{},
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.i.Rate()
			if !tt.wantErr(t, err, fmt.Sprintf("Rate()")) {
				return
			}

			assert.Equalf(t, tt.want, got, "Rate()")
		})
	}
}
