package logger

import (
	"log"
	"time"
)

func NewUserAuthService(action string, updatedBy int32, updatedOn time.Duration, apiResponseCode int, payload string) {
	log.Printf("Action: %s, UpdatedBy: %d, UpdatedOn: %s, APIResponseCode: %d, payload %s", action, updatedBy, updatedOn, apiResponseCode, payload)
}
