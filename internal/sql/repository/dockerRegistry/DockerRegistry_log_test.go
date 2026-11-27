/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package repository

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/devtron-labs/common-lib/securestore"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRegistryModelLogOmitsSecrets(t *testing.T) {
	// Every credential-bearing column holds a synthetic marker that must never reach the log.
	secrets := []string{"test-pull-credential", "test-url-password", "test-username", "test-password",
		"test-access-key", "test-secret-key", "test-cert"}
	ips := DockerRegistryIpsConfig{Id: 3, DockerArtifactStoreId: "test-registry", CredentialType: "CUSTOM_CREDENTIAL",
		CredentialValue: "test-pull-credential", Active: true}
	store := DockerArtifactStore{
		Id:                     "test-registry",
		RegistryType:           REGISTRYTYPE_DOCKER_HUB,
		IsOCICompliantRegistry: true,
		IsDefault:              true,
		Active:                 true,
		RegistryURL:            "https://user:test-url-password@registry.example.com",
		Username:               "test-username",
		Password:               securestore.EncryptedString("test-password"),
		AWSAccessKeyId:         "test-access-key",
		AWSSecretAccessKey:     securestore.EncryptedString("test-secret-key"),
		Cert:                   "test-cert",
		IpsConfig:              &ips,
	}
	wantStore := map[string]interface{}{
		"id":                     "test-registry",
		"registryType":           "docker-hub",
		"isOCICompliantRegistry": true,
		"isDefault":              true,
		"active":                 true,
	}
	wantIps := map[string]interface{}{
		"id":                    float64(3),
		"dockerArtifactStoreId": "test-registry",
		"credentialType":        "CUSTOM_CREDENTIAL",
		"active":                true,
	}

	// Services log these models as pointers; values must be safe too.
	tests := []struct {
		name    string
		payload interface{}
		want    map[string]interface{}
	}{
		{"registry value", store, wantStore},
		{"registry pointer", &store, wantStore},
		{"pull config value", ips, wantIps},
		{"pull config pointer", &ips, wantIps},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
			logger := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(&out), zapcore.DebugLevel)).Sugar()
			logger.Infow("updated repository", "repository", tt.payload)
			line := out.String()

			for _, secret := range secrets {
				if strings.Contains(line, secret) {
					t.Errorf("log line contains secret %q: %s", secret, line)
				}
			}
			var entry struct {
				Repository map[string]interface{} `json:"repository"`
			}
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				t.Fatalf("log line is not valid JSON: %v", err)
			}
			if !reflect.DeepEqual(entry.Repository, tt.want) {
				t.Errorf("logged model = %v, want %v", entry.Repository, tt.want)
			}
		})
	}
}
