package govaluate

import (
	"sync"
	"unicode/utf8"
)

type lexerStream struct {
	sourceString string
	source       []rune
	strPosition  int
	position     int
	length       int
}

var lexerStreamPool = sync.Pool{
	New: func() interface{} {
		return new(lexerStream)
	},
}

func newLexerStream(source string) *lexerStream {
	ret := lexerStreamPool.Get().(*lexerStream)
	if ret.source == nil {
		ret.source = make([]rune, 0, len(source))
	}
	for _, character := range source {
		ret.source = append(ret.source, character)
	}
	ret.sourceString = source
	ret.position = 0
	ret.strPosition = 0
	ret.length = len(ret.source)
	return ret
}

func (stream *lexerStream) readCharacter() rune {
	character := stream.source[stream.position]
	stream.position += 1
	stream.strPosition += utf8.RuneLen(character)
	return character
}

func (stream *lexerStream) rewind(amount int) {
	if amount < 0 {
		stream.position -= amount
		stream.strPosition -= amount
	}
	strAmount := 0
	for i := 0; i < amount; i++ {
		if stream.position >= stream.length {
			strAmount += 1
			stream.position -= 1
			continue
		}
		strAmount += utf8.RuneLen(stream.source[stream.position-1])
		stream.position -= 1
	}
	stream.strPosition -= strAmount
}

func (stream lexerStream) canRead() bool {
	return stream.position < stream.length
}

func (stream *lexerStream) close() {
	stream.source = stream.source[:0]
	lexerStreamPool.Put(stream)
}
