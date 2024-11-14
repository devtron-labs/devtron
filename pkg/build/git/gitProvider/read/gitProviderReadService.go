package read

import (
	"github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/build/git/gitProvider/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitProvider/repository"
	"go.uber.org/zap"
)

type GitProviderReadService interface {
	GetAll() ([]bean2.GitRegistry, error)
	FetchAllGitProviders() ([]bean2.GitRegistry, error)
	FetchOneGitProvider(id string) (*bean2.GitRegistry, error)
	FindByUrl(url string) (*bean2.GitRegistry, error)
}

type GitProviderReadServiceImpl struct {
	logger                *zap.SugaredLogger
	gitProviderRepository repository.GitProviderRepository
}

func NewGitProviderReadService(logger *zap.SugaredLogger,
	gitProviderRepository repository.GitProviderRepository) *GitProviderReadServiceImpl {
	return &GitProviderReadServiceImpl{
		logger:                logger,
		gitProviderRepository: gitProviderRepository,
	}
}

// get all active git providers
func (impl *GitProviderReadServiceImpl) GetAll() ([]bean2.GitRegistry, error) {
	impl.logger.Debug("get all provider request")
	providers, err := impl.gitProviderRepository.FindAllActiveForAutocomplete()
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}
	var gitProviders []bean2.GitRegistry
	for _, provider := range providers {
		providerRes := bean2.GitRegistry{
			Id:                    provider.Id,
			Name:                  provider.Name,
			Url:                   provider.Url,
			GitHostId:             provider.GitHostId,
			AuthMode:              provider.AuthMode,
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
		gitProviders = append(gitProviders, providerRes)
	}
	return gitProviders, err
}

func (impl *GitProviderReadServiceImpl) FetchAllGitProviders() ([]bean2.GitRegistry, error) {
	impl.logger.Debug("fetch all git providers from db")
	providers, err := impl.gitProviderRepository.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}
	var gitProviders []bean2.GitRegistry
	for _, provider := range providers {
		providerRes := bean2.GitRegistry{
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
		gitProviders = append(gitProviders, providerRes)
	}
	return gitProviders, err
}

func (impl *GitProviderReadServiceImpl) FetchOneGitProvider(providerId string) (*bean2.GitRegistry, error) {
	impl.logger.Debug("fetch git provider by ID from db")
	provider, err := impl.gitProviderRepository.FindOne(providerId)
	if err != nil {
		impl.logger.Errorw("error in fetch all git providers", "err", err)
		return nil, err
	}

	providerRes := &bean2.GitRegistry{
		Id:                    provider.Id,
		Name:                  provider.Name,
		Url:                   provider.Url,
		UserName:              provider.UserName,
		Password:              provider.Password,
		AuthMode:              provider.AuthMode,
		AccessToken:           provider.AccessToken,
		SshPrivateKey:         provider.SshPrivateKey,
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

	return providerRes, err
}
func (impl *GitProviderReadServiceImpl) FindByUrl(url string) (*bean2.GitRegistry, error) {
	provider, err := impl.gitProviderRepository.FindByUrl(url)
	if err != nil {
		impl.logger.Errorw("error in FindByUrl", "url", url, "err", err)
		return nil, err
	}
	gitRegistryBean := &bean2.GitRegistry{
		Id:                    provider.Id,
		Name:                  provider.Name,
		Url:                   provider.Url,
		UserName:              provider.UserName,
		Password:              provider.Password,
		AuthMode:              provider.AuthMode,
		AccessToken:           provider.AccessToken,
		SshPrivateKey:         provider.SshPrivateKey,
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
	return gitRegistryBean, nil
}
