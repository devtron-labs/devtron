package bean

type ExpressionFormat int

// TimeStamp currently timestamp format is only supported , have to add support for window,iso etc
const (
	TimeStamp      ExpressionFormat = 1
	TimeZeroFormat ExpressionFormat = 2
)
