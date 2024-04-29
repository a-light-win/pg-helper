package web_server

func (w *WebServer) registerRoutes() {
	dbGroup := w.Router.Group("/api/v1/db")
	dbGroup.Use(w.Auth.AuthMiddleware)

	// TODO: Get task status
	// dbGroup.GET("/migrate/:taskId", s.Handler.MigrateDbStatus)
}
