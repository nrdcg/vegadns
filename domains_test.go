package vegadns2client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetDomainID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"domains":[{"domain_id":1,"domain":"example.com","status":"active","owner_id":0}]}`))
	}))

	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithBasicAuth("user@example.com", "secret"), WithHTTPClient(server.Client()))
	require.NoError(t, err)

	domainID, err := client.GetDomainID(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, 1, domainID)
}
