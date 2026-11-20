package delete

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/ossirdian/url_shortener/internal/repository"
	"github.com/ossirdian/url_shortener/internal/service"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type DeleteURL interface {
	Delete(ctx context.Context, alias string) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.53.8 --name=DeleteURL --output=mocks --outpkg=mocks
func New(log *slog.Logger, deleteURL DeleteURL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.delete.New"

		log := log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("empty alias")
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{Status: "Error", Error: "invalid request"})
			return
		}

		err := deleteURL.Delete(r.Context(), alias)
		if errors.Is(err, repository.ErrNotFound) {
			log.Info("url not found", slog.String("alias", alias))
			render.Status(r, http.StatusNotFound)
			render.JSON(w, r, Response{Status: "Error", Error: "url not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidArgument) {
			log.Info("invalid alias", slog.String("alias", alias))
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, Response{Status: "Error", Error: "invalid alias"})
			return
		}
		if err != nil {
			log.Error("failed to delete url", slog.String("alias", alias), slog.Any("error", err))
			render.Status(r, http.StatusInternalServerError)
			render.JSON(w, r, Response{Status: "Error", Error: "failed to delete url"})
			return
		}

		log.Info("deleted url", slog.String("alias", alias))
		responseOK(w, r)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, Response{Status: "OK"})
}
