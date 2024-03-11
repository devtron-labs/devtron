package providers

import (
	"net/http"
	"os"
	"strings"

	"github.com/devtron-labs/common-lib/cloud-provider-identifier/bean"
	"go.uber.org/zap"
)

type IdentifyGoogle struct {
	Logger *zap.SugaredLogger
}

func (impl *IdentifyGoogle) Identify() (string, error) {
	data, err := os.ReadFile(bean.GoogleSysFile)
	if err != nil {
		impl.Logger.Errorw("error while reading file", "error", err)
		return bean.Unknown, err
	}
	if strings.Contains(string(data), bean.GoogleIdentifierString) {
		return bean.Google, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyGoogle) IdentifyViaMetadataServer(detected chan<- string) {
	req, err := http.NewRequest("GET", bean.GoogleMetadataServer, nil)
	if err != nil {
		impl.Logger.Errorw("error while creating new request", "error", err)
		detected <- bean.Unknown
		return
	}
	req.Header.Set("Metadata-Flavor", "Google")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		impl.Logger.Errorw("error while requesting", "error", err, "request", req)
		detected <- bean.Unknown
		return
	}
	if resp.StatusCode == http.StatusOK {
		detected <- bean.Google
		return
	} else {
		detected <- bean.Unknown
		return
	}
}
