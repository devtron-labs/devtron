package providers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/devtron-labs/devtron/pkg/providerIdentifier/bean"
	"go.uber.org/zap"
)

type oracleMetadataResponse struct {
	OkeTM string `json:"oke-tm"`
}

type IdentifyOracle struct {
	Logger *zap.SugaredLogger
}

func (impl *IdentifyOracle) Identify() (string, error) {
	data, err := os.ReadFile("/sys/class/dmi/id/chassis_asset_tag")
	if err != nil {
		return bean.Unknown, err
	}
	if strings.Contains(string(data), "OracleCloud") {
		return bean.Oracle, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyOracle) IdentifyViaMetadataServer(detected chan<- string) {
	r := oracleMetadataResponse{}
	req, err := http.NewRequest("GET", "http://169.254.169.254/opc/v1/instance/metadata/", nil)
	if err != nil {
		detected <- bean.Unknown
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		detected <- bean.Unknown
		return
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			detected <- bean.Unknown
			return
		}
		err = json.Unmarshal(body, &r)
		if err != nil {
			detected <- bean.Unknown
			return
		}
		if strings.Contains(r.OkeTM, "oke") {
			detected <- bean.Oracle
		}
	} else {
		detected <- bean.Unknown
		return
	}
}
