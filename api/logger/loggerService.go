package logger

import (
	"log"
	"time"
)

func NewUserAuthService(action string, updatedBy int32, updatedOn time.Duration, apiResponseCode int) {
	log.Println(action)
	log.Println(updatedBy)
	log.Println(updatedOn)
	log.Println(apiResponseCode)
}
