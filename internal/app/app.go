package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testSrv/internal/config"
	"testSrv/internal/http/handlers/cost"
	"testSrv/internal/http/handlers/create"
	"testSrv/internal/http/handlers/delete"
	"testSrv/internal/http/handlers/list"
	"testSrv/internal/http/handlers/read"
	"testSrv/internal/http/handlers/update"
	"testSrv/internal/storage"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type App struct {
	log     *slog.Logger
	server  *http.Server
	storage storage.ISubscriptionStorage
}

func New(log *slog.Logger, cfg config.Config, storage storage.ISubscriptionStorage) *App {

	router := mux.NewRouter()
	router.HandleFunc("/subscriptions", create.NewHandler(log, storage)).Methods("POST")
	router.HandleFunc("/subscriptions/{uuid}", read.NewHandler(log, storage)).Methods("GET")
	router.HandleFunc("/subscriptions", update.NewHandler(log, storage)).Methods("PATCH")
	router.HandleFunc("/subscriptions/{uuid}", delete.NewHandler(log, storage)).Methods("DELETE")
	router.HandleFunc("/subscriptions", list.NewHandler(log, storage)).Methods("GET")
	router.HandleFunc("/cost", cost.NewHandler(log, storage)).Methods("GET")

	if cfg.SwaggerConfig.NeedRuning {
		router.PathPrefix(cfg.SwaggerConfig.URL).Handler(httpSwagger.WrapHandler)
	}

	server := &http.Server{
		Addr:         cfg.Addres,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &App{
		log:     log,
		server:  server,
		storage: storage,
	}
}

func (a *App) Start() {
	const logPrefix = "app.Start"
	log := a.log.With(
		slog.String("where", logPrefix),
		slog.String("host", a.server.Addr),
	)

	log.Info("start server")
	if err := a.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			log.Info("server is stoped")
		} else {
			log.Error("server error", slog.String("error", err.Error()))
		}
	}
}

func (a *App) Stop() {
	const logPrefix = "app.Stop"
	log := a.log.With(
		slog.String("where", logPrefix),
	)

	log.Info("server stoping")

	if err := a.server.Shutdown(context.Background()); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			// log.Info("server is stoped")
			// will be printed in app.Start
		} else {
			log.Error("server error", slog.String("error", err.Error()))
		}
	}

}
