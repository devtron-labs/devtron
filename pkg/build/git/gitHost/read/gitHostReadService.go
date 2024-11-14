package read

import (
	"github.com/devtron-labs/devtron/pkg/attributes"
	bean2 "github.com/devtron-labs/devtron/pkg/build/git/gitHost/bean"
	"github.com/devtron-labs/devtron/pkg/build/git/gitHost/repository"
	"go.uber.org/zap"
)

type GitHostReadService interface {
	GetAll() ([]bean2.GitHostRequest, error)
	GetById(id int) (*bean2.GitHostRequest, error)
	GetByName(uniqueName string) (*bean2.GitHostRequest, error)
}
type GitHostReadServiceImpl struct {
	logger           *zap.SugaredLogger
	gitHostRepo      repository.GitHostRepository
	attributeService attributes.AttributesService
}

func NewGitHostReadServiceImpl(logger *zap.SugaredLogger,
	gitHostRepo repository.GitHostRepository,
	attributeService attributes.AttributesService) *GitHostReadServiceImpl {
	return &GitHostReadServiceImpl{
		gitHostRepo:      gitHostRepo,
		logger:           logger,
		attributeService: attributeService,
	}

}

// get all git hosts
func (impl *GitHostReadServiceImpl) GetAll() ([]bean2.GitHostRequest, error) {
	impl.logger.Debug("get all hosts request")
	hosts, err := impl.gitHostRepo.FindAll()
	if err != nil {
		impl.logger.Errorw("error in fetching all git hosts", "err", err)
		return nil, err
	}

	var gitHosts []bean2.GitHostRequest
	for _, host := range hosts {
		//display_name can be null for old data hence checking for name field
		displayName := host.DisplayName
		if len(displayName) == 0 {
			displayName = host.Name
		}
		hostRes := bean2.GitHostRequest{
			Id:     host.Id,
			Name:   displayName,
			Active: host.Active,
		}
		gitHosts = append(gitHosts, hostRes)
	}
	return gitHosts, err
}

// get git host by Id
func (impl *GitHostReadServiceImpl) GetById(id int) (*bean2.GitHostRequest, error) {
	impl.logger.Debug("get hosts request for Id", id)
	host, err := impl.gitHostRepo.FindOneById(id)
	if err != nil {
		impl.logger.Errorw("error in fetching git host", "err", err)
		return nil, err
	}

	return impl.processAndReturnGitHost(host)

}

// get git host by Name
func (impl *GitHostReadServiceImpl) GetByName(uniqueName string) (*bean2.GitHostRequest, error) {
	impl.logger.Debug("get hosts request for name", uniqueName)
	host, err := impl.gitHostRepo.FindOneByName(uniqueName)
	if err != nil {
		impl.logger.Errorw("error in fetching git host", "err", err)
		return nil, err
	}

	return impl.processAndReturnGitHost(host)
}

func (impl *GitHostReadServiceImpl) processAndReturnGitHost(host repository.GitHost) (*bean2.GitHostRequest, error) {
	// get orchestrator host
	orchestratorHost, err := impl.attributeService.GetByKey("url")
	if err != nil {
		impl.logger.Errorw("error in fetching orchestrator host url from db", "err", err)
		return nil, err
	}

	var webhookUrlPrepend string
	if orchestratorHost == nil || len(orchestratorHost.Value) == 0 {
		webhookUrlPrepend = "{HOST_URL_PLACEHOLDER}"
	} else {
		webhookUrlPrepend = orchestratorHost.Value
	}
	webhookUrl := webhookUrlPrepend + host.WebhookUrl

	gitHost := &bean2.GitHostRequest{
		Id:              host.Id,
		Name:            host.Name,
		Active:          host.Active,
		WebhookUrl:      webhookUrl,
		WebhookSecret:   host.WebhookSecret,
		EventTypeHeader: host.EventTypeHeader,
		SecretHeader:    host.SecretHeader,
		SecretValidator: host.SecretValidator,
	}

	return gitHost, err
}
