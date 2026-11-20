package httpauth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestBasicAuth_AcceptsEnvironmentCredentials(t *testing.T) {
	username, password := authCredentials(t)
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
	username, password := authCredentials(t)
	handler := BasicAuth(username, password)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		t.Fatal("next handler must not be called for unauthenticated request")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
	require.NotEmpty(t, rec.Header().Get("WWW-Authenticate"))
}

func authCredentials(t *testing.T) (string, string) {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "could not locate auth test file")
	envFile := filepath.Join(filepath.Dir(testFile), "..", "..", ".env")
	err := godotenv.Load(envFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		require.NoError(t, err)
	}

	username, usernameSet := os.LookupEnv("AUTH_USERNAME")
	password, passwordSet := os.LookupEnv("AUTH_PASSWORD")
	if !usernameSet || !passwordSet || username == "" || password == "" {
		t.Skip("AUTH_USERNAME and AUTH_PASSWORD must be set in .env or the service environment")
	}

	return username, password
}
