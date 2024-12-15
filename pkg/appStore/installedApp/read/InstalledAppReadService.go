package read

import (
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/adapter"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read/bean"
)

type InstalledAppReadService interface {
	InstalledAppReadServiceEA
	// GetInstalledAppByGitHash will return the installed app by git hash.
	// Only delete specific details are fetched.
	// Refer bean.InstalledAppDeleteRequest for more details.
	GetInstalledAppByGitHash(gitHash string) (*bean.InstalledAppDeleteRequest, error)
	// GetInstalledAppByGitOpsAppName will return all the active installed_apps with matching `<app_name>-<environment_name>`.
	// Only the minimum details are fetched.
	// Refer bean.InstalledAppMin for more details.
	GetInstalledAppByGitOpsAppName(acdAppName string) (*bean.InstalledAppMin, error)
}

type InstalledAppReadServiceImpl struct {
	*InstalledAppReadServiceEAImpl
}

func NewInstalledAppReadServiceImpl(
	installedAppReadServiceEAImpl *InstalledAppReadServiceEAImpl) *InstalledAppReadServiceImpl {
	return &InstalledAppReadServiceImpl{
		InstalledAppReadServiceEAImpl: installedAppReadServiceEAImpl,
	}
}

func (impl *InstalledAppReadServiceImpl) GetInstalledAppByGitHash(gitHash string) (*bean.InstalledAppDeleteRequest, error) {
	installedAppModel, err := impl.installedAppRepository.GetInstalledAppByGitHash(gitHash)
	if err != nil {
		impl.logger.Errorw("error while fetching installed app by git hash", "gitHash", gitHash, "error", err)
		return nil, err
	}
	return adapter.GetInstalledAppDeleteRequest(installedAppModel), nil
}

func (impl *InstalledAppReadServiceImpl) GetInstalledAppByGitOpsAppName(acdAppName string) (*bean.InstalledAppMin, error) {
	installedAppModel, err := impl.installedAppRepository.GetInstalledAppByGitOpsAppName(acdAppName)
	if err != nil {
		impl.logger.Errorw("error while fetching installed app by GitOps app name", "acdAppName", acdAppName, "error", err)
		return nil, err
	}
	return adapter.GetInstalledAppInternal(installedAppModel).GetInstalledAppMin(), nil
}
