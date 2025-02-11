package token

func newRootToken() *rootToken {
	return &rootToken{}
}

type rootToken struct {
}

func (token *rootToken) String() string {
	return "$"
}

func (token *rootToken) Type() string {
	return "root"
}

func (token *rootToken) Apply(root, current interface{}, next []Token) (interface{}, error) {
	if len(next) > 0 {
		return next[0].Apply(root, root, next[1:])
	}
	return root, nil
}
