/*
 * Copyright (c) 2024. Devtron Inc.
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
 */

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
