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
// @Description Создает подписку для пользователя.
// @Description
// @Description **Параметры:**
// @Description - `user_id` - UUID пользователя (обязательный).
// @Description - `service_name` - название сервиса (обязательное).
// @Description - `price` - цена подписки за месяц в рублях (целое неотрицательное число, обязательное).
// @Description - `start_date` - дата начала подписки в формате `ММ‑ГГГГ` (например, `07‑2025`). Подписка действует с первого числа указанного месяца (обязательное, не может быть `01‑1970`).
// @Description - `end_date` - дата окончания подписки в формате `ММ‑ГГГГ` (необязательное). Если указано, подписка действует до первого числа указанного месяца. Должно быть не ранее `start_date` и не равно `01‑1970`.
// @Description
// @Description **Логика работы:**
// @Description - Если пользователь с указанным `user_id` не существует - возвращается ошибка.
// @Description - Если сервис с именем `service_name` не найден - возвращается ошибка.
// @Description - Если `end_date` не указано, подписка действует 30 дней.
// @Description
// @Description **Валидация:**
// @Description - `user_id` и `service_name` не могут быть пустыми.
// @Description - `price` должно быть целым неотрицательным числом.
// @Description - `start_date` обязательно, должно соответствовать формату `ММ‑ГГГГ` и не быть `01‑1970`.
// @Description - `end_date`, если указано, должно соответствовать формату `ММ‑ГГГГ`, не быть `01‑1970` и не предшествовать `start_date`.
// @Tags Подписки
// @Accept json
// @Param json body handlers.SubscriptionRequestDTO true "Данные новой подписки"
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
			endDate := time.Time(dto.StartDate).Add(time.Hour * 24 * 30)
			tmpCustomTime := handlers.CustomTime(endDate)
			dto.EndDate = &tmpCustomTime
		}

		tmpTime := time.Time(*dto.EndDate)
		subscription := models.Subscription{
			UserUUID:    dto.UserUUID,
			ServiceName: dto.ServiceName,
			Price:       dto.Price,
			StartDate:   time.Time(dto.StartDate),
			EndDate:     &tmpTime,
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

		tmpCustomTime := handlers.CustomTime(*createdSubscription.EndDate)
		createdDto := handlers.SubscriptionResponceDTO{
			UUID: createdSubscription.UUID,
			SubscriptionDTO: handlers.SubscriptionDTO{
				ServiceName: createdSubscription.ServiceName,
				UserUUID:    createdSubscription.UserUUID,
				Price:       createdSubscription.Price,
				StartDate:   handlers.CustomTime(createdSubscription.StartDate),
				EndDate:     &tmpCustomTime,
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
