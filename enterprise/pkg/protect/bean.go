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
