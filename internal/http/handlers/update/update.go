package update

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testSrv/internal/domain/models"
	"testSrv/internal/http/handlers"
	"testSrv/internal/storage"
)

func NewHandler(log *slog.Logger, updater storage.ISubscriptionUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const logPrefix = "handlers.update.handler"
		log := log.With(
			slog.String("where", logPrefix),
		)

		log.Debug("Recive message", slog.String("method", r.Method), slog.String("url", r.URL.String()))

		changedDto := handlers.SubscriptionFullDTO{}

		if err := json.NewDecoder(r.Body).Decode(&changedDto); err != nil {
			log.Error("Error Decode DTO", slog.String("err", err.Error()))

			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")

			errDto := handlers.ErrorDTO{
				Error: fmt.Sprintf("error in request body: %s", err.Error()),
			}
			if err := json.NewEncoder(w).Encode(errDto); err != nil {
				log.Error("Error Encode DTO", slog.String("err", err.Error()))
				return
			}

			return
		}

		if err := handlers.ValidateDTO(changedDto.SubscriptionDTO); err != nil {
			log.Error("Validation error", slog.String("err", err.Error()))

			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")

			errDto := handlers.ErrorDTO{
				Error: fmt.Sprintf("validation error: %s", err.Error()),
			}
			if err := json.NewEncoder(w).Encode(errDto); err != nil {
				log.Error("Error Encode DTO", slog.String("err", err.Error()))
				return
			}
		}

		subscription := models.Subscription{
			UUID:        changedDto.UUID,
			UserUUID:    changedDto.UserUUID,
			ServiceName: changedDto.ServiceName,
			Price:       changedDto.Price,
			StartDate:   changedDto.StartDate.Time,
			EndDate:     changedDto.EndDate,
		}

		updatedSubscription, err := updater.UpdateSubscription(subscription)
		if err != nil {
			if errors.Is(err, storage.ErrorSubscriptionNotFound) {
				log.Info("Subscription not found",
					slog.String("serviceName", subscription.ServiceName),
					slog.String("userUUID", subscription.UserUUID),
				)

				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: "subscription not found",
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}
			} else {
				log.Info("Error update subscription",
					slog.String("serviceName", subscription.ServiceName),
					slog.String("userUUID", subscription.UserUUID),
				)

				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: "error update subscription",
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}

			}
		}

		updatedDto := handlers.SubscriptionFullDTO{
			UUID: updatedSubscription.UUID,
			SubscriptionDTO: handlers.SubscriptionDTO{
				ServiceName: updatedSubscription.ServiceName,
				UserUUID:    updatedSubscription.UserUUID,
				Price:       updatedSubscription.Price,
				StartDate:   handlers.CstomTime{Time: updatedSubscription.StartDate},
				EndDate:     updatedSubscription.EndDate,
			},
		}

		w.WriteHeader(http.StatusAccepted)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(updatedDto); err != nil {
			log.Error("Error Encode DTO")
			return
		}
	}
}
