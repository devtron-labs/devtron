package providers

import (
	"net/http"
	"os"
	"strings"

	"github.com/devtron-labs/devtron/pkg/providerIdentifier/bean"
	"go.uber.org/zap"
)

type IdentifyAzure struct {
	Logger *zap.SugaredLogger
}

func (impl *IdentifyAzure) Identify() (string, error) {
	data, err := os.ReadFile(bean.AzureSysFile)
	if err != nil {
		return bean.Unknown, err
	}
	if strings.Contains(string(data), bean.AzureIdentifierString) {
		return bean.Azure, nil
	}
	return bean.Unknown, nil
}

func (impl *IdentifyAzure) IdentifyViaMetadataServer(detected chan<- string) {
	req, err := http.NewRequest("GET", bean.AzureMetadataServer, nil)
	if err != nil {
		detected <- bean.Unknown
		return
	}
	req.Header.Set("Metadata", "true")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		detected <- bean.Unknown
		return
	}
	if resp.StatusCode == http.StatusOK {
		detected <- bean.Azure
	} else {
		detected <- bean.Unknown
		return
	}
}
