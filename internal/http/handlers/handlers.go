package handlers

import "time"

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
