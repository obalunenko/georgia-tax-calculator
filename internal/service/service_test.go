package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/obalunenko/georgia-tax-calculator/internal/converter"
	"github.com/obalunenko/georgia-tax-calculator/internal/models"
	"github.com/obalunenko/georgia-tax-calculator/internal/taxes"
	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge/currencies"
)

type mockConverter struct {
}

func (m mockConverter) Convert(_ context.Context, money models.Money, toCurrency string, _ time.Time) (converter.Response, error) {
	return converter.Response{
		Money: models.Money{
			Amount:   money.Amount,
			Currency: toCurrency,
		},
	}, nil
}

type mockConverterError struct{}

func (m mockConverterError) Convert(_ context.Context, money models.Money, toCurrency string, _ time.Time) (converter.Response, error) {
	return converter.Response{}, fmt.Errorf("mocked error")
}

func Test_service_Convert(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		c converter.Converter
	}

	type args struct {
		ctx context.Context
		p   ConvertRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ConvertResponse
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "correct request",
			fields: fields{
				c: mockConverter{},
			},
			args: args{
				ctx: ctx,
				p: ConvertRequest{
					DateRequest: DateRequest{
						Year:  "2022",
						Month: "December",
						Day:   "08",
					},
					CurrencyFrom: currencies.AED,
					CurrencyTo:   currencies.EUR,
					Amount:       "568",
				},
			},
			want: &ConvertResponse{
				Date: time.Date(2022, time.December, 8, 0, 0, 0, 0, time.Local),
				Amount: models.Money{
					Amount:   568,
					Currency: currencies.AED,
				},
				Converted: models.Money{
					Amount:   568,
					Currency: currencies.EUR,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "incorrect request",
			fields: fields{
				c: mockConverter{},
			},
			args: args{
				ctx: ctx,
				p: ConvertRequest{
					DateRequest: DateRequest{
						Year:  "2022",
						Month: "12",
						Day:   "08",
					},
					CurrencyFrom: currencies.AED,
					CurrencyTo:   currencies.EUR,
					Amount:       "568",
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "converter returns error ",
			fields: fields{
				c: mockConverterError{},
			},
			args: args{
				ctx: ctx,
				p: ConvertRequest{
					DateRequest: DateRequest{
						Year:  "2022",
						Month: "12",
						Day:   "08",
					},
					CurrencyFrom: currencies.AED,
					CurrencyTo:   currencies.EUR,
					Amount:       "568",
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service{
				c: tt.fields.c,
			}

			got, err := s.Convert(tt.args.ctx, tt.args.p)
			if tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_service_Calculate(t *testing.T) {
	ctx := context.Background()

	type fields struct {
		c converter.Converter
	}

	type args struct {
		ctx context.Context
		p   CalculateRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *CalculateResponse
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "correct request",
			fields: fields{
				c: mockConverter{},
			},
			args: args{
				ctx: ctx,
				p: CalculateRequest{
					DateRequest: DateRequest{
						Year:  "2022",
						Month: "December",
						Day:   "08",
					},
					Currency: currencies.EUR,
					Amount:   "1000",
					Taxtype:  taxes.TaxTypeEmployment.String(),
				},
			},
			want: &CalculateResponse{
				Date: time.Date(2022, time.December, 8, 0, 0, 0, 0, time.Local),
				TaxRate: taxes.TaxRate{
					Type: taxes.TaxTypeEmployment,
					Rate: 0.2,
				},
				Income: models.Money{
					Amount:   1000,
					Currency: currencies.EUR,
				},
				IncomeConverted: models.Money{
					Amount:   1000,
					Currency: currencies.GEL,
				},
				Tax: models.Money{
					Amount:   200,
					Currency: currencies.GEL,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "incorrect request",
			fields: fields{
				c: mockConverter{},
			},
			args: args{
				ctx: ctx,
				p: CalculateRequest{
					DateRequest: DateRequest{
						Year:  "2022",
						Month: "12",
						Day:   "08",
					},
					Currency: currencies.GEL,
					Amount:   "568",
					Taxtype:  taxes.TaxTypeSmallBusiness.String(),
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
		{
			name: "converter returns error ",
			fields: fields{
				c: mockConverterError{},
			},
			args: args{
				ctx: ctx,
				p: CalculateRequest{
					DateRequest: DateRequest{
						Year:  "2022",
						Month: "12",
						Day:   "08",
					},
					Currency: currencies.GEL,
					Amount:   "568",
					Taxtype:  taxes.TaxTypeSmallBusiness.String(),
				},
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := service{
				c: tt.fields.c,
			}

			got, err := s.Calculate(tt.args.ctx, tt.args.p)
			if tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}
