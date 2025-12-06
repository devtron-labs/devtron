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

package gitProvider

import (
	"context"
	"github.com/devtron-labs/common-lib/securestore"
	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/constants"
	constants2 "github.com/devtron-labs/devtron/internal/sql/constants"
	"github.com/devtron-labs/devtron/internal/util"
	bean2 "github.com/devtron-labs/devtron/pkg/build/git/gitProvider/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/repository"
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/juju/errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GitRegistryConfig interface {
	Create(request *bean2.GitRegistry) (*bean2.GitRegistry, error)
	Update(request *bean2.GitRegistry) (*bean2.GitRegistry, error)
	Delete(request *bean2.GitRegistry) error
}
type GitRegistryConfigImpl struct {
	logger              *zap.SugaredLogger
	gitProviderRepo     repository.GitProviderRepository
	GitSensorGrpcClient gitSensor.Client
}

func NewGitRegistryConfigImpl(logger *zap.SugaredLogger, gitProviderRepo repository.GitProviderRepository,
	GitSensorClient gitSensor.Client) *GitRegistryConfigImpl {
	return &GitRegistryConfigImpl{
		logger:              logger,
		gitProviderRepo:     gitProviderRepo,
		GitSensorGrpcClient: GitSensorClient,
	}
}

func (impl GitRegistryConfigImpl) Create(request *bean2.GitRegistry) (*bean2.GitRegistry, error) {
	impl.logger.Debugw("get repo create request", "req", request)
	exist, err := impl.gitProviderRepo.ProviderExists(request.Url)
	if err != nil {
		impl.logger.Errorw("error in fetch ", "url", request.Url, "err", err)
		err = &util.ApiError{
			//Code:            constants.GitProviderCreateFailed,
			InternalMessage: "git provider creation failed, error in fetching by url",
			UserMessage:     "git provider creation failed, error in fetching by url",
		}
		return nil, err
	}
	if exist {
		impl.logger.Warnw("repo already exists", "url", request.Url)
		err = &util.ApiError{
			Code:            constants.GitProviderCreateFailedAlreadyExists,
			InternalMessage: "git provider already exists",
			UserMessage:     "git provider already exists",
		}
		return nil, errors.NewAlreadyExists(err, request.Url)
	}
	provider := &repository.GitProvider{
		Id:                    request.Id,
		Name:                  request.Name,
		Url:                   request.Url,
		UserName:              request.UserName,
		Password:              securestore.ToEncryptedString(request.Password),
		SshPrivateKey:         securestore.ToEncryptedString(request.SshPrivateKey),
		AccessToken:           securestore.ToEncryptedString(request.AccessToken),
		AuthMode:              request.AuthMode,
		Active:                request.Active,
		Deleted:               false,
		GitHostId:             request.GitHostId,
		EnableTLSVerification: request.EnableTLSVerification,
		AuditLog:              sql.AuditLog{CreatedBy: request.UserId, CreatedOn: time.Now(), UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}

	if request.EnableTLSVerification {
		if len(request.TLSConfig.CaData) > 0 {
			provider.CaCert = request.TLSConfig.CaData
		}
		if len(request.TLSConfig.TLSKeyData) > 0 && len(request.TLSConfig.TLSCertData) > 0 {
			provider.TlsKey = request.TLSConfig.TLSKeyData
			provider.TlsCert = request.TLSConfig.TLSCertData
		}

		if !request.IsCADataPresent {
			provider.CaCert = ""
		}
		if !request.IsTLSCertDataPresent {
			provider.TlsCert = ""
		}
		if !request.IsTLSKeyDataPresent {
			provider.TlsKey = ""
		}

		if (len(provider.TlsKey) > 0 && len(provider.TlsCert) == 0) || (len(provider.TlsKey) == 0 && len(provider.TlsCert) > 0) {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusPreconditionFailed,
				Code:            constants.GitProviderUpdateRequestIsInvalid,
				InternalMessage: "git provider failed to update in db",
				UserMessage:     "git provider failed to update in db",
			}
		}
		if len(provider.TlsKey) == 0 && len(provider.TlsCert) == 0 && len(provider.CaCert) == 0 {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusPreconditionFailed,
				Code:            constants.GitProviderUpdateRequestIsInvalid,
				InternalMessage: "git provider failed to update in db",
				UserMessage:     "git provider failed to update in db",
			}
		}
	}

	provider.SshPrivateKey = securestore.ToEncryptedString(ModifySshPrivateKey(provider.SshPrivateKey.String(), provider.AuthMode))
	err = impl.gitProviderRepo.Save(provider)
	if err != nil {
		impl.logger.Errorw("error in saving git repo config", "data", provider, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderCreateFailedInDb,
			InternalMessage: "git provider failed to create in db",
			UserMessage:     "git provider failed to create in db",
		}
		return nil, err
	}
	err = impl.UpdateGitSensor(provider)
	if err != nil {
		impl.logger.Errorw("error in updating git repo config on sensor", "data", provider, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInSync,
			InternalMessage: err.Error(),
			UserMessage:     "git provider failed to update in sync",
		}
		return nil, err
	}
	request.Id = provider.Id
	return request, nil
}

