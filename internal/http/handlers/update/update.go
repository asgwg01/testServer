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
	"time"
)

// UpdateSubscription godoc
// @Summary Изменение подписки
// @Description Изменяет существующую подписку.
// @Description
// @Description **Входные данные (JSON‑объект `in`):**
// @Description - `uuid` - идентификатор подписки, которую необходимо изменить (обязательный).
// @Description - `user_id` - UUID пользователя (обязательный, не может быть пустым).
// @Description - `service_name` - новое название сервиса (обязательное, не может быть пустым).
// @Description - `price` - новая цена подписки за месяц в рублях (целое неотрицательное число, обязательное).
// @Description - `start_date` - новая дата начала подписки в формате `ММ‑ГГГГ` (например, `07‑2025`). Подписка действует с первого числа указанного месяца (обязательное, не может быть `01‑1970`).
// @Description - `end_date` - новая дата окончания подписки в формате `ММ‑ГГГГ` (необязательное). Если указано, подписка действует до первого числа указанного месяца. Должно быть не ранее `start_date` и не равно `01‑1970`.
// @Description
// @Description **Логика работы:**
// @Description - Если подписка с указанным `uuid` не найдена - возвращается ошибка.
// @Description - Если пользователь с указанным `in.user_id` не существует - возвращается ошибка.
// @Description - Если сервис с именем `in.service_name` не найден - возвращается ошибка.
// @Description - Все указанные поля заменяют соответствующие значения в существующей подписке.
// @Description
// @Description **Валидация:**
// @Description - `in.user_id` и `in.service_name` не могут быть пустыми.
// @Description - `in.price` должно быть целым неотрицательным числом.
// @Description - `in.start_date` обязательно, должно соответствовать формату `ММ‑ГГГГ` и не быть `01‑1970`.
// @Description - `in.end_date`, если указано, должно соответствовать формату `ММ‑ГГГГ`, не быть `01‑1970` и не предшествовать `in.start_date`.
// @Description
// @Tags Подписки
// @Accept json
// @Param json body handlers.SubscriptionFullDTO true "Новые данные для подписки"
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

		tmpTime := time.Time(*changedDto.EndDate)
		subscription := models.Subscription{
			UUID:        changedDto.UUID,
			UserUUID:    changedDto.UserUUID,
			ServiceName: changedDto.ServiceName,
			Price:       changedDto.Price,
			StartDate:   time.Time(changedDto.StartDate),
			EndDate:     &tmpTime,
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

		tmpCustomTime := handlers.CustomTime(*updatedSubscription.EndDate)
		updatedDto := handlers.SubscriptionResponceDTO{
			UUID: updatedSubscription.UUID,
			SubscriptionDTO: handlers.SubscriptionDTO{
				ServiceName: updatedSubscription.ServiceName,
				UserUUID:    updatedSubscription.UserUUID,
				Price:       updatedSubscription.Price,
				StartDate:   handlers.CustomTime(updatedSubscription.StartDate),
				EndDate:     &tmpCustomTime,
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
