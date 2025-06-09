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

package helper

import (
	"encoding/json"
	"go.uber.org/zap"
	"strings"
)

const (
	SANITIZED_SECRET_PLACEHOLDER = "***SANITIZED***"
	SECRET_REFERENCE_PREFIX      = "***SECRET_REF:"
	SECRET_REFERENCE_SUFFIX      = "***"
)

// SecretSanitizer handles sanitization and reconstruction of secrets in WorkflowRequest
type SecretSanitizer interface {
	// EncryptWorkflowRequest encrypts secret values using API token
	EncryptWorkflowRequest(workflowRequest interface{}, apiToken string) (interface{}, error)

	// DecryptWorkflowRequest decrypts secret values using API token
	DecryptWorkflowRequest(encryptedWorkflowRequest interface{}, apiToken string) (interface{}, error)

	// EncryptJSON encrypts secrets in JSON string
	EncryptJSON(jsonStr string, apiToken string) (string, error)

	// DecryptJSON decrypts secrets in JSON string
	DecryptJSON(encryptedJsonStr string, apiToken string) (string, error)
}

type SecretSanitizerImpl struct {
	logger           *zap.SugaredLogger
	secretEncryption SecretEncryption
	// Add dependencies for fetching current secret values
	// configMapService ConfigMapService
	// secretService    SecretService
}

func NewSecretSanitizerImpl(logger *zap.SugaredLogger) *SecretSanitizerImpl {
	return &SecretSanitizerImpl{
		logger:           logger,
		secretEncryption: NewSecretEncryptionImpl(logger),
	}
}

// SecretReference stores metadata about sanitized secrets
type SecretReference struct {
	Type          string `json:"type"`          // "configmap", "secret", "env_var"
	Name          string `json:"name"`          // name of the secret/configmap
	Key           string `json:"key"`           // key within the secret/configmap
	Namespace     string `json:"namespace"`     // namespace (for k8s secrets)
	AppId         int    `json:"appId"`         // app context
	EnvironmentId *int   `json:"environmentId"` // environment context
}

func (impl *SecretSanitizerImpl) EncryptWorkflowRequest(workflowRequest interface{}, apiToken string) (interface{}, error) {
	impl.logger.Infow("encrypting secrets in workflow request")

	// Convert to JSON for easier manipulation
	jsonBytes, err := json.Marshal(workflowRequest)
	if err != nil {
		impl.logger.Errorw("error marshaling workflow request for encryption", "err", err)
		return nil, err
	}

	// Parse as generic map for manipulation
	var workflowMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &workflowMap)
	if err != nil {
		impl.logger.Errorw("error unmarshaling workflow request for encryption", "err", err)
		return nil, err
	}

	// Encrypt different types of secrets
	encryptedMap := impl.encryptMap(workflowMap, apiToken)

	impl.logger.Infow("successfully encrypted secrets in workflow request")
	return encryptedMap, nil
}

func (impl *SecretSanitizerImpl) encryptMap(data map[string]interface{}, apiToken string) map[string]interface{} {
	return impl.encryptMapWithPath(data, "", apiToken)
}

func (impl *SecretSanitizerImpl) encryptMapWithPath(data map[string]interface{}, parentPath string, apiToken string) map[string]interface{} {
	encrypted := make(map[string]interface{})

	for key, value := range data {
		// Build full field path for nested field detection
		fullPath := key
		if parentPath != "" {
			fullPath = parentPath + "." + key
		}

		encrypted[key] = impl.encryptValueWithPath(fullPath, value, apiToken)
	}

	return encrypted
}

func (impl *SecretSanitizerImpl) encryptValueWithPath(fullPath string, value interface{}, apiToken string) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		return impl.encryptStringValueWithPath(fullPath, v, apiToken)
	case map[string]interface{}:
		return impl.encryptMapWithPath(v, fullPath, apiToken)
	case []interface{}:
		return impl.encryptArrayWithPath(v, fullPath, apiToken)
	default:
		return value
	}
}

