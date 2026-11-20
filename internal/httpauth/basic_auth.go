package httpauth

import (
	"net/http"
)

func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			providedUsername, providedPassword, ok := r.BasicAuth()
			if !ok || providedUsername != username || providedPassword != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="url-shortener", charset="UTF-8"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
