/*
 * Copyright (c) 2024. Devtron Inc.
 */

package securestore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-pg/pg/types"
	"io"
)

var decryptionFailErr = fmt.Errorf("Decryption failed")

func DidDecryptionFail(err error) bool {
	return errors.Is(err, decryptionFailErr)
}

var encryptionKey []byte

func encrypt(data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func decrypt(cipherBase64 string) (string, error) {
	// Try decrypting (normal encrypted flow)
	cipherData, err := base64.StdEncoding.DecodeString(cipherBase64)
	if err == nil {
		block, err := aes.NewCipher(encryptionKey)
		if err != nil {
			return "", err
		}
		aesGCM, err := cipher.NewGCM(block)
		if err != nil {
			return "", err
		}
		nonceSize := aesGCM.NonceSize()
		if len(cipherData) >= nonceSize {
			nonce, ciphertext := cipherData[:nonceSize], cipherData[nonceSize:]
			plainText, err := aesGCM.Open(nil, nonce, ciphertext, nil)
			if err == nil {
				return string(plainText), nil // Successfully decrypted
			}
		}
	}
	return cipherBase64, decryptionFailErr
}

// AppendValue : can be used for auto encryption of fields but is only supported in go-pg/v10. Not being used anywhere as of now, to be tested when start using.
func (e EncryptedMap) AppendValue(b []byte, quote int) ([]byte, error) {
	jsonBytes, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	encryptedString, err := encrypt(jsonBytes)
	if err != nil {
		return nil, err
	}

	return types.AppendString(b, encryptedString, quote), nil
}
