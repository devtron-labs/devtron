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

type instanceIdentityResponse struct {
	ImageID    string `json:"imageId"`
	InstanceID string `json:"instanceId"`
}

type IdentifyAmazon struct {
	Logger *zap.SugaredLogger
}

func (impl *IdentifyAmazon) Identify() (string, error) {
	data, err := os.ReadFile("/sys/class/dmi/id/product_version")
	if err != nil {
		return bean.Unknown, err
	}
	if strings.Contains(string(data), "amazon") {
		return bean.Amazon, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyAmazon) IdentifyViaMetadataServer(detected chan<- string) {
	r := instanceIdentityResponse{}
	req, err := http.NewRequest("GET", "http://169.254.169.254/latest/dynamic/instance-identity/document", nil)
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
		if strings.HasPrefix(r.ImageID, "ami-") &&
			strings.HasPrefix(r.InstanceID, "i-") {
			detected <- bean.Amazon
			return
		}
	} else {
		detected <- bean.Unknown
		return
	}

}
