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

type mockConverter struct{}

func (m mockConverter) Convert(_ context.Context, money models.Money, toCurrency string, _ time.Time) (converter.Response, error) {
	return converter.Response{
		Money: models.Money{
			Amount:   money.Amount,
			Currency: toCurrency,
		},
		Rate: 1,
	}, nil
}

type mockConverterError struct{}

func (m mockConverterError) Convert(_ context.Context, _ models.Money, _ string, _ time.Time) (converter.Response, error) {
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
				Date: time.Date(2022, time.December, 8, 0, 0, 0, 0, time.UTC),
				Amount: models.Money{
					Amount:   568,
					Currency: currencies.AED,
				},
				Converted: models.Money{
					Amount:   568,
					Currency: currencies.EUR,
				},
				Rate: models.NewMoney(1, ""),
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
			if !tt.wantErr(t, err) {
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
			name: "correct request with one income",
			fields: fields{
				c: mockConverter{},
			},
			args: args{
				ctx: ctx,
				p: CalculateRequest{
					Income: []Income{
						{
							DateRequest: DateRequest{
								Year:  "2022",
								Month: "December",
								Day:   "08",
							},
							Currency: currencies.EUR,
							Amount:   "1000",
						},
					},

					TaxType:    taxes.TaxTypeEmployment.String(),
					YearIncome: "67.99",
				},
			},
			want: &CalculateResponse{
				TaxRate: taxes.TaxRate{
					Type: taxes.TaxTypeEmployment,
					Rate: 0.2,
				},
				YearIncome: models.Money{
					Amount:   1067.99,
					Currency: currencies.GEL,
				},
				Incomes: []ConvertResponse{
					{
						Date: time.Date(2022, time.December, 8, 0, 0, 0, 0, time.UTC),
						Amount: models.Money{
							Amount:   1000,
							Currency: currencies.EUR,
						},
						Converted: models.Money{
							Amount:   1000,
							Currency: currencies.GEL,
						},
						Rate: models.NewMoney(1, ""),
					},
				},
				TotalIncomeConverted: models.Money{
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
			name: "correct request with multiple income",
			fields: fields{
				c: mockConverter{},
			},
			args: args{
				ctx: ctx,
				p: CalculateRequest{
					Income: []Income{
						{
							DateRequest: DateRequest{
								Year:  "2022",
								Month: "December",
								Day:   "08",
							},
							Currency: currencies.EUR,
							Amount:   "1000",
						},
						{
							DateRequest: DateRequest{
								Year:  "2023",
								Month: "June",
								Day:   "08",
							},
							Currency: currencies.USD,
							Amount:   "200",
						},
					},
					TaxType:    taxes.TaxTypeEmployment.String(),
					YearIncome: "67.99",
				},
			},
			want: &CalculateResponse{
				TaxRate: taxes.TaxRate{
					Type: taxes.TaxTypeEmployment,
					Rate: 0.2,
				},
				YearIncome: models.Money{
					Amount:   1267.99,
					Currency: currencies.GEL,
				},
				Incomes: []ConvertResponse{
					{
						Date: time.Date(2022, time.December, 8, 0, 0, 0, 0, time.UTC),
						Amount: models.Money{
							Amount:   1000,
							Currency: currencies.EUR,
						},
						Converted: models.Money{
							Amount:   1000,
							Currency: currencies.GEL,
						},
						Rate: models.NewMoney(1, ""),
					},
					{
						Date: time.Date(2023, time.June, 8, 0, 0, 0, 0, time.UTC),
						Amount: models.Money{
							Amount:   200,
							Currency: currencies.USD,
						},
						Converted: models.Money{
							Amount:   200,
							Currency: currencies.GEL,
						},
						Rate: models.NewMoney(1, ""),
					},
				},
				TotalIncomeConverted: models.Money{
					Amount:   1200,
					Currency: currencies.GEL,
				},
				Tax: models.Money{
					Amount:   240,
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
					Income: []Income{
						{
							DateRequest: DateRequest{
								Year:  "2022",
								Month: "12",
								Day:   "08",
							},
							Currency: currencies.GEL,
							Amount:   "568",
						},
					},

					TaxType:    taxes.TaxTypeSmallBusiness.String(),
					YearIncome: "0",
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
					Income: []Income{
						{
							DateRequest: DateRequest{
								Year:  "2022",
								Month: "12",
								Day:   "08",
							},
							Currency: currencies.GEL,
							Amount:   "568",
						},
					},

					TaxType:    taxes.TaxTypeSmallBusiness.String(),
					YearIncome: "0",
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
			if !tt.wantErr(t, err) {
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCalculateResponse_String(t *testing.T) {
	type fields struct {
		CalculateResponse
	}

	tests := []struct {
		name   string
		fields CalculateResponse
		want   string
	}{
		{
			name: "one income",
			fields: CalculateResponse{
				TaxRate: taxes.TaxRate{
					Type: taxes.TaxTypeEmployment,
					Rate: 0.2,
				},
				YearIncome: models.Money{
					Amount:   0,
					Currency: currencies.GEL,
				},
				Incomes: []ConvertResponse{
					{
						Date: time.Date(2022, time.December, 8, 0, 0, 0, 0, time.UTC),
						Amount: models.Money{
							Amount:   568.99,
							Currency: currencies.AED,
						},
						Converted: models.Money{
							Amount:   789.99,
							Currency: currencies.EUR,
						},
						Rate: models.NewMoney(1.39, ""),
					},
				},
				TotalIncomeConverted: models.Money{
					Amount:   789.99,
					Currency: currencies.EUR,
				},
				Tax: models.Money{
					Amount:   99.02,
					Currency: currencies.AMD,
				},
			},
			want: "Tax Rate: Employment 20 %\n" +
				"Year Income: 0 GEL\n" +
				"Incomes:\n" +
				"\t- 1:\n" +
				"\t\tDate: 2022-12-08\n" +
				"\t\tAmount: 568.99 AED\n" +
				"\t\tConverted: 789.99 EUR\n" +
				"\t\tRate: 1.39\n" +
				"Total Income Converted: 789.99 EUR\n" +
				"Taxes: 99.02 AMD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.fields.String(), "String()")
		})
	}
}

func TestConvertResponse_String(t *testing.T) {
	type fields struct {
		Date      time.Time
		Amount    models.Money
		Converted models.Money
		Rate      models.Money
	}

	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "",
			fields: fields{
				Date: time.Date(2022, time.December, 8, 0, 0, 0, 0, time.UTC),
				Amount: models.Money{
					Amount:   568.99,
					Currency: currencies.AED,
				},
				Converted: models.Money{
					Amount:   789.99,
					Currency: currencies.EUR,
				},
				Rate: models.NewMoney(1.39, ""),
			},
			want: "Date: 2022-12-08\n" +
				"Amount: 568.99 AED\n" +
				"Converted: 789.99 EUR\n" +
				"Rate: 1.39",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := ConvertResponse{
				Date:      tt.fields.Date,
				Amount:    tt.fields.Amount,
				Converted: tt.fields.Converted,
				Rate:      tt.fields.Rate,
			}

			assert.Equalf(t, tt.want, c.String(), "String()")
		})
	}
}
