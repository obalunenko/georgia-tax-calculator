package nbggovge_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/obalunenko/georgia-tax-calculator/pkg/nbggovge"
)

func ratesResponseFromFile(t testing.TB, path string) nbggovge.RatesResponse {
	t.Helper()

	bytes, err := os.ReadFile(path)
	require.NoError(t, err)

	resp, err := nbggovge.UnmarshalRatesResponse(bytes)
	require.NoError(t, err)

	return resp
}
