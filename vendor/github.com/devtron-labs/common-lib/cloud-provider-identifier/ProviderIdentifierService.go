/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package providerIdentifier

import (
	"github.com/devtron-labs/common-lib/cloud-provider-identifier/bean"
	"github.com/devtron-labs/common-lib/cloud-provider-identifier/providers"
	"go.uber.org/zap"
	"sync"
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

	identifiedProv := bean.Unknown
	var err error
	for _, prov := range identifiers {
		identifiedProv, err = prov.Identify()
		if err != nil {
			continue
		}
		if identifiedProv != bean.Unknown {
			return identifiedProv, nil
		}
	}

	detected := make(chan string, len(identifiers))

	provs := make([]func(chan<- string), 0, len(identifiers))
	var wg sync.WaitGroup
	for _, prov := range identifiers {
		provs = append(provs, prov.IdentifyViaMetadataServer)
	}

	wg.Add(len(provs))
	for _, function := range provs {
		go func(f func(chan<- string)) {
			defer wg.Done()
			f(detected)
		}(function)
	}
	wg.Wait()
	// closing the channel when all tasks are done
	close(detected)

	for d := range detected {
		if d != bean.Unknown {
			return d, nil
		}
	}
	return bean.Unknown, nil
}
