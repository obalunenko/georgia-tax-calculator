package nbggovge

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func ratesResponseFromFile(t testing.TB, path string) RatesResponse {
	t.Helper()

	bytes, err := os.ReadFile(path)
	require.NoError(t, err)

	resp, err := UnmarshalRatesResponse(bytes)
	require.NoError(t, err)

	return resp
}
