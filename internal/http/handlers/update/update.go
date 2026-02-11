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

// UpdateSubscription godoc
// @Summary Изменение подписки
// @Description Изменяет существующую подписку с указанным во входящем json(назовем входящий json как in) uuid.
// @Description Изменяет service_name на сервис с именем in.service_name.
// @Description Изменяет user_id на uuid пользователя in.user_id. В поле user_id на самом деле лежит UUID.
// @Description Изменяет price на in.price.
// @Description Изменяет start_date на in.start_date.
// @Description Изменяет end_date на in.end_date.
// @Description Если пользователя с in.user_id не существует, или не существует сервиса с именем in.service_name
// @Description то ничего не меняет и возвращает ошибку с указанной причиной.
// @Description Валидация: in.user_id, in.service_name не должны быть пустыми. in.price не отрицательное, целое число
// @Description in.start_date задано, имеет формат "ММ-ГГГГ" и отличается 01-1970 (т.е. не по умолчанию)
// @Description in.end_date может быть не задано, но если задано, имеет формат "ММ-ГГГГ" и отличается 01-1970 (т.е. не по умолчанию)
// @Description а также не может быть ранее in.start_date
// @Tags Подписки
// @Accept json
// @Produce json
// @Success 201 {object} handlers.SubscriptionResponceDTO "Созданная подписка"
// @Failure 400 {object} handlers.ErrorDTO "Неверный запрос, ошибка во входящем json, ошибка валидации json"
// @Failure 404 {object} handlers.ErrorDTO "Не найдена подписка с указанным in.uuid"
// @Failure 500 {object} handlers.ErrorDTO "Внутренняя ошибка работы сервера"
// @Router /subscriptions [patch]
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
			EndDate:     &changedDto.EndDate.Time,
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

		updatedDto := handlers.SubscriptionResponceDTO{
			UUID: updatedSubscription.UUID,
			SubscriptionDTO: handlers.SubscriptionDTO{
				ServiceName: updatedSubscription.ServiceName,
				UserUUID:    updatedSubscription.UserUUID,
				Price:       updatedSubscription.Price,
				StartDate:   handlers.CustomTime{Time: updatedSubscription.StartDate},
				EndDate:     &handlers.CustomTime{Time: *updatedSubscription.EndDate},
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
