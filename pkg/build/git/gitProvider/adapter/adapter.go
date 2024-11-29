package adapter

import (
	"github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/build/git/gitProvider/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/repository"
)

func ConvertGitRegistryDtoToBean(provider repository.GitProvider, withSensitiveData bool) bean2.GitRegistry {
	registryBean := bean2.GitRegistry{
		Id:                    provider.Id,
		Name:                  provider.Name,
		Url:                   provider.Url,
		UserName:              provider.UserName,
		Password:              "",
		AuthMode:              provider.AuthMode,
		AccessToken:           "",
		SshPrivateKey:         "",
		Active:                provider.Active,
		UserId:                provider.CreatedBy,
		GitHostId:             provider.GitHostId,
		EnableTLSVerification: provider.EnableTLSVerification,
		TLSConfig: bean.TLSConfig{
			CaData:      "",
			TLSCertData: "",
			TLSKeyData:  "",
		},
		IsCADataPresent:      len(provider.CaCert) > 0,
		IsTLSKeyDataPresent:  len(provider.TlsKey) > 0,
		IsTLSCertDataPresent: len(provider.TlsCert) > 0,
	}
	if withSensitiveData {
		registryBean.Password = provider.Password
		registryBean.AccessToken = provider.AccessToken
		registryBean.SshPrivateKey = provider.SshPrivateKey
	}
	return registryBean
}
