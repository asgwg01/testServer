package list

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"testSrv/internal/domain/utils"
	"testSrv/internal/http/handlers"
	"testSrv/internal/storage"
	"time"
)

func NewHandler(log *slog.Logger, geter storage.ISubscriptionGeter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const logPrefix = "handlers.list.handler"
		log := log.With(
			slog.String("where", logPrefix),
		)

		log.Debug("Recive message", slog.String("method", r.Method), slog.String("url", r.URL.String()))

		var from time.Time
		if fromStr := r.URL.Query().Get("from"); fromStr != "" {
			layout := "01-2006" // Из текста задания
			parsedTime, err := time.Parse(layout, fromStr)
			if err != nil {
				log.Error("failed to parse data", slog.String("from", fromStr), slog.String("err", err.Error()))

				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: fmt.Sprintf("failed to parse data %s: %s", "from", fromStr),
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}

				return
			}
			from = parsedTime
		}

		var to time.Time
		if toStr := r.URL.Query().Get("to"); toStr != "" {
			layout := "01-2006" // Из текста задания
			parsedTime, err := time.Parse(layout, toStr)
			if err != nil {
				log.Error("failed to parse data", slog.String("to", toStr))

				w.WriteHeader(http.StatusInternalServerError)
				w.Header().Set("Content-Type", "application/json")

				errDto := handlers.ErrorDTO{
					Error: fmt.Sprintf("failed to parse data %s: %s", "to", toStr),
				}
				if err := json.NewEncoder(w).Encode(errDto); err != nil {
					log.Error("Error Encode DTO", slog.String("err", err.Error()))
					return
				}

				return
			}
			to = parsedTime
		}

		var userUuid string
		if userUuidStr := r.URL.Query().Get("user_uuid"); userUuidStr != "" {
			userUuid = userUuidStr
		}

		var subscriptionName string
		if subscriptionNameStr := r.URL.Query().Get("subscription_name"); subscriptionNameStr != "" {
			subscriptionName = subscriptionNameStr
		}

		filter := storage.QueryFilter{
			UserUUID:         userUuid,
			SubscriptionName: subscriptionName,
		}
		if !from.IsZero() {
			filter.From = &from
		}
		if !to.IsZero() {
			filter.To = &to
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

		w.WriteHeader(http.StatusFound)
		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(dtos); err != nil {
			log.Error("Error Encode DTO", slog.String("err", err.Error()))
			return
		}
	}
}
