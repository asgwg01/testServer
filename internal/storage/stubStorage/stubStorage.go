package stubstorage

import (
	"fmt"
	"log/slog"
	"testSrv/internal/domain/filter"
	"testSrv/internal/domain/models"
)

type Storage struct {
	log *slog.Logger
}

func NewStorage(log *slog.Logger) (*Storage, error) {
	l := log.With(
		slog.String("STUB!", "stubStorage"),
	)
	return &Storage{log: l}, nil
}

func (s *Storage) CreateSubscription(subscription models.Subscription) (models.Subscription, error) {
	const logPrefix = "stubstorage.Storage.CreateSubscription"
	log := s.log.With(
		slog.String("where", logPrefix),
	)

	log.Info("Call STUB", slog.String("Subscription uuid", subscription.UUID))
	return subscription, nil
}

func (s *Storage) GetSubscription(uuid string) (models.Subscription, error) {
	const logPrefix = "stubstorage.Storage.GetSubscription"
	log := s.log.With(
		slog.String("where", logPrefix),
	)

	log.Info("Call STUB", slog.String("Subscription uuid", uuid))
	return models.Subscription{}, nil

	// endDate := time.Now().Add(time.Hour * 24 * 30)
	// return models.Subscription{
	// 	UUID:        "1",
	// 	ServiceName: "Yandex plus",
	// 	Price:       399,
	// 	UserUUID:    "1",	
	// 	StartDate:   time.Now(),
	// 	EndDate:     &endDate,
	// }, nil
}

func (s *Storage) GetSubscriptionsWithFilter(filter filter.QueryFilter) ([]models.Subscription, error) {
	const logPrefix = "stubstorage.Storage.GetSubscriptionsWithFilter"
	log := s.log.With(
		slog.String("where", logPrefix),
	)

	log.Info("Call STUB", slog.String("Filter", fmt.Sprintf("%T", filter)))
	return []models.Subscription{}, nil

	// endDate := time.Now().Add(time.Hour * 24 * 30)
	// return []models.Subscription{
	// 	{
	// 		UUID:        "1",
	// 		ServiceName: "Yandex plus",
	// 		Price:       399,
	// 		UserUUID:    "1",
	// 		StartDate:   time.Now(),
	// 		EndDate:     &endDate,
	// 	},
	// 	{
	// 		UUID:        "2",
	// 		ServiceName: "Netflix",
	// 		Price:       999,
	// 		UserUUID:    "1",
	// 		StartDate:   time.Now(),
	// 		EndDate:     &endDate,
	// 	},
	// 	{
	// 		UUID:        "3",
	// 		ServiceName: "PlayStation Plus",
	// 		Price:       9999,
	// 		UserUUID:    "2",
	// 		StartDate:   time.Now(),
	// 		EndDate:     &endDate,
	// 	},
	// }, nil
}

func (s *Storage) UpdateSubscription(subscription models.Subscription) (models.Subscription, error) {
	const logPrefix = "stubstorage.Storage.UpdateSubscriptions"
	log := s.log.With(
		slog.String("where", logPrefix),
	)

	log.Info("Call STUB", slog.String("Subscription uuid", subscription.UUID))
	return subscription, nil
}

func (s *Storage) DeleteSubscription(uuid string) error {
	const logPrefix = "stubstorage.Storage.DeleteSubscription"
	log := s.log.With(
		slog.String("where", logPrefix),
	)

	log.Info("Call STUB", slog.String("Subscription uuid", uuid))
	return nil
}
