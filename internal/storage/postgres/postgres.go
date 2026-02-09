package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"testSrv/internal/config"
	"testSrv/internal/domain/models"
	"testSrv/internal/storage"

	_ "github.com/lib/pq"
)

type Storage struct {
	log *slog.Logger
	db  *sql.DB
}

func NewStorage(log *slog.Logger, cfg config.StorageConfig) (*Storage, error) {
	const logPrefix = "postgres.Storage.CreateSubscription"
	l := log.With(
		slog.String("where", logPrefix),
	)

	connectionStr := "postgres://" + cfg.StorageUser + ":" + cfg.StoragePassword +
		"@" + cfg.StorageHost + ":" + cfg.StoragePort +
		"/" + cfg.StorageDB +
		"?sslmode=disable" // TODO: убрать если добавится dev/prod

	l.Debug("Create new psql conn", slog.String("str", connectionStr))

	//CONNECTION_STR=postgres://postgres:0000@localhost:5432/postgres?sslmode=disable

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {

		return nil, fmt.Errorf("can't open storage %s", err)
	}

	l.Info("Postgres connected", slog.String("host", cfg.StorageHost), slog.String("port", cfg.StoragePort))

	return &Storage{log: log, db: db}, nil

}

func (s *Storage) CreateSubscription(subscription models.Subscription) (models.Subscription, error) {
	return models.Subscription{}, nil
}

func (s *Storage) GetSubscription(uuid string) (models.Subscription, error) {

	const logPrefix = "postgres.Storage.GetSubscription"
	log := s.log.With(
		slog.String("where", logPrefix),
	)

	log.Info("Get subscription", slog.String("Subscription uuid", uuid))

	query, err := s.db.Prepare(`SELECT 
	subs.uuid,
	srv.name AS service_name,
	srv.price, 
	usr.uuid,
	subs.start_date,
	subs.end_date
	FROM subscriptions AS subs
	INNER JOIN services AS srv
	ON subs.service_id=srv.uuid
	INNER JOIN users AS usr
	ON subs.user_id=usr.uuid
	WHERE subs.uuid=$1`)

	if err != nil {
		return models.Subscription{}, fmt.Errorf("SELECT error %s", err)
	}

	defer func() { query.Close() }()

	result := models.Subscription{}

	err = query.QueryRow(uuid).Scan(
		&result.UUID,
		&result.ServiceName,
		&result.Price,
		&result.UserUUID,
		&result.StartDate,
		&result.EndDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Info("SELECT Query error", slog.String("err", storage.ErrorSubscriptionNotFound.Error()))
			return models.Subscription{}, fmt.Errorf("SELECT Query error %s", storage.ErrorSubscriptionNotFound)
		} else {
			log.Error("SELECT Query error", slog.String("err", err.Error()))
			return models.Subscription{}, err
		}
	}

	return result, nil
}

func (s *Storage) GetSubscriptionsWithFilter(filter storage.QueryFilter) ([]models.Subscription, error) {
	return []models.Subscription{}, nil
}

func (s *Storage) UpdateSubscription(subscription models.Subscription) (models.Subscription, error) {
	return models.Subscription{}, nil
}

func (s *Storage) DeleteSubscription(uuid string) error {
	return nil
}
