package bean

type Policy struct {
	Type PolicyType `json:"type"`
	Sub  Subject    `json:"sub"`
	Res  Resource   `json:"res"`
	Act  Action     `json:"act"`
	Obj  Object     `json:"obj"`
}

type Subject string
type Resource string
type Action string
type Object string
type PolicyType string

type GroupPolicy struct {
	Role                    string
	TimeoutWindowExpression string
	ExpressionFormat        string
}
