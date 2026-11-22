package httpauth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicAuth_AcceptsCorrectCredentials(t *testing.T) {
	username, password := testCredentials(t)
	handler := BasicAuth(username, password)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.SetBasicAuth(username, password)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
}

func TestBasicAuth_RejectsMissingCredentials(t *testing.T) {
	username, password := testCredentials(t)
	handler := BasicAuth(username, password)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler must not be called for unauthenticated request")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.NotEmpty(t, rec.Header().Get("WWW-Authenticate"))
}

func testCredentials(t *testing.T) (string, string) {
	t.Helper()
	t.Setenv("AUTH_USERNAME", "test-user")
	t.Setenv("AUTH_PASSWORD", "test-password")

	return os.Getenv("AUTH_USERNAME"), os.Getenv("AUTH_PASSWORD")
}
