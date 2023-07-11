package protect

const BASE_CONFIG_ENV_ID = -1

type ProtectionState int

const (
	EnabledProtectionState  ProtectionState = 0
	DisabledProtectionState ProtectionState = 1
)

type ResourceType int

const (
	ConfigProtectionResourceType ResourceType = 0
)

type ResourceProtectRequest struct {
	AppId           int             `json:"appId" validate:"number,required"`
	EnvId           int             `json:"envId" validate:"number,required"`
	ProtectionState ProtectionState `json:"state" validate:"number,required"`
	UserId          int32           `json:"-"`
}
