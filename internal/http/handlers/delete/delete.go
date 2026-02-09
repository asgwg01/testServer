package delete

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testSrv/internal/http/handlers"
	"testSrv/internal/storage"

	"github.com/gorilla/mux"
)

func NewHandler(log *slog.Logger, deleter storage.ISubscriptionDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const logPrefix = "handlers.delete.handler"
		log := log.With(
			slog.String("where", logPrefix),
		)

		log.Debug("Recive message", slog.String("method", r.Method), slog.String("url", r.URL.String()))

		vars := mux.Vars(r)
		uuid := vars["uuid"]

		if uuid == "" {
			log.Error("uuid is empty")

			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")

			errDto := handlers.ErrorDTO{
				Error: fmt.Sprintf("subscription not found %s: %s", "uuid", uuid),
			}
			if err := json.NewEncoder(w).Encode(errDto); err != nil {
				log.Error("Error Encode DTO", slog.String("err", err.Error()))
				return
			}

			return
		}

		err := deleter.DeleteSubscription(uuid)
		if err != nil {
			if errors.Is(err, storage.ErrorSubscriptionNotFound) {

				log.Info("subscription not found", slog.String("uuid", uuid))

				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: fmt.Sprintf("subscription not found %s: %s", "uuid", uuid),
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}
			} else {
				log.Error("internal error", slog.String("err", err.Error()))

				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: fmt.Sprintf("error delete subscription uuid: %s", uuid),
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}
			}

			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
