package bean

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"strings"
	"time"
)

func ConvertArtifactEntityToModel(ArtifactDaos []repository.CiArtifact) []CiArtifactBean {
	ciArtifacts := make([]CiArtifactBean, 0, len(ArtifactDaos))
	for _, artifactDao := range ArtifactDaos {
		ciArtifact := getCiArtifactBean(artifactDao)
		ciArtifacts = append(ciArtifacts, ciArtifact)
	}
	return ciArtifacts
}

func getCiArtifactBean(artifactDao repository.CiArtifact) CiArtifactBean {
	mInfo, err := ParseMaterialInfo([]byte(artifactDao.MaterialInfo), artifactDao.DataSource)
	if err != nil {
		mInfo = []byte("[]")
	}
	return CiArtifactBean{
		Id:                     artifactDao.Id,
		Image:                  artifactDao.Image,
		ImageDigest:            artifactDao.ImageDigest,
		MaterialInfo:           mInfo,
		ScanEnabled:            artifactDao.ScanEnabled,
		Scanned:                artifactDao.Scanned,
		Deployed:               artifactDao.Deployed,
		DeployedTime:           formatDate(artifactDao.DeployedTime, LayoutRFC3339),
		ExternalCiPipelineId:   artifactDao.ExternalCiPipelineId,
		ParentCiArtifact:       artifactDao.ParentCiArtifact,
		CreatedTime:            formatDate(artifactDao.CreatedOn, LayoutRFC3339),
		CiPipelineId:           artifactDao.PipelineId,
		DataSource:             artifactDao.DataSource,
		CredentialsSourceType:  artifactDao.CredentialsSourceType,
		CredentialsSourceValue: artifactDao.CredentialSourceValue,
	}
}

func formatDate(t time.Time, layout string) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(layout)
}

func ParseMaterialInfo(materialInfo json.RawMessage, source string) (json.RawMessage, error) {
	if source != repository.GOCD && source != repository.CI_RUNNER && source != repository.WEBHOOK && source != repository.EXT && source != repository.PRE_CD && source != repository.POST_CD && source != repository.POST_CI {
		return nil, fmt.Errorf("datasource: %s not supported", source)
	}
	var ciMaterials []repository.CiMaterialInfo
	err := json.Unmarshal(materialInfo, &ciMaterials)
	if err != nil {
		println("material info", materialInfo)
		println("unmarshal error for material info", "err", err)
	}
	var scmMapList []map[string]string

	for _, material := range ciMaterials {
		scmMap := map[string]string{}
		var url string
		if material.Material.Type == "git" {
			url = material.Material.GitConfiguration.URL
		} else if material.Material.Type == "scm" {
			url = material.Material.ScmConfiguration.URL
		} else {
			return nil, fmt.Errorf("unknown material type:%s ", material.Material.Type)
		}
		if material.Modifications != nil && len(material.Modifications) > 0 {
			_modification := material.Modifications[0]

			revision := _modification.Revision
			url = strings.TrimSpace(url)

			_webhookDataStr := ""
			_webhookDataByteArr, err := json.Marshal(_modification.WebhookData)
			if err == nil {
				_webhookDataStr = string(_webhookDataByteArr)
			}

			scmMap["url"] = url
			scmMap["revision"] = revision
			scmMap["modifiedTime"] = _modification.ModifiedTime
			scmMap["author"] = _modification.Author
			scmMap["message"] = _modification.Message
			scmMap["tag"] = _modification.Tag
			scmMap["webhookData"] = _webhookDataStr
			scmMap["branch"] = _modification.Branch
		}
		scmMapList = append(scmMapList, scmMap)
	}
	mInfo, err := json.Marshal(scmMapList)
	return mInfo, err
}
