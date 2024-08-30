package utils

import "encoding/json"

func ConvertToJsonRawMessage(request interface{}) (json.RawMessage, error) {
	var r json.RawMessage
	configMapByte, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}
	err = r.UnmarshalJSON(configMapByte)
	if err != nil {
		return nil, err
	}
	return r, nil
}
