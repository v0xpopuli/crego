package app

func (s *Server) registerRoutes() {
	s.app.Use(Recover(s.logger))
	s.app.Use(RequestID)
	s.app.Use(Logging(s.logger))

	s.app.Get("/health", Health)

	s.app.Get("/ready", ReadyHandler(s.readiness))

}
