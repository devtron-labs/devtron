package govaluate

import "sync"

type tokenStream struct {
	tokens      []ExpressionToken
	index       int
	tokenLength int
}

var tokenStreamPool = sync.Pool{
	New: func() interface{} {
		return new(tokenStream)
	},
}

func newTokenStream(tokens []ExpressionToken) *tokenStream {
	ret := tokenStreamPool.Get().(*tokenStream)
	ret.tokens = tokens
	ret.index = 0
	ret.tokenLength = len(tokens)
	return ret
}

func (t *tokenStream) rewind() {
	t.index -= 1
}

func (t *tokenStream) next() ExpressionToken {
	token := t.tokens[t.index]

	t.index += 1
	return token
}

func (t tokenStream) hasNext() bool {

	return t.index < t.tokenLength
}

func (t *tokenStream) close() {
	tokenStreamPool.Put(t)
}
