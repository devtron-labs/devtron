package providers

import (
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/devtron-labs/devtron/pkg/providerIdentifier/bean"
	"go.uber.org/zap"
)

//type IdentifyAlibabaInterface interface {
//	Identify() (string, error)
//	IdentifyViaMetadataServer(detected chan<- string)
//}

type IdentifyAlibaba struct {
	Logger *zap.SugaredLogger
}

func (impl *IdentifyAlibaba) Identify() (string, error) {
	data, err := os.ReadFile("/sys/class/dmi/id/product_name")
	if err != nil {
		impl.Logger.Errorw("error while reading file", "error", err)
		return bean.Unknown, err
	}
	if strings.Contains(string(data), "Alibaba Cloud") {
		return bean.Alibaba, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyAlibaba) IdentifyViaMetadataServer(detected chan<- string) {
	req, err := http.NewRequest("GET", "http://100.100.100.200/latest/meta-data/instance/instance-type", nil)
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
		if strings.HasPrefix(string(body), "ecs.") {
			detected <- bean.Alibaba
		}
	} else {
		detected <- bean.Unknown
		return
	}
}
