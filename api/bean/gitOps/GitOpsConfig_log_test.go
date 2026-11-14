package gitOps_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/devtron-labs/common-lib/securestore"
	"github.com/devtron-labs/devtron/api/bean"
	gitOps "github.com/devtron-labs/devtron/api/bean/gitOps"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestGitOpsConfigurationLogsExcludeCredentials(t *testing.T) {
	dto := gitOps.GitOpsConfigDto{Id: 3, Provider: "GITHUB", Active: true,
		Username: "test-git-user", Token: "test-git-token", Host: "https://user:test-host-password@git.example",
		TLSConfig: &bean.TLSConfig{TLSKeyData: "test-tls-key", TLSCertData: "test-tls-cert", CaData: "test-ca-data"}}
	model := repository.GitOpsConfig{Id: 3, Provider: "GITHUB", Active: true,
		Username: dto.Username, Token: securestore.ToEncryptedString(dto.Token), Host: dto.Host,
		TlsKey: dto.TLSConfig.TLSKeyData, TlsCert: dto.TLSConfig.TLSCertData, CaCert: dto.TLSConfig.CaData}
	for name, payload := range map[string]interface{}{
		"request value": dto, "request pointer": &dto,
		"model value": model, "model pointer": &model,
	} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel)).Sugar()
			logger.Infow("GitOps configuration", "payload", payload)
			for _, secret := range []string{"test-git-user", "test-git-token", "test-host-password", "test-tls-key", "test-tls-cert", "test-ca-data"} {
				if strings.Contains(output.String(), secret) {
					t.Fatal("structured log contains a credential field")
				}
			}
			var entry map[string]interface{}
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatal(err)
			}
			metadata, ok := entry["payload"].(map[string]interface{})
			if !ok || metadata["id"] != float64(3) || metadata["provider"] != "GITHUB" || metadata["active"] != true {
				t.Fatal("structured log lost configuration identity metadata")
			}
		})
	}
}

func TestGitOpsLogRedactionPreservesAPIJSON(t *testing.T) {
	want := gitOps.GitOpsConfigDto{Token: "test-git-token", TLSConfig: &bean.TLSConfig{TLSKeyData: "test-tls-key"}}
	raw, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	var got gitOps.GitOpsConfigDto
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Token != want.Token || got.TLSConfig.TLSKeyData != want.TLSConfig.TLSKeyData {
		t.Fatal("log redaction changed GitOps API serialization")
	}
}