func (impl GitRegistryConfigImpl) Update(request *bean2.GitRegistry) (*bean2.GitRegistry, error) {
	impl.logger.Debugw("get repo create request", "req", request)

	/*
		exist, err := impl.gitProviderRepo.ProviderExists(request.RedirectionUrl)
		if err != nil {
			impl.logger.Errorw("error in fetch ", "url", request.RedirectionUrl, "err", err)
			return nil, err
		}
		if exist {
			impl.logger.Infow("repo already exists", "url", request.RedirectionUrl)
			return nil, errors.NewAlreadyExists(err, request.RedirectionUrl)
		}
	*/

	providerId := strconv.Itoa(request.Id)
	existingProvider, err0 := impl.gitProviderRepo.FindOne(providerId)
	if err0 != nil {
		impl.logger.Errorw("No matching entry found for update.", "err", err0)
		err0 = &util.ApiError{
			Code:            constants.GitProviderUpdateProviderNotExists,
			InternalMessage: "git provider update failed, provider does not exist",
			UserMessage:     "git provider update failed, provider does not exist",
		}
		return nil, err0
	}
	if request.Password == "" {
		request.Password = existingProvider.Password.String()
	}
	if request.SshPrivateKey == "" {
		request.SshPrivateKey = existingProvider.SshPrivateKey.String()
	}
	if request.AccessToken == "" {
		request.AccessToken = existingProvider.AccessToken.String()
	}
	provider := &repository.GitProvider{
		Name:                  request.Name,
		Url:                   request.Url,
		Id:                    request.Id,
		AuthMode:              request.AuthMode,
		Password:              securestore.ToEncryptedString(request.Password),
		Active:                request.Active,
		AccessToken:           securestore.ToEncryptedString(request.AccessToken),
		SshPrivateKey:         securestore.ToEncryptedString(request.SshPrivateKey),
		UserName:              request.UserName,
		GitHostId:             request.GitHostId,
		EnableTLSVerification: request.EnableTLSVerification,
		AuditLog:              sql.AuditLog{CreatedBy: existingProvider.CreatedBy, CreatedOn: existingProvider.CreatedOn, UpdatedOn: time.Now(), UpdatedBy: request.UserId},
	}

	if request.AuthMode != constants2.AUTH_MODE_USERNAME_PASSWORD {
		provider.Password = ""
		provider.TlsCert = ""
		provider.TlsKey = ""
		provider.CaCert = ""
	}

	if provider.EnableTLSVerification {

		provider.TlsKey = existingProvider.TlsKey
		provider.TlsCert = existingProvider.TlsCert
		provider.CaCert = existingProvider.CaCert

		if len(request.TLSConfig.CaData) > 0 {
			provider.CaCert = request.TLSConfig.CaData
		}
		if len(request.TLSConfig.TLSKeyData) > 0 && len(request.TLSConfig.TLSCertData) > 0 {
			provider.TlsKey = request.TLSConfig.TLSKeyData
			provider.TlsCert = request.TLSConfig.TLSCertData
		}

		if !request.IsCADataPresent {
			provider.CaCert = ""
		}
		if !request.IsTLSCertDataPresent {
			provider.TlsCert = ""
		}
		if !request.IsTLSKeyDataPresent {
			provider.TlsKey = ""
		}

		if (len(provider.TlsKey) > 0 && len(provider.TlsCert) == 0) || (len(provider.TlsKey) == 0 && len(provider.TlsCert) > 0) {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusPreconditionFailed,
				Code:            constants.GitProviderUpdateRequestIsInvalid,
				InternalMessage: "git provider failed to update in db",
				UserMessage:     "git provider failed to update in db",
			}
		}
		if len(provider.TlsKey) == 0 && len(provider.TlsCert) == 0 && len(provider.CaCert) == 0 {
			return nil, &util.ApiError{
				HttpStatusCode:  http.StatusPreconditionFailed,
				Code:            constants.GitProviderUpdateRequestIsInvalid,
				InternalMessage: "git provider failed to update in db",
				UserMessage:     "git provider failed to update in db",
			}
		}
	}

	provider.SshPrivateKey = securestore.ToEncryptedString(ModifySshPrivateKey(provider.SshPrivateKey.String(), provider.AuthMode))
	err := impl.gitProviderRepo.Update(provider)
	if err != nil {
		impl.logger.Errorw("error in updating git repo config", "data", provider, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInDb,
			InternalMessage: "git provider failed to update in db",
			UserMessage:     "git provider failed to update in db",
		}
		return nil, err
	}
	request.Id = provider.Id
	err = impl.UpdateGitSensor(provider)
	if err != nil {
		impl.logger.Errorw("error in updating git repo config on sensor", "data", provider, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInSync,
			InternalMessage: err.Error(),
			UserMessage:     "git provider failed to update in sync",
		}
		return nil, err
	}
	return request, nil
}

