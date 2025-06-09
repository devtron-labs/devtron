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
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	ENCRYPTED_SECRET_PREFIX = "***ENC:"
	ENCRYPTED_SECRET_SUFFIX = "***"
)

// SecretEncryption provides optimized symmetric encryption for secrets
type SecretEncryption interface {
	// EncryptSecret encrypts a secret value using the API token
	EncryptSecret(plaintext string, apiToken string) (string, error)

	// DecryptSecret decrypts an encrypted secret using the API token
	DecryptSecret(encryptedValue string, apiToken string) (string, error)

	// IsEncryptedSecret checks if a value is an encrypted secret
	IsEncryptedSecret(value string) bool

	// EncryptSecretBytes encrypts raw bytes (for large secrets)
	EncryptSecretBytes(plaintext []byte, apiToken string) ([]byte, error)

	// DecryptSecretBytes decrypts to raw bytes
	DecryptSecretBytes(ciphertext []byte, apiToken string) ([]byte, error)
}

type SecretEncryptionImpl struct {
	logger *zap.SugaredLogger
}

func NewSecretEncryptionImpl(logger *zap.SugaredLogger) *SecretEncryptionImpl {
	return &SecretEncryptionImpl{
		logger: logger,
	}
}

// deriveKey derives a 32-byte encryption key from the API token
func (impl *SecretEncryptionImpl) deriveKey(apiToken string) []byte {
	// Use SHA-256 to derive a consistent 32-byte key from API token
	hash := sha256.Sum256([]byte(apiToken))
	return hash[:]
}

func (impl *SecretEncryptionImpl) EncryptSecret(plaintext string, apiToken string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Derive encryption key from API token
	key := impl.deriveKey(apiToken)

	// Create ChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		impl.logger.Errorw("failed to create cipher", "err", err)
		return "", err
	}

	// Generate random nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		impl.logger.Errorw("failed to generate nonce", "err", err)
		return "", err
	}

	// Encrypt the plaintext
	ciphertext := aead.Seal(nil, nonce, []byte(plaintext), nil)

	// Combine nonce + ciphertext
	encrypted := append(nonce, ciphertext...)

	// Encode to base64 and wrap with markers
	encoded := base64.StdEncoding.EncodeToString(encrypted)
	return ENCRYPTED_SECRET_PREFIX + encoded + ENCRYPTED_SECRET_SUFFIX, nil
}

func (impl *SecretEncryptionImpl) DecryptSecret(encryptedValue string, apiToken string) (string, error) {
	if !impl.IsEncryptedSecret(encryptedValue) {
		// Not an encrypted value, return as-is
		return encryptedValue, nil
	}

	// Extract base64 content
	prefixLen := len(ENCRYPTED_SECRET_PREFIX)
	suffixLen := len(ENCRYPTED_SECRET_SUFFIX)
	if len(encryptedValue) <= prefixLen+suffixLen {
		return "", errors.New("invalid encrypted secret format")
	}

	encoded := encryptedValue[prefixLen : len(encryptedValue)-suffixLen]

	// Decode from base64
	encrypted, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		impl.logger.Errorw("failed to decode encrypted secret", "err", err)
		return "", err
	}

	// Derive encryption key from API token
	key := impl.deriveKey(apiToken)

	// Create ChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		impl.logger.Errorw("failed to create cipher for decryption", "err", err)
		return "", err
	}

	// Check minimum length (nonce + ciphertext + auth tag)
	nonceSize := aead.NonceSize()
	if len(encrypted) < nonceSize {
		return "", errors.New("encrypted data too short")
	}

	// Split nonce and ciphertext
	nonce := encrypted[:nonceSize]
	ciphertext := encrypted[nonceSize:]

	// Decrypt
	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		impl.logger.Errorw("failed to decrypt secret", "err", err)
		return "", err
	}

	return string(plaintext), nil
}

func (impl *SecretEncryptionImpl) IsEncryptedSecret(value string) bool {
	return len(value) > len(ENCRYPTED_SECRET_PREFIX)+len(ENCRYPTED_SECRET_SUFFIX) &&
		value[:len(ENCRYPTED_SECRET_PREFIX)] == ENCRYPTED_SECRET_PREFIX &&
		value[len(value)-len(ENCRYPTED_SECRET_SUFFIX):] == ENCRYPTED_SECRET_SUFFIX
}

