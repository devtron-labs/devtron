package logger

import "time"

type UserAuthDTO struct {
	action          string    `json:"action"`
	UserID          int       `json:"userID"`
	updatedBy       int       `json:"updatedBy"`
	updatedOn       time.Time `json:"updatedOn"`
	apiResponseCode int       `json:"apiResponseCode"`
}