func (impl GitRegistryConfigImpl) Delete(request *bean2.GitRegistry) error {
	providerId := strconv.Itoa(request.Id)
	gitProviderConfig, err := impl.gitProviderRepo.FindOne(providerId)
	if err != nil {
		impl.logger.Errorw("No matching entry found for delete.", "id", request.Id, "err", err)
		return err
	}
	deleteReq := gitProviderConfig
	deleteReq.UpdatedOn = time.Now()
	deleteReq.UpdatedBy = request.UserId
	err = impl.gitProviderRepo.MarkProviderDeleted(&deleteReq)
	if err != nil {
		impl.logger.Errorw("err in deleting git account", "id", request.Id, "err", err)
		return err
	}
	deleteReq.Active = false
	err = impl.UpdateGitSensor(&deleteReq)
	if err != nil {
		impl.logger.Errorw("error in updating git repo config on sensor after deleting", "deleteReq", deleteReq, "err", err)
		err = &util.ApiError{
			Code:            constants.GitProviderUpdateFailedInSync,
			InternalMessage: err.Error(),
			UserMessage:     "git provider failed to update in sync",
		}
		return err
	}
	return nil
}

func (impl GitRegistryConfigImpl) UpdateGitSensor(provider *repository.GitProvider) error {
	sensorGitProvider := &gitSensor.GitProvider{
		Id:                    provider.Id,
		Name:                  provider.Name,
		Url:                   provider.Url,
		UserName:              provider.UserName,
		Password:              provider.Password.String(),
		SshPrivateKey:         provider.SshPrivateKey.String(),
		AccessToken:           provider.AccessToken.String(),
		Active:                provider.Active,
		AuthMode:              provider.AuthMode,
		CaCert:                provider.CaCert,
		TlsCert:               provider.TlsCert,
		TlsKey:                provider.TlsKey,
		EnableTlsVerification: provider.EnableTLSVerification,
	}
	return impl.GitSensorGrpcClient.SaveGitProvider(context.Background(), sensorGitProvider)
}

// Modifying Ssh Private Key because Ssh key authentication requires a new-line at the end of string & there are chances that user skips sending \n
func ModifySshPrivateKey(sshPrivateKey string, authMode constants2.AuthMode) string {
	if authMode == constants2.AUTH_MODE_SSH {
		if !strings.HasSuffix(sshPrivateKey, "\n") {
			sshPrivateKey += "\n"
		}
	}
	return sshPrivateKey
}
