package logger

import (
	"log"
	"time"
)

func NewUserAuthService(action string, updatedBy int32, updatedOn time.Duration, apiResponseCode int) {
	log.Print("-------------")
	log.Print(action)
	log.Print(updatedBy)
	log.Print(updatedOn)
	log.Print(apiResponseCode)
	log.Print("-------------")
}