func (impl *SecretSanitizerImpl) encryptArrayWithPath(arr []interface{}, parentPath string, apiToken string) []interface{} {
	encrypted := make([]interface{}, len(arr))
	for i, item := range arr {
		if itemMap, ok := item.(map[string]interface{}); ok {
			encrypted[i] = impl.encryptMapWithPath(itemMap, parentPath, apiToken)
		} else {
			encrypted[i] = item
		}
	}
	return encrypted
}

func (impl *SecretSanitizerImpl) encryptStringValueWithPath(fullPath string, value string, apiToken string) interface{} {
	// Check if this is a known secret field
	if impl.isKnownSecretField(fullPath) {
		impl.logger.Debugw("encrypting secret field", "field", fullPath, "originalLength", len(value))
		encrypted, err := impl.secretEncryption.EncryptSecret(value, apiToken)
		if err != nil {
			impl.logger.Errorw("failed to encrypt secret field", "field", fullPath, "err", err)
			return SANITIZED_SECRET_PLACEHOLDER // Fallback to sanitization
		}
		return encrypted
	}

	// Special handling for runtime parameters
	if impl.isRuntimeParameterSecret(fullPath, value) {
		impl.logger.Debugw("encrypting runtime parameter secret", "field", fullPath, "originalLength", len(value))
		encrypted, err := impl.secretEncryption.EncryptSecret(value, apiToken)
		if err != nil {
			impl.logger.Errorw("failed to encrypt runtime parameter secret", "field", fullPath, "err", err)
			return SANITIZED_SECRET_PLACEHOLDER // Fallback to sanitization
		}
		return encrypted
	}

	return value
}

func (impl *SecretSanitizerImpl) isKnownSecretField(fieldName string) bool {
	// Exact field names that should be sanitized
	knownSecretFields := map[string]bool{
		// Docker registry secrets
		"dockerPassword":  true,
		"docker_password": true,

		// AWS/S3 secrets
		"accessKey":  true,
		"access_key": true,
		"secretKey":  true,
		"secret_key": true,
		"passkey":    true,
		"pass_key":   true,

		// Docker certificates
		"dockerCert":         true,
		"docker_cert":        true,
		"dockerCertificate":  true,
		"docker_certificate": true,

		// Azure blob storage
		"accountKey":  true,
		"account_key": true,

		// GCP blob storage
		"credentialFileData":   true,
		"credential_file_data": true,
		"serviceAccountKey":    true,
		"service_account_key":  true,

		// Additional common secret fields
		"password":      true,
		"token":         true,
		"apiKey":        true,
		"api_key":       true,
		"privateKey":    true,
		"private_key":   true,
		"clientSecret":  true,
		"client_secret": true,
	}

	// Check exact field name
	if knownSecretFields[fieldName] {
		return true
	}

	// Check for nested field paths (e.g., "blobStorageS3Config.accessKey")
	return impl.isNestedSecretField(fieldName)
}

func (impl *SecretSanitizerImpl) isNestedSecretField(fieldPath string) bool {
	// Define nested secret field patterns
	nestedSecretPatterns := []string{
		// Blob storage configurations
		"blobStorageS3Config.accessKey",
		"blobStorageS3Config.secretKey",
		"blobStorageS3Config.passkey",
		"blobStorageS3Config.password",

		"azureBlobConfig.accountKey",
		"azureBlobConfig.accountName",
		"azureBlobConfig.password",

		"gcpBlobConfig.credentialFileData",
		"gcpBlobConfig.serviceAccountKey",
		"gcpBlobConfig.privateKey",

		// Docker registry configurations
		"dockerRegistry.password",
		"dockerRegistry.dockerPassword",
		"dockerRegistry.accessKey",
		"dockerRegistry.secretKey",
		"dockerRegistry.cert",
		"dockerRegistry.certificate",

		// Database configurations
		"database.password",
		"database.connectionString",
		"database.url",

		// External service configurations
		"externalService.apiKey",
		"externalService.token",
		"externalService.secret",
		"externalService.password",

		// CI/CD tool configurations
		"gitCredentials.password",
		"gitCredentials.token",
		"gitCredentials.privateKey",

		// Cloud provider configurations
		"awsConfig.secretAccessKey",
		"awsConfig.accessKeyId",
		"gcpConfig.serviceAccountKey",
		"azureConfig.clientSecret",
		"azureConfig.password",
	}

	// Check if the field path matches any known nested secret pattern
	for _, pattern := range nestedSecretPatterns {
		if strings.HasSuffix(fieldPath, pattern) || fieldPath == pattern {
			return true
		}
	}

	// Check if any part of the path contains secret field names
	pathParts := strings.Split(fieldPath, ".")
	if len(pathParts) > 1 {
		lastPart := pathParts[len(pathParts)-1]
		return impl.isKnownSecretField(lastPart)
	}

	return false
}

