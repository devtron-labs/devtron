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
	"github.com/devtron-labs/devtron/internal/sql/constants"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestGitOpsConfigLogOmitsSecrets(t *testing.T) {
	// Every credential-bearing column holds a synthetic marker that must never reach the log.
	secrets := []string{"test-username", "test-token", "test-host-password",
		"test-ca-cert", "test-tls-cert", "test-tls-key", "test-email"}
	model := GitOpsConfig{
		Id:       7,
		Provider: "GITLAB",
		Active:   true,
		AuthMode: constants.AUTH_MODE_USERNAME_PASSWORD,
		Username: "test-username",
		Token:    securestore.EncryptedString("test-token"),
		Host:     "https://user:test-host-password@gitlab.example.com",
		CaCert:   "test-ca-cert",
		TlsCert:  "test-tls-cert",
		TlsKey:   "test-tls-key",
		EmailId:  "test-email",
	}
	wantPayload := map[string]interface{}{
		"id":                    float64(7),
		"provider":              "GITLAB",
		"active":                true,
		"authMode":              "USERNAME_PASSWORD",
		"allowCustomRepository": false,
		"enableTLSVerification": false,
	}

	for name, payload := range map[string]interface{}{"value": model, "pointer": &model} {
		t.Run(name, func(t *testing.T) {
			var out bytes.Buffer
			encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
			logger := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(&out), zapcore.DebugLevel)).Sugar()
			logger.Errorw("error in saving gitops config", "data", payload)
			line := out.String()

			for _, secret := range secrets {
				if strings.Contains(line, secret) {
					t.Errorf("log line contains secret %q: %s", secret, line)
				}
			}
			var entry struct {
				Data map[string]interface{} `json:"data"`
			}
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				t.Fatalf("log line is not valid JSON: %v", err)
			}
			if !reflect.DeepEqual(entry.Data, wantPayload) {
				t.Errorf("logged model = %v, want %v", entry.Data, wantPayload)
			}
		})
	}
}
