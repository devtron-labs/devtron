/*
 * Copyright (c) 2024. Devtron Inc.
 */

package serverDataStore

type ServerDataStore struct {
	CurrentVersion           string
	InstallerCrdObjectStatus string
	InstallerCrdObjectExists bool
}

func InitServerDataStore() *ServerDataStore {
	return &ServerDataStore{}
}
