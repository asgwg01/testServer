package utils

import (
	"testSrv/internal/domain/models"
	"testSrv/internal/http/handlers"

	"github.com/google/uuid"
)

func SubscriptionsToDTO(subs []models.Subscription) []handlers.SubscriptionDTO {
	result := make([]handlers.SubscriptionDTO, len(subs))

	for i, s := range subs {
		result[i] = handlers.SubscriptionDTO{
			ServiceName: s.ServiceName,
			Price:       s.Price,
			UserUUID:    s.UserUUID,
			StartDate:   handlers.CstomTime{Time: s.StartDate},
			EndDate:     s.EndDate,
		}
	}

	return result
}

func GenerateUUID() string {
	return uuid.NewString()
}
