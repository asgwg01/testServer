package storage

import (
	"errors"
	"testSrv/internal/domain/filter"
	"testSrv/internal/domain/models"
)

var (
	ErrorSubscriptionNotFound     = errors.New("sbscription not found")
	ErrorSubscriptionAlreadyExist = errors.New("sbscription already exist")
	ErrorUserNotExist             = errors.New("user is not exist")
	ErrorServiceNotExist          = errors.New("service is not exist")
)

type ISubscriptionStorage interface {
	CreateSubscription(subscription models.Subscription) (models.Subscription, error)
	GetSubscription(uuid string) (models.Subscription, error)
	GetSubscriptionsWithFilter(filter filter.QueryFilter) ([]models.Subscription, error)
	UpdateSubscription(subscription models.Subscription) (models.Subscription, error)
	DeleteSubscription(uuid string) error
}

type ISubscriptionCreator interface {
	CreateSubscription(subscription models.Subscription) (models.Subscription, error)
}

type ISubscriptionGeter interface {
	GetSubscription(uuid string) (models.Subscription, error)
	GetSubscriptionsWithFilter(filter filter.QueryFilter) ([]models.Subscription, error)
}

type ISubscriptionUpdater interface {
	UpdateSubscription(subscription models.Subscription) (models.Subscription, error)
}

type ISubscriptionDeleter interface {
	DeleteSubscription(uuid string) error
}
