package handlers

import (
	"errors"
	"net/url"
	"testSrv/internal/domain/filter"
	"time"
)

var (
	ErrValidateUserUUID    = errors.New("invalid user uuid")
	ErrValidateServiceName = errors.New("invalid service name")
	ErrValidatePrice       = errors.New("invalid price")
	ErrValidateStartDate   = errors.New("invalid start date")
	ErrValidateEndDate     = errors.New("invalid end date")
)

type ErrorDTO struct {
	Error string `json:"error"`
}

// Для коректного анмаршалинга “07-2025”
type CustomTime struct {
	Time time.Time
}

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	str := string(data[1 : len(data)-1])
	layout := "01-2006" // Из текста задания
	parsedTime, err := time.Parse(layout, str)
	if err != nil {
		return err
	}
	ct.Time = parsedTime
	return nil
}

// SubscriptionDTO DTO для описания подписки
type SubscriptionDTO struct {
	// ServiceName название сервиса подписки
	// minLength: 1
	// maxLength: 100
	// example: Яндекс плюс
	ServiceName string `json:"service_name"`
	// Price стоимость подписки в рублях в месяц, целое число
	// minimum: 1
	// example: 399
	Price int `json:"price"`
	// UserUUID UUID пользователя
	// format: uuid
	// example: 123e4567-e89b-12d3-a456-426614174000
	UserUUID string `json:"user_id"`
	// StartDate месяц старта подписки
	// format: date-time
	// example: 07-2025
	StartDate CustomTime `json:"start_date"`
	// EndDate месяц окончания подписки
	// format: date-time
	// nullable: true
	// example: null
	EndDate *CustomTime `json:"end_date,omitempty"`
}

type SubscriptionRequestDTO = SubscriptionDTO

// SubscriptionFullDTO DTO для описания подписки с UUID
type SubscriptionFullDTO struct {
	// UserUUID UUID подписки
	// format: uuid
	// example: 123e4567-e89b-12d3-a456-426614174000
	UUID string `json:"uuid"`
	SubscriptionDTO
}

type SubscriptionResponceDTO = SubscriptionFullDTO

// FilterDTO DTO для описания фильтрации подписок
type FilterDTO struct {
	// ServiceName название сервиса подписки
	// minLength: 1
	// maxLength: 100
	// example: Яндекс плюс
	ServiceName string `json:"service_name,omitempty"`
	// UserUUID UUID пользователя
	// format: uuid
	// example: 123e4567-e89b-12d3-a456-426614174000
	UserUUID string `json:"user_id,omitempty"`
	// From месяц старта подписки
	// format: date-time
	// example: 07-2025
	From *CustomTime `json:"from,omitempty"`
	// EndDate месяц окончания подписки
	// format: date-time
	// nullable: true
	// example: null
	To *CustomTime `json:"to,omitempty"`
}

// FilterDTO DTO для описания суммарной стоимости и количества отфильтрованных подписок
type CostDTO struct {
	// Cost суммарная стоимость всех отфильтрованных подписок в рублях в месяц, целое число
	// minimum: 0
	// example: 798
	Cost int `json:"total_cost"`
	// ServiceCount количество всех отфильтрованных подписок
	// minimum: 0
	// example: 4
	ServiceCount int `json:"service_count"`
	// Filter фильтр использованный для получения результата
	Filter FilterDTO `json:"filter"`
}

func ValidateDTO(dto SubscriptionRequestDTO) error {
	var err error = nil

	if dto.UserUUID == "" {
		err = errors.Join(err, ErrValidateUserUUID)
	}

	if dto.ServiceName == "" {
		err = errors.Join(err, ErrValidateServiceName)
	}

	if dto.Price < 0 {
		err = errors.Join(err, ErrValidatePrice)
	}

	if dto.StartDate.Time.IsZero() {
		err = errors.Join(err, ErrValidateStartDate)
	}

	if dto.EndDate != nil {
		if dto.EndDate.Time.IsZero() || dto.EndDate.Time.Before(dto.StartDate.Time) {
			err = errors.Join(err, ErrValidateEndDate)
		}
	}

	return err
}

func ParseQueryFilter(url *url.URL) (filter.QueryFilter, error) {
	var from time.Time
	if fromStr := url.Query().Get("from"); fromStr != "" {
		layout := "01-2006" // Из текста задания
		parsedTime, err := time.Parse(layout, fromStr)
		if err != nil {
			return filter.QueryFilter{}, err
		}
		from = parsedTime
	}

	var to time.Time
	if toStr := url.Query().Get("to"); toStr != "" {
		layout := "01-2006" // Из текста задания
		parsedTime, err := time.Parse(layout, toStr)
		if err != nil {
			return filter.QueryFilter{}, err
		}
		to = parsedTime
	}

	var userUuid string
	if userUuidStr := url.Query().Get("user_uuid"); userUuidStr != "" {
		userUuid = userUuidStr
	}

	var serviceName string
	if serviceNameStr := url.Query().Get("service_name"); serviceNameStr != "" {
		serviceName = serviceNameStr
	}

	filter := filter.QueryFilter{
		UserUUID:    userUuid,
		ServiceName: serviceName,
	}
	if !from.IsZero() {
		filter.From = &from
	}
	if !to.IsZero() {
		filter.To = &to
	}

	return filter, nil
}
