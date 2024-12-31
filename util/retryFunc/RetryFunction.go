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
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/runTime"
	"go.uber.org/zap"
	"time"
)

// Retry performs a function with retries, delay, and a max number of attempts.
func Retry(fn func() error, shouldRetry func(err error) bool, maxRetries int, delay time.Duration, logger *zap.SugaredLogger) error {
	var err error
	logger.Debugw("retrying function",
		"maxRetries", maxRetries, "delay", delay,
		"callerFunc", runTime.GetCallerFunctionName(),
		"path", fmt.Sprintf("%s:%d", runTime.GetCallerFileName(), runTime.GetCallerLineNumber()))
	for i := 0; i < maxRetries; i++ {
		logger.Debugw("function called with retry", "attempt", i+1, "maxRetries", maxRetries, "delay", delay)
		err = fn()
		if err == nil {
			return nil
		}
		if !shouldRetry(err) {
			return sanitiseError(err)
		}
		logger.Infow(fmt.Sprintf("Attempt %d failed, retrying in %v", i+1, delay), "err", err)
		time.Sleep(delay)
	}
	return fmt.Errorf("after %d attempts, last error: %s", maxRetries, err)
}

// RetryableError is the error returned by callback function under Retry.
// for RetryableError errors can be handled by shouldRetry func
type RetryableError struct {
	err error
}

func NewRetryableError(err error) *RetryableError {
	return &RetryableError{
		err: err,
	}
}

func (r *RetryableError) Error() string { return r.err.Error() }

func (r *RetryableError) GetError() error { return r.err }

func sanitiseError(err error) error {
	if retryErr := (&RetryableError{}); errors.As(err, &retryErr) {
		return retryErr.GetError()
	}
	return err
}
