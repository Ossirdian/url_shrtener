package create

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/ossirdian/url_shortener/internal/model"
	"github.com/ossirdian/url_shortener/internal/repository"
	"github.com/ossirdian/url_shortener/internal/service"
)

type Request struct {
	URL   string `json:"url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	Status string `json:"status"`
	Alias  string `json:"alias,omitempty"`
	Error  string `json:"error,omitempty"`
}

type URLSaver interface {
	Create(ctx context.Context, input model.URLCreate) (model.URL, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.8 --name=URLSaver --output=mocks --outpkg=mocks
func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.create.New"

		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var request Request
		if err := render.DecodeJSON(r.Body, &request); err != nil {
			log.Error("failed to decode request body", slog.Any("error", err))
			responseError(w, r, http.StatusBadRequest, "failed to decode request body")
			return
		}

		log.Info("request body decoded", slog.Any("request", request))
		createdURL, err := urlSaver.Create(r.Context(), model.URLCreate{
			OriginalUrl: request.URL,
			Alias:       request.Alias,
		})
		if errors.Is(err, service.ErrInvalidURL) {
			log.Info("invalid url", slog.String("url", request.URL))
			responseError(w, r, http.StatusBadRequest, "invalid url")
			return
		}
		if errors.Is(err, service.ErrInvalidArgument) {
			log.Info("invalid alias", slog.String("alias", request.Alias))
			responseError(w, r, http.StatusBadRequest, "invalid alias")
			return
		}
		if errors.Is(err, repository.ErrConflict) {
			log.Info("alias already exists", slog.String("alias", request.Alias))
			responseError(w, r, http.StatusConflict, "url already exists")
			return
		}
		if err != nil {
			log.Error("failed to create url", slog.Any("error", err))
			responseError(w, r, http.StatusInternalServerError, "failed to create url")
			return
		}

		log.Info("url created", slog.String("alias", createdURL.Alias))
		responseOK(w, r, createdURL.Alias)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, alias string) {
	render.Status(r, http.StatusCreated)
	render.JSON(w, r, Response{
		Status: "OK",
		Alias:  alias,
	})
}

func responseError(w http.ResponseWriter, r *http.Request, status int, message string) {
	render.Status(r, status)
	render.JSON(w, r, Response{
		Status: "Error",
		Error:  message,
	})
}
