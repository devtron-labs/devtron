package bean

type InstalledAppMin struct {
	// Installed App details
	Id     int
	Active bool
	// Deprecated; currently in use for backward compatibility
	GitOpsRepoName string
	// Deprecated; use deployment_config table instead GitOpsRepoName has been migrated to GitOpsRepoUrl; Make sure to migrate from GitOpsRepoName for future flows
	GitOpsRepoUrl      string
	IsCustomRepository bool
	// Deprecated; use deployment_config table instead
	DeploymentAppType          string
	DeploymentAppDeleteRequest bool
	EnvironmentId              int
	AppId                      int
}

type InstalledAppWithAppDetails struct {
	InstalledAppMin
	// Extra App details
	AppName         string
	AppOfferingMode string
	TeamId          int
}

type InstalledAppWithEnvDetails struct {
	InstalledAppWithAppDetails
	// Extra Environment details
	EnvironmentName       string
	EnvironmentIdentifier string
	Namespace             string
	ClusterId             int
}

type InstalledAppDeleteRequest struct {
	InstalledAppId  int
	AppName         string
	AppId           int
	EnvironmentId   int
	AppOfferingMode string
	ClusterId       int
	Namespace       string
}

type InstalledAppWithEnvAndClusterDetails struct {
	InstalledAppWithEnvDetails
	// Extra Cluster details
	ClusterName string
}

func (i *InstalledAppWithEnvAndClusterDetails) GetInstalledAppMin() *InstalledAppMin {
	if i == nil {
		return nil
	}
	return &i.InstalledAppMin
}

func (i *InstalledAppWithEnvAndClusterDetails) GetInstalledAppWithAppDetails() *InstalledAppWithAppDetails {
	if i == nil {
		return nil
	}
	return &i.InstalledAppWithAppDetails
}

func (i *InstalledAppWithEnvAndClusterDetails) GetInstalledAppWithEnvDetails() *InstalledAppWithEnvDetails {
	if i == nil {
		return nil
	}
	return &i.InstalledAppWithEnvDetails
}
