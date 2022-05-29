package apiToken

type ApiTokenSecretStore struct {
	Secret string
}

func InitApiTokenSecretStore() *ApiTokenSecretStore {
	return &ApiTokenSecretStore{}
}
