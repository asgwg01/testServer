package postgres

import (
	"database/sql"
	"log/slog"
	"testSrv/internal/config"
	"testSrv/internal/domain/models"
	"testSrv/internal/storage"
)

type Storage struct {
	log *slog.Logger
	db  *sql.DB
}

func NewStorage(log *slog.Logger, cfg config.StorageConfig) (*Storage, error) {
	//CONNECTION_STR=postgres://postgres:0000@localhost:5432/postgres?sslmode=disable
	return &Storage{db: nil}, nil

}

func (s *Storage) CreateSubscription(subscription models.Subscription) (models.Subscription, error) {
	return models.Subscription{}, nil
}

func (s *Storage) GetSubscription(uuid string) (models.Subscription, error) {
	return models.Subscription{}, nil
}

func (s *Storage) GetSubscriptionsWithFilter(filter storage.QueryFilter) ([]models.Subscription, error) {
	return []models.Subscription{}, nil
}

func (s *Storage) UpdateSubscriptions(subscription models.Subscription) (models.Subscription, error) {
	return models.Subscription{}, nil
}

func (s *Storage) DeleteSubscription(uuid string) error {
	return nil
}
