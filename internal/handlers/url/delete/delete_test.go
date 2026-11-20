package delete_test

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ossirdian/url_shortener/internal/handlers/url/delete"
	"github.com/ossirdian/url_shortener/internal/handlers/url/delete/mocks"
	"github.com/ossirdian/url_shortener/internal/repository"
)

func TestDelete_Success(t *testing.T) {
	const alias = "test_alias"
	deleteURL := mocks.NewDeleteURL(t)
	deleteURL.On("Delete", mock.Anything, alias).Return(nil).Once()

	router := chi.NewRouter()
	router.Delete("/{alias}", delete.New(testLogger(), deleteURL))
	req := httptest.NewRequest(http.MethodDelete, "/"+alias, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	deleteURL.AssertExpectations(t)
}

func TestDelete_NotFound(t *testing.T) {
	const alias = "missing_alias"
	deleteURL := mocks.NewDeleteURL(t)
	deleteURL.On("Delete", mock.Anything, alias).Return(repository.ErrNotFound).Once()

	router := chi.NewRouter()
	router.Delete("/{alias}", delete.New(testLogger(), deleteURL))
	req := httptest.NewRequest(http.MethodDelete, "/"+alias, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNotFound, rec.Code)
	deleteURL.AssertExpectations(t)
}

func TestDelete_InternalError(t *testing.T) {
	const alias = "test_alias"
	deleteURL := mocks.NewDeleteURL(t)
	deleteURL.On("Delete", mock.Anything, alias).Return(errors.New("database unavailable")).Once()

	router := chi.NewRouter()
	router.Delete("/{alias}", delete.New(testLogger(), deleteURL))
	req := httptest.NewRequest(http.MethodDelete, "/"+alias, nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	deleteURL.AssertExpectations(t)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
