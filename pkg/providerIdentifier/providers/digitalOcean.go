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

type digitalOceanMetadataResponse struct {
	DropletID int `json:"droplet_id"`
}

type IdentifyDigitalOcean struct {
	Logger *zap.SugaredLogger
}

func (impl *IdentifyDigitalOcean) Identify() (string, error) {
	data, err := os.ReadFile(bean.DigitalOceanSysFile)
	if err != nil {
		return bean.Unknown, err
	}
	if strings.Contains(string(data), bean.DigitalOceanIdentifierString) {
		return bean.DigitalOcean, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyDigitalOcean) IdentifyViaMetadataServer(detected chan<- string) {
	r := digitalOceanMetadataResponse{}
	req, err := http.NewRequest("GET", bean.DigitalOceanMetadataServer, nil)
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
		if r.DropletID > 0 {
			detected <- bean.DigitalOcean
		}
	} else {
		detected <- bean.Unknown
		return
	}
}
