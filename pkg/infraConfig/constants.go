package infraConfig

import "github.com/pkg/errors"

type ConfigKey int
type ConfigKeyStr string
type ProfileType string
type IdentifierType string

const ALL_PROFILES = ""
const CannotDeleteDefaultProfile = "cannot delete default profile"
const profileApplyErr = "selected filter does not match any apps, cannot apply the given profile"
const APPLICATION IdentifierType = "application"
const InvalidIdentifierType = "identifier %s is not valid"
const PROFILE_IDS_REQUIRED = "profile ids cannot be empty"
const ActiveAppIdsQuery = "SELECT id " +
	"FROM app WHERE active=true"

const NORMAL ProfileType = "NORMAL"
const InvalidUnit = "invalid %s unit found in %s "
const DEFAULT_PROFILE_NAME = "default"
const DEFAULT_PROFILE_EXISTS = "default profile exists"
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

const CREATION_BLOCKED_FOR_DEFAULT_PROFILE_CONFIGURATIONS = "cannot create new configuration for default profile"

var NO_PROPERTIES_FOUND_ERROR = errors.New("no properties found")
