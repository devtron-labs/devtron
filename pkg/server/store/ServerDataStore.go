package serverDataStore

type ServerDataStore struct {
	CurrentVersion           string
	InstallerCrdObjectStatus string
}

func InitServerDataStore() *ServerDataStore {
	return &ServerDataStore{}
}