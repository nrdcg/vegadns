package vegadns

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fromTestData(filename string) http.HandlerFunc {
	return func(rw http.ResponseWriter, _ *http.Request) {
		rw.Header().Set("Content-Type", "application/json")

		file, err := os.Open(filepath.Join("testdata", filename))
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)

			return
		}

		defer func() { _ = file.Close() }()

		rw.WriteHeader(http.StatusOK)

		_, err = io.Copy(rw, file)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)

			return
		}
	}
}

func checkFormData(req *http.Request, key, value string) error {
	if req.Form.Get(key) == value {
		return nil
	}

	return fmt.Errorf("%s: got '%s', want '%s'", key, req.Form.Get(key), value)
}

func setupTest(t *testing.T) (*Client, *http.ServeMux) {
	t.Helper()

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithHTTPClient(server.Client()))
	require.NoError(t, err)

	return client, mux
}

func TestClient_setAuth_oauth(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /1.0/token", fromTestData("token.json"))

	server := httptest.NewServer(mux)

	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithOAuth("user", "secret"), WithHTTPClient(server.Client()))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)

	err = client.setAuth(t.Context(), req)
	require.NoError(t, err)

	assert.Equal(t, "Bearer X123Y", req.Header.Get("Authorization"))
}

func TestClient_setAuth_oauth_existing_token(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /1.0/token", func(rw http.ResponseWriter, _ *http.Request) {
		// To be sure, there is no call to the token endpoint
		rw.WriteHeader(http.StatusUnauthorized)
	})

	server := httptest.NewServer(mux)

	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithOAuth("user", "secret"), WithHTTPClient(server.Client()))
	require.NoError(t, err)

	client.token = Token{
		Token:     "X123Z",
		TokenType: "bearer",
		ExpiresIn: 50,
		ExpiresAt: time.Now().Add(50 * time.Second),
	}

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)

	err = client.setAuth(t.Context(), req)
	require.NoError(t, err)

	assert.Equal(t, "Bearer X123Z", req.Header.Get("Authorization"))
}

func TestClient_setAuth_basic_auth(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /1.0/token", fromTestData("token.json"))

	server := httptest.NewServer(mux)

	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithBasicAuth("user", "secret"), WithHTTPClient(server.Client()))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "https://example.com", nil)

	err = client.setAuth(t.Context(), req)
	require.NoError(t, err)

	username, password, ok := req.BasicAuth()
	require.True(t, ok)

	assert.Equal(t, "user", username)
	assert.Equal(t, "secret", password)
}
