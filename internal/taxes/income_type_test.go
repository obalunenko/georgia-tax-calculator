package taxes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllIncomeTypes(t *testing.T) {
	type want struct {
		name  string
		field int
		label string
	}

	expected := []want{
		{name: "Cash", field: 18, label: "Cash (18)"},
		{name: "POS terminal", field: 19, label: "POS terminal (19)"},
		{name: "Non-cash", field: 20, label: "Non-cash (20)"},
		{name: "Other", field: 21, label: "Other (21)"},
		{name: "Unpaid supply", field: 22, label: "Unpaid supply (22)"},
	}

	got := AllIncomeTypes()
	require.Len(t, got, len(expected))

	for i, tt := range expected {
		t.Run(tt.name, func(t *testing.T) {
			it := got[i]

			assert.True(t, it.Valid())
			assert.Equal(t, tt.name, it.String())
			assert.Equal(t, tt.field, it.Field())
			assert.Equal(t, tt.label, it.Label())
			assert.NotEmpty(t, it.Description())
		})
	}
}

func TestIncomeTypeInvalid(t *testing.T) {
	for _, it := range []IncomeType{incomeTypeUnknown, incomeTypeSentinel} {
		assert.False(t, it.Valid())
		assert.Equal(t, 0, it.Field())
		assert.Empty(t, it.Label())
		assert.Empty(t, it.Description())
	}

	assert.Equal(t, "IncomeType(6)", incomeTypeSentinel.String())
}

func TestParseIncomeType(t *testing.T) {
	tests := []struct {
		raw     string
		want    IncomeType
		wantErr string
	}{
		{raw: "Cash", want: IncomeTypeCash},
		{raw: " Cash ", want: IncomeTypeCash},
		{raw: "non-cash", want: IncomeTypeNonCash},
		{raw: "POS TERMINAL", want: IncomeTypePOSTerminal},
		{raw: "Unpaid supply", want: IncomeTypeUnpaidSupply},
		{raw: "", wantErr: `"": invalid income type`},
		{raw: "Card", wantErr: `"Card": invalid income type`},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			got, err := ParseIncomeType(tt.raw)
			if tt.wantErr != "" {
				require.ErrorIs(t, err, ErrInvalidIncomeType)
				assert.EqualError(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIncomeTypeLabel(t *testing.T) {
	assert.Equal(t, "Non-cash (20)", IncomeTypeLabel("non-cash"))
	assert.Empty(t, IncomeTypeLabel(""))
	assert.Empty(t, IncomeTypeLabel("Card"))
}
