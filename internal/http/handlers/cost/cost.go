package cost

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"testSrv/internal/http/handlers"
	"testSrv/internal/storage"
)

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
			dto.Filter.From = &handlers.CustomTime{Time: *filter.From}
		}
		if filter.To != nil {
			dto.Filter.To = &handlers.CustomTime{Time: *filter.To}
		}

		w.WriteHeader(http.StatusFound)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(dto); err != nil {
			log.Error("Error Encode DTO", slog.String("err", err.Error()))
			return
		}
	}
}
