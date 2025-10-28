package vegadns

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_GetRecordID(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("GET /1.0/records", func(rw http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Content-Type") != contentType {
			http.Error(rw,
				fmt.Sprintf("Content-Type header: got '%s', want '%s'",
					req.Header.Get("Content-Type"), contentType),
				http.StatusBadRequest)

			return
		}

		if req.URL.Query().Get("domain_id") != "1" {
			http.Error(rw, fmt.Sprintf("domain_id: got '%s', want '1'", req.URL.Query().Get("domain_id")), http.StatusBadRequest)

			return
		}

		fromTestData("records.json").ServeHTTP(rw, req)
	})

	recordID, err := client.GetRecordID(t.Context(), 1, "foo", "TXT")
	require.NoError(t, err)

	assert.Equal(t, 10, recordID)
}

func TestClient_CreateTXTRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("POST /1.0/records", func(rw http.ResponseWriter, req *http.Request) {
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

		data := map[string]string{
			"record_type": "TXT",
			"ttl":         "120",
			"domain_id":   "1",
			"name":        "foo.example.com",
			"value":       "txt",
		}

		for k, v := range data {
			err = checkFormData(req, k, v)
			if err != nil {
				http.Error(rw, err.Error(), http.StatusBadRequest)

				return
			}
		}

		rw.WriteHeader(http.StatusCreated)
	})

	err := client.CreateTXTRecord(t.Context(), 1, "foo.example.com", "txt", 120)
	require.NoError(t, err)
}

func TestClient_DeleteRecord(t *testing.T) {
	client, mux := setupTest(t)

	mux.HandleFunc("DELETE /1.0/records/2", func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	err := client.DeleteRecord(t.Context(), 2)
	require.NoError(t, err)
}
