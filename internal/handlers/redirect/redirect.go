package redirect

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ossirdian/url_shortener/internal/model"
	"github.com/ossirdian/url_shortener/internal/repository"
	"github.com/ossirdian/url_shortener/internal/service"
)

type URLGetter interface {
	Resolve(ctx context.Context, alias string) (model.URL, error)
}

func New(log *slog.Logger, urlGetter URLGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const fn = "handlers.url.redirect.New"

		log = log.With(
			slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		alias := chi.URLParam(r, "alias")
		if alias == "" {
			log.Error("empty alias")
			writeRedirectError(w, http.StatusBadRequest, "invalid request")
			return
		}

		url, err := urlGetter.Resolve(r.Context(), alias)
		if errors.Is(err, repository.ErrNotFound) {
			log.Info("url not found", "alias", alias)
			writeRedirectError(w, http.StatusNotFound, "url not found")
			return
		}
		if errors.Is(err, service.ErrInvalidArgument) {
			log.Info("invalid alias", "alias", alias)
			writeRedirectError(w, http.StatusBadRequest, "invalid alias")
			return
		}
		if err != nil {
			log.Error("failed to get url", "alias", alias)
			writeRedirectError(w, http.StatusInternalServerError, "failed to get url")
			return
		}

		log.Info("got url", slog.String("url", url.OriginalUrl))

		// redirect to the url
		http.Redirect(w, r, url.OriginalUrl, http.StatusFound)
	}
}

func writeRedirectError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "Error",
		"error":  message,
	})
}
