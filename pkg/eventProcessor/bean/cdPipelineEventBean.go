/*
 * Copyright (c) 2024. Devtron Inc.
 */

package bean

import "github.com/devtron-labs/devtron/api/bean"

type BulkCDDeployEvent struct {
	ValuesOverrideRequest *bean.ValuesOverrideRequest `json:"valuesOverrideRequest"` //TODO migrate this
	UserId                int32                       `json:"userId"`
}
