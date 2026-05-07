package app

import (
	"net/http"
)

func Recover(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered", "error", recovered, "method", r.Method, "path", r.URL.Path)
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("{\"status\":\"error\"}\n"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
