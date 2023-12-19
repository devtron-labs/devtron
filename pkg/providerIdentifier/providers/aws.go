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
	data, err := os.ReadFile(bean.AmazonSysFile)
	if err != nil {
		impl.Logger.Errorw("error while reading file", "error", err)
		return bean.Unknown, err
	}
	if strings.Contains(string(data), bean.AmazonIdentifierString) {
		return bean.Amazon, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyAmazon) IdentifyViaMetadataServer(detected chan<- string) {
	r := instanceIdentityResponse{}
	req, err := http.NewRequest("GET", bean.AmazonMetadataServer, nil)
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
