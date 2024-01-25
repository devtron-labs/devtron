package providers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/devtron-labs/common-lib/cloud-provider-identifier/bean"
	"go.uber.org/zap"
)

type oracleMetadataResponse struct {
	OkeTM string `json:"oke-tm"`
}

type IdentifyOracle struct {
	Logger *zap.SugaredLogger
}

func (impl *IdentifyOracle) Identify() (string, error) {
	data, err := os.ReadFile(bean.OracleSysFile)
	if err != nil {
		impl.Logger.Errorw("error while reading file", "error", err)
		return bean.Unknown, err
	}
	if strings.Contains(string(data), bean.OracleIdentifierString) {
		return bean.Oracle, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyOracle) IdentifyViaMetadataServer(detected chan<- string) {
	r := oracleMetadataResponse{}
	req, err := http.NewRequest("GET", bean.OracleMetadataServerV1, nil)
	if err != nil {
		impl.Logger.Errorw("error while creating new request", "error", err)
		detected <- bean.Unknown
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		impl.Logger.Errorw("error while requesting", "error", err, "request", req)
		detected <- bean.Unknown
		return
	}
	if resp.StatusCode == http.StatusNotFound {
		req, err = http.NewRequest("GET", bean.OracleMetadataServerV2, nil)
		if err != nil {
			impl.Logger.Errorw("error while creating new request", "error", err)
			detected <- bean.Unknown
			return
		}
		req.Header.Set("Authorization", "Bearer Oracle")
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			impl.Logger.Errorw("error while requesting", "error", err, "request", req)
			detected <- bean.Unknown
			return
		}
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			impl.Logger.Errorw("error while reading response body", "error", err, "respBody", resp.Body)
			detected <- bean.Unknown
			return
		}
		err = json.Unmarshal(body, &r)
		if err != nil {
			impl.Logger.Errorw("error while unmarshaling json", "error", err, "body", body)
			detected <- bean.Unknown
			return
		}
		if strings.Contains(r.OkeTM, "oke") {
			detected <- bean.Oracle
			return
		}
	} else {
		detected <- bean.Unknown
		return
	}
}
