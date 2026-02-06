package app

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"testSrv/internal/config"
	"testSrv/internal/http/handlers/create"
	"testSrv/internal/http/handlers/list"
	"testSrv/internal/http/handlers/read"
	"testSrv/internal/storage"

	"github.com/gorilla/mux"
)

type App struct {
	log     *slog.Logger
	server  *http.Server
	storage storage.ISubscriptionStorage
}

func New(log *slog.Logger, cfg config.ServerConfig, storage storage.ISubscriptionStorage) *App {

	router := mux.NewRouter()
	router.HandleFunc("/subscriptions/{uuid}", read.NewHandler(log, storage)).Methods("GET")
	router.HandleFunc("/subscriptions", list.NewHandler(log, storage)).Methods("GET")
	router.HandleFunc("/subscriptions", create.NewHandler(log, storage)).Methods("POST")
	//  Queries("filter", "{filter}").
	//r.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler)
	//router.HandleFunc("/products", ProductsHandler)
	//router.HandleFunc("/articles", ArticlesHandler)

	// run server
	server := &http.Server{
		Addr:         cfg.Addres,
		Handler:      router,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	/*if err = server.ListenAndServe(); err != nil {
		log.Error("failed to start server")
	}
	*/
	log.Info("server stoped")
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
			//log.Info("server is stoped")
			// will be printed in app.Start
		} else {
			log.Error("server error", slog.String("error", err.Error()))
		}
	}

}
