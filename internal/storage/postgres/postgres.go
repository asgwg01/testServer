package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"testSrv/internal/config"
	"testSrv/internal/domain/filter"
	"testSrv/internal/domain/models"
	"testSrv/internal/domain/utils"
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

	db, err := sql.Open("postgres", connectionStr)
	if err != nil {

		return nil, fmt.Errorf("can't open storage %s", err)
	}

	l.Info("Postgres connected", slog.String("host", cfg.StorageHost), slog.String("port", cfg.StoragePort))

	return &Storage{log: log, db: db}, nil

}

func (s *Storage) CreateSubscription(subscription models.Subscription) (models.Subscription, error) {
	const logPrefix = "postgres.Storage.CreateSubscription"
	log := s.log.With(
		slog.String("where", logPrefix),
	)
	log.Info("Creare subscription", utils.SubscriptionToSlog(subscription))

	tx, err := s.db.Begin()
	if err != nil {
		log.Error("Can not create transaction", slog.String("err", err.Error()))
		return models.Subscription{}, err
	} else { // transaction

		serviceUuid, err := s.checkUserAndServiceExist(tx, subscription)
		if err != nil {
			// rolback in checkUserAndServiceExist
			return models.Subscription{}, err
		}

		// Check current active subscriptions
		query, err := tx.Prepare(`
		SELECT uuid
		FROM subscriptions
		WHERE user_id=$1
		AND service_id=$2
		AND (
			(start_date <= $3 and $3 <= end_date) 
			OR (start_date <= $4 and $4 <= end_date)
		);`)
		if err != nil {
			tx.Rollback()
			return models.Subscription{}, fmt.Errorf("SELECT error %s", err)
		}

		var existSubscriptionUuid string
		err = query.QueryRow(subscription.UserUUID, serviceUuid, subscription.StartDate, subscription.EndDate).Scan(&existSubscriptionUuid)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// all Ok
			} else {
				tx.Rollback()
				log.Error("Error check subscription exists", slog.String("name", subscription.ServiceName))
				log.Error(err.Error())
				return models.Subscription{}, err
			}
		} else {
			tx.Rollback()
			log.Error("Error subscription already exist",
				slog.String("user uuid", subscription.UserUUID),
				slog.String("service name", subscription.ServiceName),
			)
			return models.Subscription{
				UUID: existSubscriptionUuid,
			}, storage.ErrorSubscriptionAlreadyExist
		}

		// Add Subs
		query, err = tx.Prepare(`
		INSERT INTO
		subscriptions(
			service_id,
			user_id,
			price,
			start_date,
			end_date
		)
		VALUES
		($1, $2, $3, $4, $5)
		RETURNING uuid;
		`)
		if err != nil {
			tx.Rollback()
			return models.Subscription{}, fmt.Errorf("INSERT error %s", err)
		}

		var subscriptionUuid string
		err = query.QueryRow(serviceUuid, subscription.UserUUID, subscription.Price, subscription.StartDate, subscription.EndDate).Scan(&subscriptionUuid)
		if err != nil {
			tx.Rollback()
			log.Error("Error inserting subscription", utils.SubscriptionToSlog(subscription))
			return models.Subscription{}, fmt.Errorf("INSERT error %s", err)
		}
		if subscriptionUuid == "" {
			tx.Rollback()
			log.Error("Error get last insert uuid", utils.SubscriptionToSlog(subscription))
			return models.Subscription{}, fmt.Errorf("INSERT error %s", err)
		}

		// Get new subscription uuid

		err = tx.Commit()
		if err != nil {
			log.Error("error commit transaction", slog.String("err", err.Error()))
			return models.Subscription{}, fmt.Errorf("error commit transaction %s", err)
		}

		log.Info("Creare subscription success!", utils.SubscriptionToSlog(subscription))

		subscription.UUID = subscriptionUuid
	} // end transaction

	return subscription, nil
}

