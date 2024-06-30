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

package cron

import (
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/common-lib/pubsub-lib/metrics"
	"go.uber.org/zap"
)

const PANIC = "panic"

type CronLoggerImpl struct {
	logger *zap.SugaredLogger
}

func (impl *CronLoggerImpl) Info(msg string, keysAndValues ...interface{}) {
	impl.logger.Infow(msg, keysAndValues...)
}

func (impl *CronLoggerImpl) Error(err error, msg string, keysAndValues ...interface{}) {
	if msg == PANIC {
		metrics.IncPanicRecoveryCount("cron", "", "", "")
	}
	keysAndValues = append([]interface{}{"err", err}, keysAndValues...)
	impl.logger.Errorw(constants.PanicLogIdentifier+": "+msg, keysAndValues...)
}

func NewCronLoggerImpl(logger *zap.SugaredLogger) *CronLoggerImpl {
	return &CronLoggerImpl{logger: logger}
}
