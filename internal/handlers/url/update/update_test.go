package update_test

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ossirdian/url_shortener/internal/handlers/url/update"
	"github.com/ossirdian/url_shortener/internal/handlers/url/update/mocks"
	"github.com/ossirdian/url_shortener/internal/model"
	"github.com/ossirdian/url_shortener/internal/repository"
)

func TestUpdate_Success(t *testing.T) {
	const oldAlias, newAlias = "old_alias", "new_alias"
	updatedURL := model.URL{Alias: newAlias, OriginalUrl: "https://example.com"}
	updateURL := mocks.NewUpdateURL(t)
	updateURL.On("Update", mock.Anything, model.URLUpdate{
		OldAlias: oldAlias,
		NewAlias: newAlias,
	}).Return(updatedURL, nil).Once()

	router := chi.NewRouter()
	router.Patch("/{alias}", update.New(testLogger(), updateURL))
	req := httptest.NewRequest(http.MethodPatch, "/"+oldAlias, strings.NewReader(
		`{"newAlias":"new_alias"}`,
	))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	var response update.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, "OK", response.Status)
	require.Equal(t, newAlias, response.Alias)
	require.Equal(t, updatedURL.OriginalUrl, response.URL)
	updateURL.AssertExpectations(t)
}

func TestUpdate_BadRequest(t *testing.T) {
	updateURL := mocks.NewUpdateURL(t)
	router := chi.NewRouter()
	router.Patch("/{alias}", update.New(testLogger(), updateURL))
	req := httptest.NewRequest(http.MethodPatch, "/old_alias", strings.NewReader(`{"newAlias":`))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	updateURL.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestUpdate_Conflict(t *testing.T) {
	updateURL := mocks.NewUpdateURL(t)
	updateURL.On("Update", mock.Anything, model.URLUpdate{
		OldAlias: "old_alias",
		NewAlias: "new_alias",
	}).Return(model.URL{}, repository.ErrConflict).Once()

	router := chi.NewRouter()
	router.Patch("/{alias}", update.New(testLogger(), updateURL))
	req := httptest.NewRequest(http.MethodPatch, "/old_alias", strings.NewReader(
		`{"newAlias":"new_alias"}`,
	))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	updateURL.AssertExpectations(t)
}

func TestUpdate_InternalError(t *testing.T) {
	updateURL := mocks.NewUpdateURL(t)
	updateURL.On("Update", mock.Anything, model.URLUpdate{
		OldAlias: "old_alias",
		NewAlias: "new_alias",
	}).Return(model.URL{}, errors.New("database unavailable")).Once()

	router := chi.NewRouter()
	router.Patch("/{alias}", update.New(testLogger(), updateURL))
	req := httptest.NewRequest(http.MethodPatch, "/old_alias", strings.NewReader(
		`{"newAlias":"new_alias"}`,
	))
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	updateURL.AssertExpectations(t)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