func (s *Storage) GetSubscription(uuid string) (models.Subscription, error) {
	const logPrefix = "postgres.Storage.GetSubscription"
	log := s.log.With(
		slog.String("where", logPrefix),
	)

	log.Info("Get subscription", slog.String("Subscription uuid", uuid))

	query, err := s.db.Prepare(`
	SELECT 
	subs.uuid,
	srv.name AS service_name,
	subs.price, 
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
			return models.Subscription{}, storage.ErrorSubscriptionNotFound
		} else {
			log.Error("SELECT Query error", slog.String("err", err.Error()))
			return models.Subscription{}, err
		}
	}

	return result, nil
}

func (s *Storage) UpdateSubscription(subscription models.Subscription) (models.Subscription, error) {
	const logPrefix = "postgres.Storage.UpdateSubscription"
	log := s.log.With(
		slog.String("where", logPrefix),
	)
	log.Info("Update subscription", utils.SubscriptionToSlog(subscription))

	tx, err := s.db.Begin()
	if err != nil {
		log.Error("Can not create transaction", slog.String("err", err.Error()))
		return models.Subscription{}, err
	} else { // transaction

		// Check subs exist
		query, err := tx.Prepare(`
		SELECT EXISTS ( 
			SELECT 1
			FROM subscriptions
			WHERE uuid=$1
		);`)
		if err != nil {
			tx.Rollback()
			log.Error("Error check subscription exists", slog.String("uuid", subscription.UUID))
			return models.Subscription{}, fmt.Errorf("SELECT error %s", err)
		}

		var exists bool
		err = query.QueryRow(subscription.UUID).Scan(&exists)
		if err != nil {
			tx.Rollback()
			log.Error("Error subscription check exists", slog.String("uuid", subscription.UUID))
			return models.Subscription{}, err
		}
		if !exists {
			tx.Rollback()
			log.Info("subscription is not exist", slog.String("uuid", subscription.UUID))
			return models.Subscription{}, storage.ErrorSubscriptionNotFound
		}

		// check user and service
		serviceUuid, err := s.checkUserAndServiceExist(tx, subscription)
		if err != nil {
			// rolback in checkUserAndServiceExist
			return models.Subscription{}, err
		}

		// Update Subs
		query, err = tx.Prepare(`
		UPDATE subscriptions
		SET 
		service_id = $2,
    	user_id = $3,
		price = $4,
		start_date = $5,
		end_date = $6
		WHERE uuid = $1;
		`)
		if err != nil {
			tx.Rollback()
			return models.Subscription{}, fmt.Errorf("UPDATE error %s", err)
		}

		_, err = query.Exec(
			subscription.UUID,
			serviceUuid,
			subscription.UserUUID,
			subscription.Price,
			subscription.StartDate,
			subscription.EndDate,
		)
		if err != nil {
			tx.Rollback()
			log.Error("Error update subscription", utils.SubscriptionToSlog(subscription))
			return models.Subscription{}, fmt.Errorf("UPDATE error %s", err)
		}

		err = tx.Commit()
		if err != nil {
			log.Error("error commit transaction", slog.String("err", err.Error()))
			return models.Subscription{}, fmt.Errorf("error commit transaction %s", err)
		}

		log.Info("Update subscription success!", utils.SubscriptionToSlog(subscription))

	} // end transaction

	return subscription, nil
}

func (s *Storage) DeleteSubscription(uuid string) error {
	const logPrefix = "postgres.Storage.DeleteSubscription"
	log := s.log.With(
		slog.String("where", logPrefix),
	)
	log.Info("Delete subscription", slog.String("uuid", uuid))

	tx, err := s.db.Begin()
	if err != nil {
		log.Error("Can not create transaction", slog.String("err", err.Error()))
		return err
	} else { // transaction

		// check exist
		query, err := tx.Prepare(`
		SELECT EXISTS ( 
			SELECT 1
			FROM subscriptions
			WHERE uuid=$1
		);`)
		if err != nil {
			tx.Rollback()
			log.Error("Error check subscription exists", slog.String("uuid", uuid))
			return storage.ErrorSubscriptionNotFound
		}

		var exists bool
		err = query.QueryRow(uuid).Scan(&exists)
		if err != nil {
			tx.Rollback()
			log.Error("Error subscription check exists", slog.String("uuid", uuid))
			return err
		}
		if !exists {
			tx.Rollback()
			log.Info("subscription is not exist", slog.String("uuid", uuid))
			return storage.ErrorSubscriptionNotFound
		}

		// delete
		query, err = tx.Prepare(`
		DELETE FROM subscriptions
		WHERE uuid=$1;
		`)
		if err != nil {
			tx.Rollback()
			log.Error("Error delete subscription", slog.String("uuid", uuid))
			return fmt.Errorf("DELETE error %s", err)
		}

		_, err = query.Exec(uuid)
		if err != nil {
			tx.Rollback()
			log.Error("Error delete subscription", slog.String("uuid", uuid))
			return err
		}

		err = tx.Commit()
		if err != nil {
			log.Error("error commit transaction", slog.String("err", err.Error()))
			return fmt.Errorf("error commit transaction %s", err)
		}

		log.Info("Delete subscription success!", slog.String("uuid", uuid))
	} // transaction end

	return nil
}

func (s *Storage) GetSubscriptionsWithFilter(filter filter.QueryFilter) ([]models.Subscription, error) {
	const logPrefix = "postgres.Storage.GetSubscriptionsWithFilter"
	log := s.log.With(
		slog.String("where", logPrefix),
	)
	log.Info("Get subscriptions filtered", utils.FilterToSlog(filter))

	strQuery := `
	SELECT 
	sub.uuid,
	srv.name AS sevice_name,
	sub.user_id,
	sub.price,
	sub.start_date,
	sub.end_date
	FROM
	subscriptions AS sub
	INNER JOIN services AS srv
		ON sub.service_id=srv.uuid
	WHERE
	`
	isEmptyFilter := true
	setParams := []any{}

	if filter.UserUUID != "" {
		if isEmptyFilter {
			isEmptyFilter = false
		} else {
			strQuery += " AND\n"
		}
		setParams = append(setParams, &filter.UserUUID)
		strQuery += fmt.Sprintf("sub.user_id=$%d", len(setParams))
	}

	if filter.ServiceName != "" {
		if isEmptyFilter {
			isEmptyFilter = false
		} else {
			strQuery += " AND\n"
		}
		setParams = append(setParams, &filter.ServiceName)
		strQuery += fmt.Sprintf("srv.name=$%d", len(setParams))
	}

	if filter.From != nil {
		if isEmptyFilter {
			isEmptyFilter = false
		} else {
			strQuery += " AND\n"
		}
		setParams = append(setParams, &filter.From)
		strQuery += fmt.Sprintf("sub.start_date >= $%d", len(setParams))
	}

	if filter.To != nil {
		if isEmptyFilter {
			isEmptyFilter = false
		} else {
			strQuery += " AND\n"
		}
		setParams = append(setParams, &filter.To)
		strQuery += fmt.Sprintf("sub.end_date <= $%d", len(setParams))
	}

	if isEmptyFilter {
		strQuery += "TRUE"
	}
	strQuery += ";"

	log.Debug("Buildet query", utils.FilterToSlog(filter), slog.String("query", strQuery))

	query, err := s.db.Prepare(strQuery)
	if err != nil {
		return []models.Subscription{}, fmt.Errorf("SELECT error %s", err)
	}

	rows, err := query.Query(setParams...)
	if err != nil {
		log.Error("Error select with filter", slog.String("err", err.Error()))
		return []models.Subscription{}, err
	}
	defer rows.Close()

	result := []models.Subscription{}
	for rows.Next() {
		subs := models.Subscription{}

		err := rows.Scan(
			&subs.UUID,
			&subs.ServiceName,
			&subs.UserUUID,
			&subs.Price,
			&subs.StartDate,
			&subs.EndDate,
		)
		if err != nil {
			log.Error("Error parse subscription")
			return []models.Subscription{}, err
		}
		result = append(result, subs)
	}

	return result, nil
}

// checkUserAndServiceExist return service uuid if is exist
func (s *Storage) checkUserAndServiceExist(tx *sql.Tx, subscription models.Subscription) (string, error) {
	const logPrefix = "postgres.Storage.checkUserAndServiceExist"
	log := s.log.With(
		slog.String("where", logPrefix),
	)
	log.Info("check exist user and service", utils.SubscriptionToSlog(subscription))

	// Check User
	query, err := tx.Prepare(`
		SELECT EXISTS ( 
			SELECT 1
			FROM users
			WHERE uuid=$1
		);`)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("SELECT error %s", err)
	}

	var exists bool
	err = query.QueryRow(subscription.UserUUID).Scan(&exists)
	if err != nil {
		tx.Rollback()
		log.Error("Error check user exists", slog.String("uuid", subscription.UserUUID))
		return "", err
	}
	if !exists {
		tx.Rollback()
		log.Info("User is not exist", slog.String("uuid", subscription.UserUUID))
		return "", storage.ErrorUserNotExist
	}

	// Check Service
	query, err = tx.Prepare(`
		SELECT uuid
		FROM services
		WHERE name=$1;`)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("SELECT error %s", err)
	}

	var serviceUuid string
	err = query.QueryRow(subscription.ServiceName).Scan(&serviceUuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			tx.Rollback()
			log.Info("Service is not exist", slog.String("name", subscription.ServiceName))
			return "", storage.ErrorServiceNotExist
		} else {
			tx.Rollback()
			log.Error("Error check service exists", slog.String("name", subscription.ServiceName))
			return "", err
		}
	}

	return serviceUuid, nil
}
