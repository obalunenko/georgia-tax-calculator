package nbggovge

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func ratesFromFile(t testing.TB, path string) Rates {
	t.Helper()

	bytes, err := os.ReadFile(path)
	require.NoError(t, err)

	resp, err := unmarshalRatesResponse(bytes)
	require.NoError(t, err)

	return resp.Rates()
}
