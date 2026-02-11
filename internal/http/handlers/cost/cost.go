package cost

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"testSrv/internal/http/handlers"
	"testSrv/internal/storage"
)

// GetSubscriptionsCost godoc
// @Summary Получение информации о стоимости подписок с использованием фильтра
// @Description Возвращает информацию о суммарной стоимости и количестве подписок удовлетворяющих фильтрам.
// @Description
// @Description Также возвращает указанные фильтры
// @Tags Стоимость
// @Produce json
// @Param user_uuid query string false "uuid пользователя для которого будут фильтроваться подписки" example("123e4567-e89b-12d3-a456-426614174000", "10000000-0000-0000-0000-000000000000") default()
// @Param service_name query string false "Имя сервиса для которого будут фильтроваться подписки" example("Яндекс плюс") default()
// @Param from query string false "С какого месяца фильтруются подписки" format:"date-time" example("07-2025")
// @Param to query string false "До какого месяца фильтруются подписки" format:"date-time" example("07-2025")
// @Success 200 {object} handlers.CostDTO "Все подписки удовлетворяющие фильтрам"
// @Failure 400 {object} handlers.ErrorDTO "Неверный запрос, неверные фильтры, ошибка в формате даты"
// @Failure 500 {object} handlers.ErrorDTO "Внутренняя ошибка работы сервера"
// @Router /cost [get]
func NewHandler(log *slog.Logger, geter storage.ISubscriptionGeter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const logPrefix = "handlers.cost.handler"
		log := log.With(
			slog.String("where", logPrefix),
		)

		log.Debug("Recive message", slog.String("method", r.Method), slog.String("url", r.URL.String()))

		filter, err := handlers.ParseQueryFilter(r.URL)
		if err != nil {
			log.Error("failed to parse data", slog.String("err", err.Error()))

			w.WriteHeader(http.StatusInternalServerError)
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
		//cost
		dto := handlers.CostDTO{}

		for _, subs := range subscriptions {
			dto.Cost += subs.Price
			dto.ServiceCount++
		}
		dto.Filter = handlers.FilterDTO{
			ServiceName: filter.ServiceName,
			UserUUID:    filter.UserUUID,
		}
		if filter.From != nil {
			tmpCustomTime := handlers.CustomTime(*filter.From)
			dto.Filter.From = &tmpCustomTime
		}
		if filter.To != nil {
			tmpCustomTime := handlers.CustomTime(*filter.To)
			dto.Filter.To = &tmpCustomTime
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(dto); err != nil {
			log.Error("Error Encode DTO", slog.String("err", err.Error()))
			return
		}
	}
}
