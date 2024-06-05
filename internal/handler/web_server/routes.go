package web_server

func (w *WebServer) registerRoutes() {
	dbGroup := w.Router.Group("/api/v1/db")
	dbGroup.Use(w.Auth.AuthMiddleware)

	dbHandler := NewDbHandler(w.sourceHandler, w.dbReadyWaiter)

	// TODO: Get task status
	dbGroup.GET("/ready", WebHandleWrapper(dbHandler, NewIsDbReadyRequest))
	dbGroup.POST("", WebHandleWrapper(dbHandler, NewCreateDbRequest))
}
