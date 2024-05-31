/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package bean

import (
	"github.com/devtron-labs/devtron/pkg/auth/authorisation/casbin/bean"
)

type PolicyRequest struct {
	Data []bean.Policy
}
