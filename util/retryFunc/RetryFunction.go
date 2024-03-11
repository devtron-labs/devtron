package retryFunc

import (
	"fmt"
	"go.uber.org/zap"
	"time"
)

// Retry performs a function with retries, delay, and a max number of attempts.
func Retry(fn func() error, shouldRetry func(err error) bool, maxRetries int, delay time.Duration, logger *zap.SugaredLogger) error {
	var err error
	for i := 0; i < maxRetries; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !shouldRetry(err) {
			return err
		}
		logger.Infow(fmt.Sprintf("Attempt %d failed, retrying in %v", i+1, delay), "err", err)
		time.Sleep(delay)
	}
	return fmt.Errorf("after %d attempts, last error: %s", maxRetries, err)
}
