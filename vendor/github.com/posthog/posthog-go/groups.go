package posthog

// This type is used to represent groups in messages that support it.
// It is a free-form object so the application can set any value it sees fit but
// a few helper method are defined to make it easier to instantiate groups with
// common fields.

type Groups map[string]interface{}

func NewGroups() Groups {
	return make(Groups, 10)
}

func (p Groups) Set(name string, value interface{}) Groups {
	p[name] = value
	return p
}
