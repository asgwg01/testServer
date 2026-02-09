package handlers

import (
	"errors"
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
type CstomTime struct {
	Time time.Time
}

func (ct *CstomTime) UnmarshalJSON(data []byte) error {
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
	ServiceName string     `json:"service_name"`
	Price       int        `json:"price"`
	UserUUID    string     `json:"user_id"`
	StartDate   CstomTime  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
}

type SubscriptionRequestDTO = SubscriptionDTO

type SubscriptionFullDTO struct {
	UUID string `json:"uuid"`
	SubscriptionDTO
}

type SubscriptionResponceDTO = SubscriptionFullDTO

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
		if dto.EndDate.IsZero() || dto.EndDate.Before(dto.StartDate.Time) {
			err = errors.Join(err, ErrValidateEndDate)
		}
	}

	return err
}
