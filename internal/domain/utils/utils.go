package utils

import (
	"log/slog"
	"testSrv/internal/domain/filter"
	"testSrv/internal/domain/models"
	"testSrv/internal/http/handlers"
)

func SubscriptionsToDTO(subs []models.Subscription) []handlers.SubscriptionFullDTO {
	result := make([]handlers.SubscriptionFullDTO, len(subs))

	for i, s := range subs {
		tmpCustomTime := handlers.CustomTime(*s.EndDate)
		result[i] = handlers.SubscriptionFullDTO{
			UUID: s.UUID,
			SubscriptionDTO: handlers.SubscriptionDTO{
				ServiceName: s.ServiceName,
				Price:       s.Price,
				UserUUID:    s.UserUUID,
				StartDate:   handlers.CustomTime(s.StartDate),
				EndDate:     &tmpCustomTime,
			},
		}
	}

	return result
}

func SubscriptionToSlog(sub models.Subscription) slog.Attr {
	result := slog.Group(
		"subscription",
		slog.String("uuid", sub.UUID),
		slog.String("service_name", sub.ServiceName),
		slog.Int("price", sub.Price),
		slog.String("user_id", sub.UserUUID),
		slog.Time("start_date", sub.StartDate),
		slog.Time("end_date", *sub.EndDate),
	)

	return result
}

func FilterToSlog(filter filter.QueryFilter) slog.Attr {
	reqAttrs := []slog.Attr{
		slog.String("user_id", filter.UserUUID),
		slog.String("service_name", filter.ServiceName),
	}

	if filter.From != nil {
		reqAttrs = append(reqAttrs, slog.Time("from", *filter.From))
	}

	if filter.To != nil {
		reqAttrs = append(reqAttrs, slog.Time("to", *filter.To))
	}

	result := slog.GroupAttrs("filter", reqAttrs...)

	return result
}
