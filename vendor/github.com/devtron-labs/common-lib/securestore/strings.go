package securestore

import (
	"fmt"
)

type EncryptedString string

func (e *EncryptedString) String() string {
	return string(*e)
}

func ToEncryptedString(s string) EncryptedString {
	return EncryptedString(s)
}

func EncryptString(data string) (EncryptedString, error) {
	encryptedStr, err := encrypt([]byte(data))
	if err != nil {
		return "", err
	}
	return EncryptedString(encryptedStr), nil
}

func decryptString(cipherBase64 string) (string, error) {
	decryptedBytes, err := decrypt(cipherBase64)
	if err != nil && !DidDecryptionFail(err) {
		return "", err
	}
	// Fallback: decrypting failed, considering it as just normal string
	if DidDecryptionFail(err) {
		return cipherBase64, nil
	} else {
		return decryptedBytes, nil
	}
}

func (e *EncryptedString) Scan(value interface{}) error {
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
	decryptedBytes, err := decryptString(string(encrypted))
	if err != nil {
		return err
	}
	if decryptedBytes == "" {
		return nil
	}
	*e = EncryptedString(decryptedBytes)
	return nil
}
