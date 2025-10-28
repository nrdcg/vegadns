package vegadns

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetDomainID(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("GET /1.0/domains", func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != contentType {
			http.Error(rw,
				fmt.Sprintf("Content-Type header: got '%s', want '%s'",
					req.Header.Get("Content-Type"), contentType),
				http.StatusBadRequest)

			return
		}

		if req.URL.Query().Get("search") != "example.com" {
			http.Error(rw, fmt.Sprintf("search: got '%s', want 'example.com'", req.URL.Query().Get("search")), http.StatusBadRequest)

			return
		}

		fromTestData("domains.json").ServeHTTP(rw, req)
	})

	domainID, err := client.GetDomainID(t.Context(), "example.com")
	require.NoError(t, err)

	assert.Equal(t, 1, domainID)
}

func TestClient_GetAuthZone(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("GET /1.0/domains", func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != contentType {
			http.Error(rw,
				fmt.Sprintf("Content-Type header: got '%s', want '%s'",
					req.Header.Get("Content-Type"), contentType),
				http.StatusBadRequest)

			return
		}

		if req.URL.Query().Get("search") != "example.com" {
			http.Error(rw, fmt.Sprintf("search: got '%s', want 'example.com'", req.URL.Query().Get("search")), http.StatusBadRequest)

			return
		}

		fromTestData("domains.json").ServeHTTP(rw, req)
	})

	zone, domainID, err := client.GetAuthZone(t.Context(), "foo.example.com")
	require.NoError(t, err)

	assert.Equal(t, 1, domainID)
	assert.Equal(t, "example.com", zone)
}
