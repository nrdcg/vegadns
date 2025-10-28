package vegadns

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_getAuthToken(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /1.0/token", func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != contentType {
			http.Error(rw,
				fmt.Sprintf("Content-Type header: got '%s', want '%s'",
					req.Header.Get("Content-Type"), contentType),
				http.StatusBadRequest)

			return
		}

		err := req.ParseForm()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		err = checkFormData(req, "grant_type", "client_credentials")
		if err != nil {
			http.Error(rw, err.Error(), http.StatusBadRequest)

			return
		}

		fromTestData("token.json").ServeHTTP(rw, req)
	})

	server := httptest.NewServer(mux)

	t.Cleanup(server.Close)

	client, err := NewClient(server.URL, WithOAuth("user", "secret"), WithHTTPClient(server.Client()))
	require.NoError(t, err)

	token, err := client.getAuthToken(t.Context())
	require.NoError(t, err)

	assert.NoError(t, token.valid())
}

func TestToken_valid(t *testing.T) {
	testCases := []struct {
		desc    string
		token   Token
		require require.ErrorAssertionFunc
	}{
		{
			desc:    "empty token",
			require: require.Error,
		},
		{
			desc: "valid",
			token: Token{
				ExpiresAt: time.Now().Add(1 * time.Second),
			},
			require: require.NoError,
		},
	}

	for _, test := range testCases {
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			test.require(t, test.token.valid())
		})
	}
}
