package logger

import "time"

type UserAuthDTO struct {
	action          string
	UserID          int
	updatedBy       int
	updatedOn       time.Time
	apiResponseCode int
}
