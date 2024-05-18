package web_server

func (w *WebServer) registerRoutes() {
	dbGroup := w.Router.Group("/api/v1/db")
	dbGroup.Use(w.Auth.AuthMiddleware)

	// TODO: Get task status
	dbGroup.GET("/ready", w.Handler.IsDbReady)
	dbGroup.POST("", w.Handler.CreateDb)
}
