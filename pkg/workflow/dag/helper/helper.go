package helper

import (
	"bytes"
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
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

// IsCdQualifiedForAutoTriggerForWebhookCiEvent returns bool, if a cd/pre-cd is qualified for auto trigger for a webhook ci event
func IsCdQualifiedForAutoTriggerForWebhookCiEvent(pipeline *pipelineConfig.Pipeline) bool {
	/*
		A cd is qualified for auto trigger for webhookCiEvent if it satisfies below two conditions:-
			1. If pre-cd exists and is set on auto.
			2. If only cd exists and is set on auto.
	*/
	if len(pipeline.PreTriggerType) > 0 && pipeline.PreTriggerType.IsAuto() {
		return true
	} else if len(pipeline.PreTriggerType) == 0 && pipeline.TriggerType.IsAuto() {
		return true
	}
	return false
}
