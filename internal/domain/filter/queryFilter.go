package filter

import "time"

type QueryFilter struct {
	From        *time.Time // pointer, because may be nill
	To          *time.Time // pointer, because may be nill
	UserUUID    string
	ServiceName string
}