func (impl *SecretSanitizerImpl) isRuntimeParameterSecret(fieldPath string, value string) bool {
	// Check if this is within runtime parameters
	if !strings.Contains(fieldPath, "runtimeParameters") &&
		!strings.Contains(fieldPath, "runtime_parameters") &&
		!strings.Contains(fieldPath, "RuntimeParameters") {
		return false
	}

	// Simple heuristics for runtime parameters
	return impl.looksLikeSecret(value)
}

func (impl *SecretSanitizerImpl) looksLikeSecret(value string) bool {
	// Skip short values
	if len(value) < 12 {
		return false
	}

	// Check if field name suggests it's a secret
	lowerValue := strings.ToLower(value)
	if strings.Contains(lowerValue, "password") ||
		strings.Contains(lowerValue, "secret") ||
		strings.Contains(lowerValue, "token") ||
		strings.Contains(lowerValue, "key") {
		return false // These are likely field names, not values
	}

	// Simple patterns that are very likely secrets
	if strings.HasPrefix(value, "sk-") || // OpenAI, Stripe keys
		strings.HasPrefix(value, "ghp_") || // GitHub tokens
		strings.HasPrefix(value, "xoxb-") || // Slack tokens
		strings.HasPrefix(value, "AKIA") || // AWS keys
		strings.Contains(value, "-----BEGIN") { // Certificates
		return true
	}

	// If it's long and looks random (no spaces, mixed case/numbers)
	if len(value) > 20 && !strings.Contains(value, " ") {
		hasUpper := false
		hasLower := false
		hasDigit := false

		for _, char := range value {
			if char >= 'A' && char <= 'Z' {
				hasUpper = true
			} else if char >= 'a' && char <= 'z' {
				hasLower = true
			} else if char >= '0' && char <= '9' {
				hasDigit = true
			}
		}

		// If it has mixed case and numbers, likely a secret
		return hasUpper && hasLower && hasDigit
	}

	return false
}

func (impl *SecretSanitizerImpl) createSecretReference(key string, value string) *SecretReference {
	// Try to extract secret reference information from context
	// This would need to be enhanced based on your WorkflowRequest structure

	// For now, return a basic reference
	return &SecretReference{
		Type: "unknown",
		Name: key,
		Key:  key,
	}
}

func (impl *SecretSanitizerImpl) DecryptWorkflowRequest(encryptedWorkflowRequest interface{}, apiToken string) (interface{}, error) {
	impl.logger.Infow("decrypting secrets in workflow request")

	// Convert to JSON for manipulation
	jsonBytes, err := json.Marshal(encryptedWorkflowRequest)
	if err != nil {
		return nil, err
	}

	// Parse as generic map
	var workflowMap map[string]interface{}
	err = json.Unmarshal(jsonBytes, &workflowMap)
	if err != nil {
		return nil, err
	}

	// Decrypt secrets
	decryptedMap := impl.decryptMap(workflowMap, apiToken)

	impl.logger.Infow("successfully decrypted secrets in workflow request")
	return decryptedMap, nil
}

func (impl *SecretSanitizerImpl) decryptMap(data map[string]interface{}, apiToken string) map[string]interface{} {
	return impl.decryptMapWithPath(data, "", apiToken)
}

func (impl *SecretSanitizerImpl) decryptMapWithPath(data map[string]interface{}, parentPath string, apiToken string) map[string]interface{} {
	decrypted := make(map[string]interface{})

	for key, value := range data {
		// Build full field path for nested field detection
		fullPath := key
		if parentPath != "" {
			fullPath = parentPath + "." + key
		}

		decrypted[key] = impl.decryptValueWithPath(fullPath, value, apiToken)
	}

	return decrypted
}

