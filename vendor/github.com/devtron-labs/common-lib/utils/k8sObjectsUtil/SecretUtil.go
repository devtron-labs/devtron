/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8sObjectsUtil

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/bean"
	yamlUtil "github.com/devtron-labs/common-lib/utils/yaml"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

func HideValuesIfSecretForWholeYamlInput(wholeYamlFileContent string) (string, error) {
	manifests, err := yamlUtil.SplitYAMLs([]byte(wholeYamlFileContent))
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, manifest := range manifests {
		modifiedManifest, err := HideValuesIfSecret(manifest.DeepCopy())
		if err != nil {
			return "", err
		}
		yamlStr, err := yaml.Marshal(modifiedManifest.UnstructuredContent())
		if err != nil {
			return "", err
		}
		sb.WriteString(bean.YamlSeparator)
		sb.WriteString(string(yamlStr))
	}

	return sb.String(), nil
}

func HideValuesIfSecretForManifestStringInput(manifest string, kind string, group string) (string, error) {
	if !isSecret(kind, group) {
		return manifest, nil
	}

	decoder, err := runtime.Decode(unstructured.UnstructuredJSONScheme, []byte(manifest))
	if err != nil {
		return "", err
	}

	modifiedSecret, err := HideValuesIfSecret(decoder.(*unstructured.Unstructured))
	if err != nil {
		return "", err
	}

	outBytes, err := runtime.Encode(unstructured.UnstructuredJSONScheme, modifiedSecret)
	if err != nil {
		return "", err
	}

	return string(outBytes), nil
}

func HideValuesIfSecret(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if !isSecret(obj.GetKind(), obj.GroupVersionKind().Group) {
		return obj, nil
	}

	// hide secret
	_, obj, err := hideSecretData(nil, obj)
	if err != nil {
		return nil, err
	}
	return obj, err
}

// HideSecretData replaces secret data values in specified target, live secrets and in last applied configuration of live secret with stars. Also preserves differences between
// target, live and last applied config values. E.g. if all three are equal the values would be replaced with same number of stars. If all the are different then number of stars
// in replacement should be different.
func hideSecretData(target *unstructured.Unstructured, live *unstructured.Unstructured) (*unstructured.Unstructured, *unstructured.Unstructured, error) {
	var orig *unstructured.Unstructured
	if live != nil {
		orig = getLastAppliedConfigAnnotation(live)
		live = live.DeepCopy()
	}
	if target != nil {
		target = target.DeepCopy()
	}

	keys := map[string]bool{}
	for _, obj := range []*unstructured.Unstructured{target, live, orig} {
		if obj == nil {
			continue
		}
		normalizeSecret(obj)
		if data, found, err := unstructured.NestedMap(obj.Object, "data"); found && err == nil {
			for k := range data {
				keys[k] = true
			}
		}
	}

	for k := range keys {
		// we use "+" rather than the more common "*"
		nextReplacement := "++++++++"
		valToReplacement := make(map[string]string)
		for _, obj := range []*unstructured.Unstructured{target, live, orig} {
			var data map[string]interface{}
			if obj != nil {
				var err error
				data, _, err = unstructured.NestedMap(obj.Object, "data")
				if err != nil {
					return nil, nil, err
				}
			}
			if data == nil {
				data = make(map[string]interface{})
			}
			valData, ok := data[k]
			if !ok {
				continue
			}
			val := toString(valData)
			replacement, ok := valToReplacement[val]
			if !ok {
				replacement = nextReplacement
				nextReplacement = nextReplacement + "++++"
				valToReplacement[val] = replacement
			}
			data[k] = replacement
			err := unstructured.SetNestedField(obj.Object, data, "data")
			if err != nil {
				return nil, nil, err
			}
		}
	}
	if live != nil && orig != nil {
		annotations := live.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		lastAppliedData, err := json.Marshal(orig)
		if err != nil {
			return nil, nil, err
		}
		annotations[corev1.LastAppliedConfigAnnotation] = string(lastAppliedData)
		live.SetAnnotations(annotations)
	}
	return target, live, nil
}

func getLastAppliedConfigAnnotation(live *unstructured.Unstructured) *unstructured.Unstructured {
	if live == nil {
		return nil
	}
	annots := live.GetAnnotations()
	if annots == nil {
		return nil
	}
	lastAppliedStr, ok := annots[corev1.LastAppliedConfigAnnotation]
	if !ok {
		return nil
	}
	var obj unstructured.Unstructured
	err := json.Unmarshal([]byte(lastAppliedStr), &obj)
	if err != nil {
		return nil
	}
	return &obj
}

// NormalizeSecret mutates the supplied object and encodes stringData to data, and converts nils to
// empty strings. If the object is not a secret, or is an invalid secret, then returns the same object.
func normalizeSecret(un *unstructured.Unstructured) {
	if un == nil {
		return
	}
	gvk := un.GroupVersionKind()
	if gvk.Group != "" || gvk.Kind != "Secret" {
		return
	}
	var secret corev1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, &secret)
	if err != nil {
		return
	}
	// We normalize nils to empty string to handle: https://github.com/argoproj/argo-cd/issues/943
	for k, v := range secret.Data {
		if len(v) == 0 {
			secret.Data[k] = []byte("")
		}
	}
	if len(secret.StringData) > 0 {
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		for k, v := range secret.StringData {
			secret.Data[k] = []byte(v)
		}
		delete(un.Object, "stringData")
	}
	newObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&secret)
	if err != nil {
		return
	}
	if secret.Data != nil {
		err = unstructured.SetNestedField(un.Object, newObj["data"], "data")
		if err != nil {
			return
		}
	}
}

func toString(val interface{}) string {
	if val == nil {
		return ""
	}
	return fmt.Sprintf("%s", val)
}

func isSecret(kind string, group string) bool {
	return kind == "Secret" && group == ""
}
