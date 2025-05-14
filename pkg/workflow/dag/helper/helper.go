/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
