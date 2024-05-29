/*
 * Copyright (c) 2024. Devtron Inc.
 */

package moduleDataStore

type ModuleDataStore struct {
	ModuleStatusCronInProgress bool
}

func InitModuleDataStore() *ModuleDataStore {
	return &ModuleDataStore{}
}
