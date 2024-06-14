package adapter

import (
	"github.com/devtron-labs/devtron/pkg/devtronResource/bean"
	"strings"
)

func BuildFilterCriteriaDecoder(kindReq, identifierType, value string) *bean.FilterCriteriaDecoder {
	var kind, subKind string
	kindSplits := strings.Split(kindReq, "/")
	if len(kindSplits) == 2 {
		subKind = kindSplits[1]
	}
	kind = kindSplits[0]
	return &bean.FilterCriteriaDecoder{
		Kind:    bean.DevtronResourceKind(kind),
		SubKind: bean.DevtronResourceKind(subKind),
		Type:    bean.FilterCriteriaIdentifier(identifierType),
		Value:   value,
	}
}
