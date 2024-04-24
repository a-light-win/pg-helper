package web_server

import (
	"context"
	"errors"
	"net/http"

	"github.com/a-light-win/pg-helper/internal/config/server"
	"github.com/a-light-win/pg-helper/internal/validate"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

type WebServer struct {
	Config *server.WebConfig
	Router *gin.Engine
	Server *http.Server
}

func (w *WebServer) Init(config any) error {
	c, ok := config.(*server.WebConfig)
	if !ok {
		err := errors.New("failed to get server config")
		log.Error().Err(err).Msg("Failed to init web server")
		return err
	}

	w.Config = c
	w.Router = gin.Default()
	w.Server = &http.Server{
		Addr:    c.ListenOn(),
		Handler: w.Router,
	}

	validatorEngine, ok := binding.Validator.Engine().(*validator.Validate)
	if ok {
		validate.RegisterCustomValidations(validatorEngine)
	}

	err := w.initWebServerByConfig()
	if err != nil {
		return err
	}
	return w.registerRoutes()
}

func (w *WebServer) initWebServerByConfig() error {
	w.Router.UseH2C = w.Config.UseH2C
	w.Router.SetTrustedProxies(w.Config.TrustedProxies)
	return nil
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
