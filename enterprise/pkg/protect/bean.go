package protect

const BASE_CONFIG_ENV_ID = -1

type ProtectionState int

const (
	EnabledProtectionState  ProtectionState = 1
	DisabledProtectionState ProtectionState = 2
)

type ResourceType int

const (
	ConfigProtectionResourceType ResourceType = 1
)

type ResourceProtectRequest struct {
	AppId           int             `json:"appId" validate:"number,required"`
	EnvId           int             `json:"envId" validate:"number,required"`
	ProtectionState ProtectionState `json:"state" validate:"number,required"`
	UserId          int32           `json:"-"`
}
