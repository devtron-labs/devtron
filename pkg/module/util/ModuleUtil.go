/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package moduleUtil

import (
	"fmt"
	"strings"
)

func BuildAllModuleEnableKeys(basePath string, moduleName string) []string {
	var keys []string
	keys = append(keys, BuildModuleEnableKey(basePath, moduleName))
	if strings.Contains(moduleName, ".") {
		parent := strings.Split(moduleName, ".")[0]
		keys = append(keys, BuildModuleEnableKey(basePath, parent))
	}
	return keys
}

func BuildModuleEnableKey(basePath string, moduleName string) string {
	if len(basePath) > 0 {
		return fmt.Sprintf("%s.%s.%s", basePath, moduleName, "enabled")
	}
	return fmt.Sprintf("%s.%s", moduleName, "enabled")
}
