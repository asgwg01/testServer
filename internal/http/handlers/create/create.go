package create

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testSrv/internal/domain/models"
	"testSrv/internal/http/handlers"
	"testSrv/internal/storage"
	"time"
)

func NewHandler(log *slog.Logger, creator storage.ISubscriptionCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const logPrefix = "handlers.create.handler"
		log := log.With(
			slog.String("where", logPrefix),
		)

		log.Debug("Recive message", slog.String("method", r.Method), slog.String("url", r.URL.String()))

		dto := handlers.SubscriptionRequestDTO{}

		if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
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

		if err := handlers.ValidateDTO(dto); err != nil {
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

		if dto.EndDate == nil {
			// По умолчанию подписка на месяц
			endDate := dto.StartDate.Time.Add(time.Hour * 24 * 30)
			dto.EndDate = &handlers.CustomTime{Time: endDate}
		}

		subscription := models.Subscription{
			UserUUID:    dto.UserUUID,
			ServiceName: dto.ServiceName,
			Price:       dto.Price,
			StartDate:   dto.StartDate.Time,
			EndDate:     &dto.EndDate.Time,
		}

		createdSubscription, err := creator.CreateSubscription(subscription)
		if err != nil {
			if errors.Is(err, storage.ErrorSubscriptionAlreadyExist) {
				log.Info("Subscription already exist",
					slog.String("serviceName", subscription.ServiceName),
					slog.String("userUUID", subscription.UserUUID),
				)

				w.WriteHeader(http.StatusConflict)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: fmt.Sprintf("subscription already exist, exist subscription uuid: %s", createdSubscription.UUID),
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}
				return
			} else if errors.Is(err, storage.ErrorServiceNotExist) {
				log.Info("Service is not exist",
					slog.String("serviceName", subscription.ServiceName),
				)

				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: fmt.Sprintf("service is not exist, service name: %s", subscription.ServiceName),
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}
				return
			} else if errors.Is(err, storage.ErrorUserNotExist) {
				log.Info("User is not exist",
					slog.String("user uuid", subscription.UserUUID),
				)

				w.WriteHeader(http.StatusNotFound)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: fmt.Sprintf("user is not exist, user uuid: %s", subscription.UserUUID),
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}
				return
			} else {
				log.Info("Error save subscription",
					slog.String("serviceName", subscription.ServiceName),
					slog.String("userUUID", subscription.UserUUID),
				)

				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: "error save subscription",
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}
				return
			}
		}

		createdDto := handlers.SubscriptionFullDTO{
			UUID: createdSubscription.UUID,
			SubscriptionDTO: handlers.SubscriptionDTO{
				ServiceName: createdSubscription.ServiceName,
				UserUUID:    createdSubscription.UserUUID,
				Price:       createdSubscription.Price,
				StartDate:   handlers.CustomTime{Time: createdSubscription.StartDate},
				EndDate:     &handlers.CustomTime{Time: *createdSubscription.EndDate},
			},
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(createdDto); err != nil {
			log.Error("Error Encode DTO")
			return
		}
	}
}
