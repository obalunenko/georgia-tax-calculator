package taxes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalc(t *testing.T) {
	const taxTypeNotExist = TaxType(999)

	type args struct {
		income  float64
		taxType TaxType
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "small business",
			args: args{
				income:  100278.88,
				taxType: TaxTypeSmallBusiness,
			},
			want:    1002.79,
			wantErr: assert.NoError,
		},
		{
			name: "Individual Entrepreneur",
			args: args{
				income:  100278.88,
				taxType: TaxTypeIndividualEntrepreneur,
			},
			want:    3008.37,
			wantErr: assert.NoError,
		},
		{
			name: "Employment",
			args: args{
				income:  100278.88,
				taxType: TaxTypeEmployment,
			},
			want:    20055.78,
			wantErr: assert.NoError,
		},
		{
			name: "Not exist - error",
			args: args{
				income:  100278.88,
				taxType: taxTypeNotExist,
			},
			want:    0,
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
