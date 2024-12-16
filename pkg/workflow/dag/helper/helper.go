package helper

import (
	"bytes"
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository"
)

func GetMaterialInfoJson(materialInfo json.RawMessage) ([]byte, error) {
	var matJson []byte
	materialJson, err := materialInfo.MarshalJSON()
	if err != nil {
		return matJson, err
	}
	dst := new(bytes.Buffer)
	err = json.Compact(dst, materialJson)
	if err != nil {
		return matJson, err
	}
	matJson = dst.Bytes()
	return matJson, nil
}

func UpdateScanStatusInCiArtifact(ciArtifact *repository.CiArtifact, isScanPluginConfigured, isScanningDoneViaPlugin bool) {
	if isScanPluginConfigured {
		ciArtifact.ScanEnabled = true
	}
	if isScanningDoneViaPlugin {
		ciArtifact.Scanned = true
	}
}
