package app

import (
	"net/http"
)

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", Health)

	mux.HandleFunc("/ready", ReadyHandler(s.readiness))

	var h http.Handler = mux
	h = Logging(s.logger)(h)
	h = RequestID(h)
	h = Recover(s.logger)(h)
	return h
}
