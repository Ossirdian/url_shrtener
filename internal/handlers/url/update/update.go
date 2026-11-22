package update

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/ossirdian/url_shortener/internal/model"
	"github.com/ossirdian/url_shortener/internal/repository"
	"github.com/ossirdian/url_shortener/internal/service"
)

type Request struct {
	OldAlias string `json:"oldAlias,omitempty"`
	NewAlias string `json:"newAlias"`
}

type Response struct {
	Status string `json:"status"`
	Alias  string `json:"alias,omitempty"`
	URL    string `json:"url,omitempty"`
	Error  string `json:"error,omitempty"`
}

type UpdateURL interface {
	Update(ctx context.Context, input model.URLUpdate) (model.URL, error)
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.8 --name=UpdateURL --output=mocks --outpkg=mocks
func New(log *slog.Logger, updateURL UpdateURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.update.New"

		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var request Request
		if err := render.DecodeJSON(r.Body, &request); err != nil {
			log.Error("failed to decode request", slog.Any("error", err))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{Status: "Error", Error: "invalid request"})
			return
		}

		alias := request.OldAlias
		if alias == "" {
			alias = chi.URLParam(r, "alias")
		}
		if alias == "" {
			log.Error("empty old alias")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{Status: "Error", Error: "invalid request"})
			return
		}
		if request.NewAlias == "" {
			log.Error("empty new alias", slog.String("alias", alias))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{Status: "Error", Error: "invalid alias"})
			return
		}

		updatedURL, err := updateURL.Update(r.Context(), model.URLUpdate{
			OldAlias: alias,
			NewAlias: request.NewAlias,
		})
		if errors.Is(err, service.ErrInvalidArgument) {
			log.Info("invalid alias", slog.String("alias", alias), slog.String("new_alias", request.NewAlias))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{Status: "Error", Error: "invalid alias"})
			return
		}
		if errors.Is(err, repository.ErrNotFound) {
			log.Info("url not found", slog.String("alias", alias))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, Response{Status: "Error", Error: "url not found"})
			return
		}
		if errors.Is(err, repository.ErrConflict) {
			log.Info("alias already exists", slog.String("new_alias", request.NewAlias))
			render.Status(r, http.StatusConflict)
			render.JSON(w, r, Response{Status: "Error", Error: "alias already exists"})
			return
		}
		if err != nil {
			log.Error("failed to update url", slog.String("alias", alias), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{Status: "Error", Error: "failed to update url"})
			return
		}

		log.Info("updated url",
			slog.String("old_alias", alias),
			slog.String("new_alias", updatedURL.Alias),
		)
		responseOK(w, r, updatedURL)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, updatedURL model.URL) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, Response{
		Status: "OK",
		Alias:  updatedURL.Alias,
		URL:    updatedURL.OriginalUrl,
	})
}
