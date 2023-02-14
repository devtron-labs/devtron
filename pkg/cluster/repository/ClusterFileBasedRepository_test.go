package repository

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClusterRepositoryFileBased_FindAll(t *testing.T) {
	t.SkipNow()
	t.Run("abcd", func(t *testing.T) {
		//sugaredLogger, err := util.InitLogger()
		//assert.Nil(t, err)
		repositoryFileBased := NewClusterRepositoryFileBased(nil)
		cluster := &Cluster{
			Id:          1,
			ClusterName: "k8s16-cluster",
			ServerUrl:   "http://127.0.0.1:16443",
			Config:      map[string]string{"bearer_token": "defgh"},
			Active:      false,
		}
		err := repositoryFileBased.Update(cluster)
		assert.Nil(t, err)
		clusters, err := repositoryFileBased.FindAll()
		assert.Nil(t, err)
		for _, cluster := range clusters {
			fmt.Println(cluster)
		}
	})
}
