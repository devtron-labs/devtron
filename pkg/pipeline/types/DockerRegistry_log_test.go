package types

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestRegistryConfigurationLogsExcludeCredentials(t *testing.T) {
	ips := DockerRegistryIpsConfigBean{Id: 7, CredentialValue: "test-pull-credential", Active: true}
	registry := DockerArtifactStoreBean{
		Id: "example-registry", RegistryType: "other", RegistryURL: "https://test-user:test-url-password@registry.example",
		Username: "test-registry-user", Password: "test-registry-password",
		AWSAccessKeyId: "test-access-key", AWSSecretAccessKey: "test-secret-key",
		Cert: "test-certificate-data", OCIRegistryConfig: map[string]string{"password": "test-oci-password"},
		DockerRegistryIpsConfig: &ips, Active: true,
	}
	for name, payload := range map[string]interface{}{
		"registry value": registry, "registry pointer": &registry,
		"pull configuration value": ips, "pull configuration pointer": &ips,
	} {
		t.Run(name, func(t *testing.T) {
			var output bytes.Buffer
			logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel)).Sugar()
			logger.Infow("configuration request", "payload", payload)
			for _, secret := range []string{
				"test-pull-credential", "test-registry-user", "test-registry-password", "test-access-key",
				"test-secret-key", "test-certificate-data", "test-oci-password", "test-url-password",
			} {
				if strings.Contains(output.String(), secret) {
					t.Fatal("structured log contains a credential field")
				}
			}
			var entry map[string]interface{}
			if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
				t.Fatal(err)
			}
			metadata, ok := entry["payload"].(map[string]interface{})
			if !ok || metadata["active"] != true || metadata["id"] == nil {
				t.Fatal("structured log lost configuration identity metadata")
			}
		})
	}
}

func TestRegistryLogRedactionPreservesAPIJSON(t *testing.T) {
	registry := DockerArtifactStoreBean{Id: "example-registry", Password: "test-password", DockerRegistryIpsConfig: &DockerRegistryIpsConfigBean{CredentialValue: "test-pull-credential"}}
	raw, err := json.Marshal(registry)
	if err != nil {
		t.Fatal(err)
	}
	var decoded DockerArtifactStoreBean
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Password != registry.Password || decoded.DockerRegistryIpsConfig.CredentialValue != registry.DockerRegistryIpsConfig.CredentialValue {
		t.Fatal("log redaction changed registry API serialization")
	}
}
