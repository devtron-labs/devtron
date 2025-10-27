package securestore

import (
	"encoding/json"
	"fmt"
)

type EncryptedMap map[string]string

const (
	ENCRYPTED_KEY = "encrypted_data"
)

func EncryptMap(m EncryptedMap) (map[string]string, error) {
	bytesM, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	encryptedData, err := encrypt(bytesM)
	if err != nil {
		return nil, err
	}
	return EncryptedMap{ENCRYPTED_KEY: encryptedData}, nil
}

func decryptMap(mCipher []byte) (map[string]string, error) {
	var m map[string]string
	err := json.Unmarshal(mCipher, &m)
	if err != nil {
		return nil, err
	}
	if len(m[ENCRYPTED_KEY]) > 0 {
		decryptedMapBytes, err := decrypt(m[ENCRYPTED_KEY])
		if err != nil && !DidDecryptionFail(err) {
			return nil, err
		}
		if err == nil {
			var decryptedMap map[string]string
			err = json.Unmarshal([]byte(decryptedMapBytes), &decryptedMap)
			if err != nil {
				return nil, err
			}
			return decryptedMap, nil
		}
	} else {
		// Fallback: maybe it's just raw JSON
		// Validate that it's a valid JSON map[string]string
		// Neither decrypted nor valid plaintext JSON
		return m, nil
	}
	return nil, fmt.Errorf("decryption failed and data is not valid plaintext JSON")
}

func (e *EncryptedMap) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var encrypted []byte
	switch v := value.(type) {
	case string:
		encrypted = []byte(v)
	case []byte:
		encrypted = v
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
	var err error
	*e, err = decryptMap(encrypted)
	if err != nil {
		return err
	}
	return nil
}
