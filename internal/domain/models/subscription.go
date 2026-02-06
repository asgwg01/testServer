package models

import "time"

type Subscription struct {
	UUID        string
	ServiceName string
	Price       int
	UserUUID    string
	StartDate   time.Time
	EndDate     *time.Time // pointer, EndDate may be nil
}
