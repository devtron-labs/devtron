package repository

type ClusterDescriptionFileBasedRepositoryImpl struct {
}

func NewClusterDescriptionFileBasedRepository() *ClusterDescriptionFileBasedRepositoryImpl {
	return &ClusterDescriptionFileBasedRepositoryImpl{}
}

func (impl ClusterDescriptionFileBasedRepositoryImpl) FindByClusterIdWithClusterDetails(clusterId int) (*ClusterDescription, error) {
	//TODO implement me
	panic("implement me")
}

