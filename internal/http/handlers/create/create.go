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

// CreateSubscription godoc
// @Summary Создание новой подписки
// @Description Создает для пользователя user_id (на самом деле передается UUID)
// @Description подписку на сервис с именем service_name.
// @Description price - цена подписки за месяц, указывается целое число рублей
// @Description Подписка действует с первого числа месяца указанного в start_date. Формат "ММ-ГГГГ" (например 07-2025).
// @Description По умолчанию подписка дейтвует 30 дней, можно указать end_date в таком же формате - "ММ-ГГГГ",
// @Description тогда подписка будет действовать до первого числа указанного в end_date месяца. end_date - не обязательный параметр.
// @Description Если пользователя с таким uuid не существует, или не существует сервиса с именем service_name
// @Description то ничего не создает и возвращает ошибку с указанной причиной.
// @Description Валидация: user_id, service_name не должны быть пустыми. price не отрицательное, целое число
// @Description start_date задано, имеет формат "ММ-ГГГГ" и отличается 01-1970 (т.е. не по умолчанию)
// @Description end_date может быть не задано, но если задано, имеет формат "ММ-ГГГГ" и отличается 01-1970 (т.е. не по умолчанию)
// @Description а также не может быть ранее start_date
// @Tags Подписки
// @Accept json
// @Produce json
// @Success 201 {object} handlers.SubscriptionResponceDTO "Созданная подписка"
// @Failure 400 {object} handlers.ErrorDTO "Неверный запрос, ошибка во входящем json, ошибка валидации json"
// @Failure 409 {object} handlers.ErrorDTO "У пользователя уже существует подписка на сервис и время действия подписок пересекается"
// @Failure 404 {object} handlers.ErrorDTO "Не найден пользователем с указанным user_id или сервис с именем service_name"
// @Failure 500 {object} handlers.ErrorDTO "Внутренняя ошибка работы сервера"
// @Router /subscriptions [post]
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

		createdDto := handlers.SubscriptionResponceDTO{
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
