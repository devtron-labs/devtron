package infraConfig

type ConfigKey int
type ConfigKeyStr string
type ProfileType string

const NORMAL ProfileType = "NORMAL"
const InvalidUnit = "invalid %s unit found in %s "
const DEFAULT_PROFILE_NAME = "default"
const DEFAULT_PROFILE_EXISTS = "default profile exists"
const NO_PROPERTIES_FOUND = "no properties found"
const DEFAULT ProfileType = "DEFAULT"
const InvalidProfileName = "profile name is invalid"
const PayloadValidationError = "payload validation failed"
const CPULimReqErrorCompErr = "cpu limit should not be less than cpu request"
const MEMLimReqErrorCompErr = "memory limit should not be less than memory request"

const CPULimit ConfigKey = 1
const CPURequest ConfigKey = 2
const MemoryLimit ConfigKey = 3
const MemoryRequest ConfigKey = 4
const TimeOut ConfigKey = 5

// whenever new constant gets added here ,
// we need to add it in GetDefaultConfigKeysMap method as well

const CPU_LIMIT ConfigKeyStr = "cpu_limit"
const CPU_REQUEST ConfigKeyStr = "cpu_request"
const MEMORY_LIMIT ConfigKeyStr = "memory_limit"
const MEMORY_REQUEST ConfigKeyStr = "memory_request"
const TIME_OUT ConfigKeyStr = "timeout"
