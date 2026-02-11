package list

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"testSrv/internal/domain/utils"
	"testSrv/internal/http/handlers"
	"testSrv/internal/storage"
)

// GetSubscriptionList godoc
// @Summary Получение информации о подписках с использованием фильтра
// @Description Возвращает полную информацию о нескольких подписках, удовлетворяющих параметрам фильтрации
// @Tags Подписки
// @Produce json
// @Param user_uuid query string false "uuid пользователя для которого будут фильтроваться подписки" example("123e4567-e89b-12d3-a456-426614174000", "10000000-0000-0000-0000-000000000000") default()
// @Param service_name query string false "Имя сервиса для которого будут фильтроваться подписки" example("Яндекс плюс") default()
// @Param from query string false "С какого месяца фильтруются подписки" format:"date-time" example("07-2025")
// @Param to query string false "До какого месяца фильтруются подписки" format:"date-time" example("07-2025")
// @Success 200 {array} handlers.SubscriptionResponceDTO "Все подписки удовлетворяющие фильтрам"
// @Failure 400 {object} handlers.ErrorDTO "Неверный запрос, неверные фильтры, ошибка в формате даты"
// @Failure 500 {object} handlers.ErrorDTO "Внутренняя ошибка работы сервера"
// @Router /subscriptions [get]
func NewHandler(log *slog.Logger, geter storage.ISubscriptionGeter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const logPrefix = "handlers.list.handler"
		log := log.With(
			slog.String("where", logPrefix),
		)

		log.Debug("Recive message", slog.String("method", r.Method), slog.String("url", r.URL.String()))

		filter, err := handlers.ParseQueryFilter(r.URL)
		if err != nil {
			log.Error("failed to parse data", slog.String("err", err.Error()))

			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "application/json")

			errDto := handlers.ErrorDTO{
				Error: fmt.Sprintf("failed to parse data %s", err.Error()),
			}
			if err := json.NewEncoder(w).Encode(errDto); err != nil {
				log.Error("Error Encode DTO", slog.String("err", err.Error()))
				return
			}
			return
		}

		log.Debug("Recive filters",
			slog.String("filter", fmt.Sprintf("%#v", filter)),
		)

		subscriptions, err := geter.GetSubscriptionsWithFilter(filter)
		if err != nil {
			log.Error("internal error", slog.String("err", err.Error()))

			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/json")

			errDto := handlers.ErrorDTO{
				Error: "error get subscriptions",
			}
			if err := json.NewEncoder(w).Encode(errDto); err != nil {
				log.Error("Error Encode DTO", slog.String("err", err.Error()))
				return
			}
			return
		}

		dtos := utils.SubscriptionsToDTO(subscriptions)

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(dtos); err != nil {
			log.Error("Error Encode DTO", slog.String("err", err.Error()))
			return
		}
	}
}
