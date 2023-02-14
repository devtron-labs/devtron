package casbin

type NoopEnforcer struct {
	*EnforcerImpl
}

func NewNoopEnforcer() *NoopEnforcer {
	return &NoopEnforcer{}
}

func (e *NoopEnforcer) Enforce(token string, resource string, action string, resourceItem string) bool {
	return true
}

func (e *NoopEnforcer) EnforceByEmail(emailId string, resource string, action string, resourceItem string) bool {
	return true
}
