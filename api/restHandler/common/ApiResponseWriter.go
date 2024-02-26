package common

import (
	"encoding/json"
	"github.com/devtron-labs/devtron/internal/util"
	"net/http"
)

type ApiResponse struct {
	Success bool           `json:"success,notnull" validate:"required"`
	Error   *ErrorResponse `json:"error,omitempty"`
	Result  interface{}    `json:"result,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code,notnull" validate:"required"`
	Message string `json:"message,notnull" validate:"required"`
}

func WriteApiJsonResponse(w http.ResponseWriter, result interface{}, statusCode int, errCode string, errorMessage string) {
	apiResponse := ApiResponse{}
	if len(errCode) == 0 {
		apiResponse.Success = true
		apiResponse.Result = result
	} else {
		apiResponse.Success = false
		apiResponse.Error = &ErrorResponse{
			Code: errCode,
		}
		if len(errorMessage) == 0 {
			apiResponse.Error.Message = ErrorMessage(errCode)
		} else {
			apiResponse.Error.Message = errorMessage
		}

	}
	WriteApiJsonResponseStructured(w, &apiResponse, statusCode)
}

func WriteApiJsonResponseStructured(w http.ResponseWriter, apiResponse *ApiResponse, statusCode int) {
	apiResponseByteArr, err := json.Marshal(&apiResponse)
	if err != nil {
		util.GetLogger().Error("error in marshaling api response object", err)
		statusCode = 500
	}
	w.Header().Set(CONTENT_TYPE, APPLICATION_JSON)
	w.WriteHeader(statusCode)
	_, err = w.Write(apiResponseByteArr)
	if err != nil {
		util.GetLogger().Error(err)
	}
}

func WriteOctetStreamResp(w http.ResponseWriter, r *http.Request, byteArr []byte, defaultFilename string) {
	w.Header().Set(CONTENT_TYPE, "application/octet-stream")
	if defaultFilename != "" {
		w.Header().Set(CONTENT_DISPOSITION, "attachment; filename="+defaultFilename)
	}
	w.Header().Set(CONTENT_LENGTH, r.Header.Get(CONTENT_LENGTH))
	w.WriteHeader(http.StatusOK)
	w.Write(byteArr)
}
