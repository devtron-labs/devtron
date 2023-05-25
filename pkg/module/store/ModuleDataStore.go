package moduleDataStore

type ModuleDataStore struct {
	ModuleStatusCronInProgress bool
}

func InitModuleDataStore() *ModuleDataStore {
	return &ModuleDataStore{}
}