func (impl *SecretSanitizerImpl) decryptValueWithPath(fullPath string, value interface{}, apiToken string) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		return impl.decryptStringValueWithPath(fullPath, v, apiToken)
	case map[string]interface{}:
		return impl.decryptMapWithPath(v, fullPath, apiToken)
	case []interface{}:
		return impl.decryptArrayWithPath(v, fullPath, apiToken)
	default:
		return value
	}
}

func (impl *SecretSanitizerImpl) decryptArrayWithPath(arr []interface{}, parentPath string, apiToken string) []interface{} {
	decrypted := make([]interface{}, len(arr))
	for i, item := range arr {
		decrypted[i] = impl.decryptValueWithPath(parentPath, item, apiToken)
	}
	return decrypted
}

func (impl *SecretSanitizerImpl) decryptStringValueWithPath(fullPath string, value string, apiToken string) interface{} {
	// Check if it's an encrypted secret
	if impl.secretEncryption.IsEncryptedSecret(value) {
		impl.logger.Debugw("decrypting secret field", "field", fullPath)
		decrypted, err := impl.secretEncryption.DecryptSecret(value, apiToken)
		if err != nil {
			impl.logger.Errorw("failed to decrypt secret", "field", fullPath, "err", err)
			return "" // Return empty string on decryption failure
		}
		return decrypted
	}

	// Check if it's a sanitized placeholder (legacy support)
	if value == SANITIZED_SECRET_PLACEHOLDER {
		impl.logger.Debugw("found sanitized placeholder, returning empty", "field", fullPath)
		return ""
	}

	return value
}

func (impl *SecretSanitizerImpl) EncryptJSON(jsonStr string, apiToken string) (string, error) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	encrypted, err := impl.EncryptWorkflowRequest(data, apiToken)
	if err != nil {
		return "", err
	}

	encryptedBytes, err := json.Marshal(encrypted)
	if err != nil {
		return "", err
	}

	return string(encryptedBytes), nil
}

func (impl *SecretSanitizerImpl) DecryptJSON(encryptedJsonStr string, apiToken string) (string, error) {
	var data interface{}
	err := json.Unmarshal([]byte(encryptedJsonStr), &data)
	if err != nil {
		return "", err
	}

	decrypted, err := impl.DecryptWorkflowRequest(data, apiToken)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := json.Marshal(decrypted)
	if err != nil {
		return "", err
	}

	return string(decryptedBytes), nil
}

func (impl *SecretSanitizerImpl) reconstructMap(data map[string]interface{}, environmentId *int, appId int) map[string]interface{} {
	return impl.reconstructMapWithPath(data, "", environmentId, appId)
}

func (impl *SecretSanitizerImpl) reconstructMapWithPath(data map[string]interface{}, parentPath string, environmentId *int, appId int) map[string]interface{} {
	reconstructed := make(map[string]interface{})

	for key, value := range data {
		// Build full field path for nested field detection
		fullPath := key
		if parentPath != "" {
			fullPath = parentPath + "." + key
		}

		reconstructed[key] = impl.reconstructValueWithPath(fullPath, value, environmentId, appId)
	}

	return reconstructed
}

func (impl *SecretSanitizerImpl) reconstructValueWithPath(fullPath string, value interface{}, environmentId *int, appId int) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		return impl.reconstructStringValueWithPath(fullPath, v, environmentId, appId)
	case map[string]interface{}:
		return impl.reconstructMapWithPath(v, fullPath, environmentId, appId)
	case []interface{}:
		return impl.reconstructArrayWithPath(v, fullPath, environmentId, appId)
	default:
		return value
	}
}

func (impl *SecretSanitizerImpl) reconstructArrayWithPath(arr []interface{}, parentPath string, environmentId *int, appId int) []interface{} {
	reconstructed := make([]interface{}, len(arr))
	for i, item := range arr {
		reconstructed[i] = impl.reconstructValueWithPath(parentPath, item, environmentId, appId)
	}
	return reconstructed
}

