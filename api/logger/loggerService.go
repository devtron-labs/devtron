package logger

import (
	"log"
	"time"
)

func NewUserAuthService(action string, updatedBy int32, updatedOn time.Duration, apiResponseCode int) {
	log.Printf("Action: %s, UpdatedBy: %d, UpdatedOn: %s, APIResponseCode: %d", action, updatedBy, updatedOn, apiResponseCode)
}
