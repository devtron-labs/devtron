package token

func newCurrentToken() *currentToken {
	return &currentToken{}
}

type currentToken struct {
}

func (token *currentToken) String() string {
	return "@"
}

func (token *currentToken) Type() string {
	return "current"
}

func (token *currentToken) Apply(root, current interface{}, next []Token) (interface{}, error) {
	if len(next) > 0 {
		return next[0].Apply(root, current, next[1:])
	}
	return current, nil
}
