package create

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testSrv/internal/domain/models"
	"testSrv/internal/domain/utils"
	"testSrv/internal/http/handlers"
	"testSrv/internal/storage"
	"time"
)

var (
	errValidateUserUUID    = errors.New("invalid user uuid")
	errValidateServiceName = errors.New("invalid service name")
	errValidatePrice       = errors.New("invalid price")
	errValidateStartDate   = errors.New("invalid start date")
)

func validateDTO(dto handlers.SubscriptionDTO) error {
	var err error = nil

	if dto.UserUUID == "" {
		err = errors.Join(err, errValidateUserUUID)
	}

	if dto.ServiceName == "" {
		err = errors.Join(err, errValidateServiceName)
	}

	if dto.Price < 0 {
		err = errors.Join(err, errValidatePrice)
	}

	if dto.StartDate.Time.IsZero() {
		err = errors.Join(err, errValidateStartDate)
	}

	return err
}

func NewHandler(log *slog.Logger, creator storage.ISubscriptionCreator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const logPrefix = "handlers.create.handler"
		log := log.With(
			slog.String("where", logPrefix),
		)

		log.Debug("Recive message", slog.String("method", r.Method), slog.String("url", r.URL.String()))

		dto := handlers.SubscriptionDTO{}

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

		if err := validateDTO(dto); err != nil {
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
			dto.EndDate = &endDate
		}

		subscription := models.Subscription{
			UUID:        utils.GenerateUUID(),
			UserUUID:    dto.UserUUID,
			ServiceName: dto.ServiceName,
			Price:       dto.Price,
			StartDate:   dto.StartDate.Time,
			EndDate:     dto.EndDate,
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
					Error: "subscription already exist",
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}
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

			}
		}

		createdDto := handlers.SubscriptionDTO{
			ServiceName: createdSubscription.ServiceName,
			UserUUID:    createdSubscription.UserUUID,
			Price:       createdSubscription.Price,
			StartDate:   handlers.CstomTime{Time: createdSubscription.StartDate},
			EndDate:     createdSubscription.EndDate,
		}

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(createdDto); err != nil {
			log.Error("Error Encode DTO")
			return
		}
	}
}
