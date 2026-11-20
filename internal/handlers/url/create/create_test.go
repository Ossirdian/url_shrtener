package create_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ossirdian/url_shortener/internal/handlers/url/create"
	"github.com/ossirdian/url_shortener/internal/handlers/url/create/mocks"
	"github.com/ossirdian/url_shortener/internal/model"
	"github.com/ossirdian/url_shortener/internal/repository"
)

func TestCreate_Success(t *testing.T) {
	const (
		destinationURL = "https://example.com"
		alias          = "test_alias"
	)
	urlSaver := mocks.NewURLSaver(t)
	urlSaver.On("Create", mock.Anything, model.URLCreate{
		OriginalUrl: destinationURL,
		Alias:       alias,
	}).Return(model.URL{Alias: alias, OriginalUrl: destinationURL}, nil).Once()

	handler := create.New(testLogger(), urlSaver)
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(
		`{"url":"https://example.com","alias":"test_alias"}`,
	))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusCreated, rec.Code)
	var response create.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, "OK", response.Status)
	require.Equal(t, alias, response.Alias)
	urlSaver.AssertExpectations(t)
}

func TestCreate_BadRequest(t *testing.T) {
	urlSaver := mocks.NewURLSaver(t)
	handler := create.New(testLogger(), urlSaver)
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(`{"url":`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	urlSaver.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreate_Conflict(t *testing.T) {
	const (
		destinationURL = "https://example.com"
		alias          = "test_alias"
	)
	urlSaver := mocks.NewURLSaver(t)
	urlSaver.On("Create", mock.Anything, model.URLCreate{
		OriginalUrl: destinationURL,
		Alias:       alias,
	}).Return(model.URL{}, repository.ErrConflict).Once()

	handler := create.New(testLogger(), urlSaver)
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(
		`{"url":"https://example.com","alias":"test_alias"}`,
	))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusConflict, rec.Code)
	var response create.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	require.Equal(t, "url already exists", response.Error)
	urlSaver.AssertExpectations(t)
}

func TestCreate_InternalError(t *testing.T) {
	const (
		destinationURL = "https://example.com"
		alias          = "test_alias"
	)
	urlSaver := mocks.NewURLSaver(t)
	urlSaver.On("Create", mock.Anything, model.URLCreate{
		OriginalUrl: destinationURL,
		Alias:       alias,
	}).Return(model.URL{}, errors.New("database unavailable")).Once()

	handler := create.New(testLogger(), urlSaver)
	req := httptest.NewRequest(http.MethodPost, "/url/", bytes.NewBufferString(
		`{"url":"https://example.com","alias":"test_alias"}`,
	))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	urlSaver.AssertExpectations(t)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
