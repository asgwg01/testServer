package storage

import (
	"errors"
	"testSrv/internal/domain/models"
	"time"
)

var (
	ErrorSubscriptionNotFound     = errors.New("sbscription not found")
	ErrorSubscriptionAlreadyExist = errors.New("sbscription already exist")
)

type QueryFilter struct {
	From             *time.Time // pointer, because may be nill
	To               *time.Time // pointer, because may be nill
	UserUUID         string
	SubscriptionName string
}

type ISubscriptionStorage interface {
	CreateSubscription(subscription models.Subscription) (models.Subscription, error)
	GetSubscription(uuid string) (models.Subscription, error)
	GetSubscriptionsWithFilter(filter QueryFilter) ([]models.Subscription, error)
	UpdateSubscriptions(subscription models.Subscription) (models.Subscription, error)
	DeleteSubscription(uuid string) error
}

type ISubscriptionCreator interface {
	CreateSubscription(subscription models.Subscription) (models.Subscription, error)
}

type ISubscriptionGeter interface {
	GetSubscription(uuid string) (models.Subscription, error)
	GetSubscriptionsWithFilter(filter QueryFilter) ([]models.Subscription, error)
}

type ISubscriptionUpdater interface {
	UpdateSubscriptions(subscription models.Subscription) (models.Subscription, error)
}

type ISubscriptionDeleter interface {
	DeleteSubscription(uuid string) error
}
