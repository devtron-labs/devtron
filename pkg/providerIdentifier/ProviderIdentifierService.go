package providerIdentifier

import (
	bean2 "github.com/devtron-labs/devtron/pkg/providerIdentifier/bean"
	"github.com/devtron-labs/devtron/pkg/providerIdentifier/providers"
	"go.uber.org/zap"
)

type Identifier interface {
	Identify() (string, error)
	IdentifyViaMetadataServer(detected chan<- string)
}

type ProviderIdentifierService interface {
	IdentifyProvider() (string, error)
}

type ProviderIdentifierServiceImpl struct {
	logger *zap.SugaredLogger
}

func NewProviderIdentifierServiceImpl(
	logger *zap.SugaredLogger) *ProviderIdentifierServiceImpl {
	providerIdentifierService := &ProviderIdentifierServiceImpl{
		logger: logger,
	}
	return providerIdentifierService
}

func (impl *ProviderIdentifierServiceImpl) IdentifyProvider() (string, error) {
	identifiers := []Identifier{
		&providers.IdentifyAlibaba{Logger: impl.logger},
		&providers.IdentifyAmazon{Logger: impl.logger},
		&providers.IdentifyAzure{Logger: impl.logger},
		&providers.IdentifyDigitalOcean{Logger: impl.logger},
		&providers.IdentifyGoogle{Logger: impl.logger},
		&providers.IdentifyOracle{Logger: impl.logger},
	}

	identifiedProv := bean2.Unknown
	var err error
	for _, prov := range identifiers {
		identifiedProv, err = prov.Identify()
		if err != nil {
			continue
		}
		if identifiedProv != bean2.Unknown {
			return identifiedProv, nil
		}
	}

	detected := make(chan string)
	defer close(detected)

	provs := make([]func(chan<- string), 0)
	for _, prov := range identifiers {
		provs = append(provs, prov.IdentifyViaMetadataServer)
	}

	for _, functions := range provs {
		go functions(detected)
	}
	for range provs {
		d := <-detected
		if d != bean2.Unknown {
			return d, nil
		}
	}
	return bean2.Unknown, nil
}
