/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package util

import (
	"fmt"
	"go.uber.org/zap"
	"math/rand"
	"strconv"
	"strings"
)

func ContainsString(list []string, element string) bool {
	if len(list) == 0 {
		return false
	}
	for _, l := range list {
		if l == element {
			return true
		}
	}
	return false
}

func AppendErrorString(errs []string, err error) []string {
	if err != nil {
		errs = append(errs, err.Error())
	}
	return errs
}

func GetErrorOrNil(errs []string) error {
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func ExtractChartVersion(chartVersion string) (int, int, error) {
	if len(chartVersion) == 0 {
		return 0, 0, nil
	}
	chartVersions := strings.Split(chartVersion, ".")
	chartMajorVersion, err := strconv.Atoi(chartVersions[0])
	if err != nil {
		return 0, 0, err
	}
	chartMinorVersion, err := strconv.Atoi(chartVersions[1])
	if err != nil {
		return 0, 0, err
	}
	return chartMajorVersion, chartMinorVersion, nil
}

type Closer interface {
	Close() error
}

func Close(c Closer, logger *zap.SugaredLogger) {
	if err := c.Close(); err != nil {
		logger.Warnf("failed to close %v: %v", c, err)
	}
}

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func Generate(size int) string {
	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}
