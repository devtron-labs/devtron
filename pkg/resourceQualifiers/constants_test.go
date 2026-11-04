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

package resourceQualifiers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSystemVariable(t *testing.T) {
	cases := []struct {
		name     string
		variable string
		want     bool
	}{
		{"fixed system variable", "DEVTRON_APP_NAME", true},
		{"another fixed system variable", "DEVTRON_ENV_NAME", true},
		{"user-defined variable", "FULL_GIT_URL", false},
		{"user-defined variable starting with DEVTRON", "DEVTRON_CUSTOM_THING", false},
		{"empty string", "", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, IsSystemVariable(c.variable))
		})
	}
}
