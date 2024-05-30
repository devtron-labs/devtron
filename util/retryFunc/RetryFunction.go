/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
