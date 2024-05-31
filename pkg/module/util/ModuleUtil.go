/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package moduleUtil

import (
	"fmt"
	"strings"
)

func BuildAllModuleEnableKeys(moduleName string) []string {
	var keys []string
	keys = append(keys, BuildModuleEnableKey(moduleName))
	if strings.Contains(moduleName, ".") {
		parent := strings.Split(moduleName, ".")[0]
		keys = append(keys, BuildModuleEnableKey(parent))
	}
	return keys
}

func BuildModuleEnableKey(moduleName string) string {
	return fmt.Sprintf("%s.%s", moduleName, "enabled")
}
