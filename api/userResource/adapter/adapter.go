package adapter

import (
	"github.com/devtron-labs/devtron/api/userResource/bean"
)

func BuildPathParams(kind, version string) *bean.PathParams {
	return &bean.PathParams{
		Kind:    kind,
		Version: version,
	}
}