func (impl *SecretSanitizerImpl) reconstructStringValueWithPath(fullPath string, value string, environmentId *int, appId int) interface{} {
	// Check if it's a sanitized placeholder
	if value == SANITIZED_SECRET_PLACEHOLDER {
		// For field-specific approach, we know this was a secret field
		// Return empty string as we don't store references for reconstruction
		impl.logger.Debugw("reconstructing sanitized secret field", "field", fullPath)
		return ""
	}

	// Check if it's a secret reference (legacy support)
	if strings.HasPrefix(value, SECRET_REFERENCE_PREFIX) && strings.HasSuffix(value, SECRET_REFERENCE_SUFFIX) {
		refJson := value[len(SECRET_REFERENCE_PREFIX) : len(value)-len(SECRET_REFERENCE_SUFFIX)]

		var secretRef SecretReference
		err := json.Unmarshal([]byte(refJson), &secretRef)
		if err != nil {
			impl.logger.Errorw("error unmarshaling secret reference", "err", err)
			return ""
		}

		// Fetch current secret value
		currentValue := impl.fetchCurrentSecretValue(&secretRef, environmentId, appId)
		return currentValue
	}

	return value
}

func (impl *SecretSanitizerImpl) fetchCurrentSecretValue(secretRef *SecretReference, environmentId *int, appId int) string {
	// This method would integrate with your secret management system
	// to fetch current values of secrets/configmaps

	impl.logger.Infow("fetching current secret value",
		"type", secretRef.Type,
		"name", secretRef.Name,
		"key", secretRef.Key,
		"appId", appId,
		"environmentId", environmentId)

	// TODO: Implement actual secret fetching logic
	// Examples:
	// - Fetch from ConfigMap service
	// - Fetch from Secret service
	// - Fetch from environment variables
	// - Fetch from external secret management (Vault, etc.)

	switch secretRef.Type {
	case "configmap":
		return impl.fetchConfigMapValue(secretRef, environmentId, appId)
	case "secret":
		return impl.fetchSecretValue(secretRef, environmentId, appId)
	case "env_var":
		return impl.fetchEnvironmentVariable(secretRef, environmentId, appId)
	default:
		impl.logger.Warnw("unknown secret type", "type", secretRef.Type)
		return ""
	}
}

func (impl *SecretSanitizerImpl) fetchConfigMapValue(secretRef *SecretReference, environmentId *int, appId int) string {
	// TODO: Integrate with ConfigMapService to fetch current value
	// return impl.configMapService.GetConfigMapValue(appId, environmentId, secretRef.Name, secretRef.Key)
	return ""
}

func (impl *SecretSanitizerImpl) fetchSecretValue(secretRef *SecretReference, environmentId *int, appId int) string {
	// TODO: Integrate with SecretService to fetch current value
	// return impl.secretService.GetSecretValue(appId, environmentId, secretRef.Name, secretRef.Key)
	return ""
}

func (impl *SecretSanitizerImpl) fetchEnvironmentVariable(secretRef *SecretReference, environmentId *int, appId int) string {
	// TODO: Integrate with environment variable service
	// return impl.envVarService.GetEnvironmentVariable(appId, environmentId, secretRef.Key)
	return ""
}

func (impl *SecretSanitizerImpl) SanitizeJSON(jsonStr string) (string, error) {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", err
	}

	sanitized, err := impl.SanitizeWorkflowRequest(data)
	if err != nil {
		return "", err
	}

	sanitizedBytes, err := json.Marshal(sanitized)
	if err != nil {
		return "", err
	}

	return string(sanitizedBytes), nil
}

func (impl *SecretSanitizerImpl) ReconstructJSON(sanitizedJsonStr string, environmentId *int, appId int) (string, error) {
	var data interface{}
	err := json.Unmarshal([]byte(sanitizedJsonStr), &data)
	if err != nil {
		return "", err
	}

	reconstructed, err := impl.ReconstructSecrets(data, environmentId, appId)
	if err != nil {
		return "", err
	}

	reconstructedBytes, err := json.Marshal(reconstructed)
	if err != nil {
		return "", err
	}

	return string(reconstructedBytes), nil
}
