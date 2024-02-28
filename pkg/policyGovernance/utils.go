package policyGovernance

import "github.com/devtron-labs/devtron/pkg/globalPolicy/bean"

const NO_POLICY = "NA"

type AppEnvPolicyContainer struct {
	AppId      int    `json:"-"`
	EnvId      int    `json:"-"`
	PolicyId   int    `json:"-"`
	AppName    string `json:"appName"`
	EnvName    string `json:"envName"`
	PolicyName string `json:"policyName,omitempty"`
}

type AppEnvPolicyMappingsListFilter struct {
	PolicyType  bean.GlobalPolicyType `json:"policyType"`
	AppNames    []string              `json:"appNames"`
	EnvNames    []string              `json:"envNames"`
	PolicyNames []string              `json:"policyNames"`
	SortBy      string                `json:"sortBy,omitempty" validate:"oneof=appName environmentName"`
	SortOrder   string                `json:"sortOrder,omitempty" validate:"oneof=ASC DESC"`
	Offset      int                   `json:"offset,omitempty" validate:"min=0"`
	Size        int                   `json:"size,omitempty" validate:"min=0"`
}

type BulkPromotionPolicyApplyRequest struct {
	PolicyType              bean.GlobalPolicyType           `json:"policyType"`
	ApplicationEnvironments []AppEnvPolicyContainer         `json:"applicationEnvironments"`
	ApplyToPolicyName       string                          `json:"applyToPolicyName"`
	AppEnvPolicyListFilter  *AppEnvPolicyMappingsListFilter `json:"appEnvPolicyListFilter"`
}
