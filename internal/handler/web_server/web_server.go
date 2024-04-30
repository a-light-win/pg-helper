package web_server

import (
	"context"
	"net/http"

	"github.com/a-light-win/pg-helper/internal/config/server"
	ginAuth "github.com/a-light-win/pg-helper/pkg/auth/gin"
	"github.com/a-light-win/pg-helper/pkg/validate"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type WebServer struct {
	Config *server.WebConfig

	Server *http.Server
	Router *gin.Engine
	Auth   *ginAuth.GinAuth
}

func NewWebServer(config *server.WebConfig) *WebServer {
	w := &WebServer{
		Config: config,
		Router: gin.Default(),
	}

	w.Server = &http.Server{
		Addr:    config.ListenOn(),
		Handler: w.Router,
	}

	w.Auth = ginAuth.NewGinAuth(&config.Auth)

	validatorEngine, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		validate.RegisterCustomValidations(validatorEngine)
	}

	w.Router.UseH2C = w.Config.UseH2C
	w.Router.SetTrustedProxies(w.Config.TrustedProxies)

	w.registerRoutes()

	return w
}

func (w *WebServer) Run() {
	if !w.Config.Enabled {
		log.Log().Msg("Web server is disabled")
		return
	}

	log.Log().Msg("Start the web server")

	if w.Config.Tls.Enabled {
		err := w.Server.ListenAndServeTLS(w.Config.Tls.ServerCert, w.Config.Tls.ServerKey)
		if err != nil {
			log.Error().Err(err).Msg("Web server exit with error")
		}
		return
	}

	err := w.Server.ListenAndServe()
	if err != nil {
		log.Error().Err(err).Msg("Web server exit with error")
	}
}

func (w *WebServer) Shutdown(ctx context.Context) error {
	return w.Server.Shutdown(ctx)
}
