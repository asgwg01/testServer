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

type SubscriptionDTO struct {
	ServiceName string      `json:"service_name"`
	Price       int         `json:"price"`
	UserUUID    string      `json:"user_id"`
	StartDate   CustomTime  `json:"start_date"`
	EndDate     *CustomTime `json:"end_date,omitempty"`
}

type SubscriptionRequestDTO = SubscriptionDTO

type SubscriptionFullDTO struct {
	UUID string `json:"uuid"`
	SubscriptionDTO
}

type SubscriptionResponceDTO = SubscriptionFullDTO

type FilterDTO struct {
	ServiceName string      `json:"service_name,omitempty"`
	UserUUID    string      `json:"user_id,omitempty"`
	From        *CustomTime `json:"from,omitempty"`
	To          *CustomTime `json:"to,omitempty"`
}

type CostDTO struct {
	Cost         int       `json:"total_cost"`
	ServiceCount int       `json:"service_count"`
	Filter       FilterDTO `json:"filter"`
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
