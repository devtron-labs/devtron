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

package gitOps

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/devtron-labs/devtron/api/bean"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// logAsJSON logs payload the way the GitOps handlers do and returns the JSON line zap wrote.
func logAsJSON(t *testing.T, payload interface{}) string {
	t.Helper()
	var out bytes.Buffer
	encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	logger := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(&out), zapcore.DebugLevel)).Sugar()
	logger.Infow("request payload", "payload", payload)
	return out.String()
}

func TestGitOpsConfigDtoLogOmitsSecrets(t *testing.T) {
	// Every credential-bearing field holds a synthetic marker that must never reach the log.
	secrets := []string{"test-username", "test-token", "test-host-password",
		"test-ca-data", "test-tls-cert", "test-tls-key", "test-user-email"}
	dto := GitOpsConfigDto{
		Id:          7,
		Provider:    "GITLAB",
		Active:      true,
		AuthMode:    PASSWORD,
		Username:    "test-username",
		Token:       "test-token",
		Host:        "https://user:test-host-password@gitlab.example.com",
		TLSConfig:   &bean.TLSConfig{CaData: "test-ca-data", TLSCertData: "test-tls-cert", TLSKeyData: "test-tls-key"},
		UserEmailId: "test-user-email",
	}
	wantPayload := map[string]interface{}{
		"id":                    float64(7),
		"provider":              "GITLAB",
		"active":                true,
		"authMode":              "PASSWORD",
		"allowCustomRepository": false,
		"enableTLSVerification": false,
	}

	// Handlers log the DTO as a value and services log it as a pointer; both must be safe.
	for name, payload := range map[string]interface{}{"value": dto, "pointer": &dto} {
		t.Run(name, func(t *testing.T) {
			line := logAsJSON(t, payload)
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
			if !reflect.DeepEqual(entry.Payload, wantPayload) {
				t.Errorf("logged payload = %v, want %v", entry.Payload, wantPayload)
			}
		})
	}
}

func TestGitOpsConfigDtoAPIJSONKeepsSecrets(t *testing.T) {
	// Hiding secrets from logs must not change the API: the UI still sends and receives these fields.
	dto := GitOpsConfigDto{
		Token:     "test-token",
		TLSConfig: &bean.TLSConfig{CaData: "test-ca-data", TLSCertData: "test-tls-cert", TLSKeyData: "test-tls-key"},
	}
	raw, err := json.Marshal(dto)
	if err != nil {
		t.Fatal(err)
	}
	var decoded GitOpsConfigDto
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, dto) {
		t.Errorf("JSON round trip changed the DTO: got %+v, want %+v", decoded, dto)
	}
}