func (impl *SecretEncryptionImpl) EncryptSecretBytes(plaintext []byte, apiToken string) ([]byte, error) {
	if len(plaintext) == 0 {
		return nil, nil
	}

	// Derive encryption key from API token
	key := impl.deriveKey(apiToken)

	// Create ChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Generate random nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	// Encrypt the plaintext
	ciphertext := aead.Seal(nil, nonce, plaintext, nil)

	// Combine nonce + ciphertext
	encrypted := append(nonce, ciphertext...)

	return encrypted, nil
}

func (impl *SecretEncryptionImpl) DecryptSecretBytes(ciphertext []byte, apiToken string) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, nil
	}

	// Derive encryption key from API token
	key := impl.deriveKey(apiToken)

	// Create ChaCha20-Poly1305 cipher
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}

	// Check minimum length
	nonceSize := aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	// Split nonce and ciphertext
	nonce := ciphertext[:nonceSize]
	encrypted := ciphertext[nonceSize:]

	// Decrypt
	plaintext, err := aead.Open(nil, nonce, encrypted, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// Utility functions for performance optimization

// EncryptSecretFast encrypts without base64 encoding (for internal storage)
func (impl *SecretEncryptionImpl) EncryptSecretFast(plaintext string, apiToken string) ([]byte, error) {
	return impl.EncryptSecretBytes([]byte(plaintext), apiToken)
}

// DecryptSecretFast decrypts from raw bytes
func (impl *SecretEncryptionImpl) DecryptSecretFast(ciphertext []byte, apiToken string) (string, error) {
	plaintext, err := impl.DecryptSecretBytes(ciphertext, apiToken)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// BatchEncryptSecrets encrypts multiple secrets efficiently
func (impl *SecretEncryptionImpl) BatchEncryptSecrets(secrets map[string]string, apiToken string) (map[string]string, error) {
	encrypted := make(map[string]string)

	for key, value := range secrets {
		if value == "" {
			encrypted[key] = ""
			continue
		}

		encryptedValue, err := impl.EncryptSecret(value, apiToken)
		if err != nil {
			impl.logger.Errorw("failed to encrypt secret in batch", "key", key, "err", err)
			return nil, err
		}
		encrypted[key] = encryptedValue
	}

	return encrypted, nil
}

// BatchDecryptSecrets decrypts multiple secrets efficiently
func (impl *SecretEncryptionImpl) BatchDecryptSecrets(encryptedSecrets map[string]string, apiToken string) (map[string]string, error) {
	decrypted := make(map[string]string)

	for key, value := range encryptedSecrets {
		if value == "" {
			decrypted[key] = ""
			continue
		}

		decryptedValue, err := impl.DecryptSecret(value, apiToken)
		if err != nil {
			impl.logger.Errorw("failed to decrypt secret in batch", "key", key, "err", err)
			return nil, err
		}
		decrypted[key] = decryptedValue
	}

	return decrypted, nil
}

/*
Performance Characteristics:

1. **ChaCha20-Poly1305**:
   - Encryption: ~1-2 GB/s
   - Decryption: ~1-2 GB/s
   - Overhead: 16 bytes (nonce) + 16 bytes (auth tag) = 32 bytes
   - Key derivation: ~1 microsecond (SHA-256)

2. **Storage Efficiency**:
   - Original secret: N bytes
   - Encrypted: N + 32 bytes
   - Base64 encoded: (N + 32) * 4/3 bytes
   - With markers: (N + 32) * 4/3 + 10 bytes

3. **Example Storage Sizes**:
   - 20-byte secret → 69 bytes encrypted (3.45x)
   - 50-byte secret → 119 bytes encrypted (2.38x)
   - 100-byte secret → 186 bytes encrypted (1.86x)

Usage Example:

encryptor := NewSecretEncryptionImpl(logger)
apiToken := "your-api-token-from-attributes-table"

// Encrypt
encrypted, err := encryptor.EncryptSecret("my-secret-password", apiToken)
// Result: "***ENC:SGVsbG8gV29ybGQhIFRoaXMgaXMgYSB0ZXN0***"

// Decrypt
decrypted, err := encryptor.DecryptSecret(encrypted, apiToken)
// Result: "my-secret-password"

// Check if encrypted
isEnc := encryptor.IsEncryptedSecret(encrypted) // true
*/
