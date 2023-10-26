package util

import (
	"encoding/base64"
	"encoding/json"
)

type SecretTransformMode int

const (
	EncodeSecret SecretTransformMode = 1
	DecodeSecret SecretTransformMode = 2
)

func GetDecodedAndEncodedData(data json.RawMessage, transformer SecretTransformMode) ([]byte, error) {
	dataMap := make(map[string]string)
	err := json.Unmarshal(data, &dataMap)
	if err != nil {
		return nil, err
	}
	var transformedData []byte
	for key, value := range dataMap {
		switch transformer {
		case EncodeSecret:
			transformedData = []byte(base64.StdEncoding.EncodeToString([]byte(value)))
		case DecodeSecret:
			transformedData, err = base64.StdEncoding.DecodeString(value)
			if err != nil {
				return nil, err
			}
		}
		dataMap[key] = string(transformedData)
	}
	marshal, err := json.Marshal(dataMap)
	if err != nil {
		return nil, err
	}
	return marshal, nil
}
