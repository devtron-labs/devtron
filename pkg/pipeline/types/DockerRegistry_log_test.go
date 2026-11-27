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

package types

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// newRegistryConfigs returns a registry and its image-pull config with a synthetic
// marker in every credential-bearing field.
func newRegistryConfigs() (DockerArtifactStoreBean, DockerRegistryIpsConfigBean) {
	ips := DockerRegistryIpsConfigBean{Id: 3, CredentialType: "CUSTOM_CREDENTIAL", CredentialValue: "test-pull-credential", Active: true}
	registry := DockerArtifactStoreBean{
		Id:                     "test-registry",
		RegistryType:           "docker-hub",
		IsOCICompliantRegistry: true,
		IsDefault:              true,
		Active:                 true,
		RegistryURL:            "https://user:test-url-password@registry.example.com",
		Username:               "test-username",
		Password:               "test-password",
		AWSAccessKeyId:         "test-access-key",
		AWSSecretAccessKey:     "test-secret-key",
		Cert:                   "test-cert",
		OCIRegistryConfig:      map[string]string{"CHART": "test-oci-config"},
		DockerRegistryIpsConfig: &DockerRegistryIpsConfigBean{
			Id: 3, CredentialValue: "test-nested-pull-credential",
		},
	}
	return registry, ips
}

func TestRegistryConfigLogOmitsSecrets(t *testing.T) {
	secrets := []string{"test-pull-credential", "test-url-password", "test-username", "test-password",
		"test-access-key", "test-secret-key", "test-cert", "test-oci-config", "test-nested-pull-credential"}
	registry, ips := newRegistryConfigs()
	wantRegistry := map[string]interface{}{
		"id":                     "test-registry",
		"registryType":           "docker-hub",
		"isOCICompliantRegistry": true,
		"isPublic":               false,
		"isDefault":              true,
		"active":                 true,
	}
	wantIps := map[string]interface{}{"id": float64(3), "credentialType": "CUSTOM_CREDENTIAL", "active": true}

	// Handlers log these as values and services log them as pointers; both must be safe.
	tests := []struct {
		name    string
		payload interface{}
		want    map[string]interface{}
	}{
		{"registry value", registry, wantRegistry},
		{"registry pointer", &registry, wantRegistry},
		{"pull config value", ips, wantIps},
		{"pull config pointer", &ips, wantIps},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
			logger := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(&out), zapcore.DebugLevel)).Sugar()
			logger.Infow("request payload", "payload", tt.payload)
			line := out.String()

			for _, secret := range secrets {
				if strings.Contains(line, secret) {
					t.Errorf("log line contains secret %q: %s", secret, line)
				}
			}
			var entry struct {
				Payload map[string]interface{} `json:"payload"`
			}
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				t.Fatalf("log line is not valid JSON: %v", err)
			}
			if !reflect.DeepEqual(entry.Payload, tt.want) {
				t.Errorf("logged payload = %v, want %v", entry.Payload, tt.want)
			}
		})
	}
}

func TestRegistryConfigAPIJSONKeepsSecrets(t *testing.T) {
	// Hiding secrets from logs must not change the API: the UI still sends and receives these fields.
	registry, _ := newRegistryConfigs()
	raw, err := json.Marshal(registry)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DockerArtifactStoreBean
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, registry) {
		t.Errorf("JSON round trip changed the registry: got %+v, want %+v", decoded, registry)
	}
}
