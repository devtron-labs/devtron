package providers

import (
	"encoding/json"
	"io/ioutil"
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
	data, err := os.ReadFile("/sys/class/dmi/id/sys_vendor")
	if err != nil {
		return bean.Unknown, err
	}
	if strings.Contains(string(data), "DigitalOcean") {
		return bean.DigitalOcean, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyDigitalOcean) IdentifyViaMetadataServer(detected chan<- string) {
	r := digitalOceanMetadataResponse{}
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/v1.json", nil)
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
		body, err := ioutil.ReadAll(resp.Body)
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
