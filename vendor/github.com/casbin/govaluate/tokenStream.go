package govaluate

type tokenStream struct {
	tokens      []ExpressionToken
	index       int
	tokenLength int
}

func newTokenStream(tokens []ExpressionToken) *tokenStream {
	ret := new(tokenStream)
	ret.tokens = tokens
	ret.tokenLength = len(tokens)
	return ret
}

func (this *tokenStream) rewind() {
	this.index -= 1
}

func (this *tokenStream) next() ExpressionToken {
	token := this.tokens[this.index]

	this.index += 1
	return token
}

func (this tokenStream) hasNext() bool {

	return this.index < this.tokenLength
}
