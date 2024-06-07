// Copyright (c) 2020, Peter Ohler, All rights reserved.

package oj

import "fmt"

type tracker struct {
	line int
	noff int // Offset of last newline from start of buf. Can be negative when using a reader.

	// OnlyOne returns an error if more than one JSON is in the string or stream.
	OnlyOne bool
}

func (t *tracker) newError(off int, format string, args ...any) error {
	return &ParseError{
		Message: fmt.Sprintf(format, args...),
		Line:    t.line,
		Column:  off - t.noff,
	}
}

func (t *tracker) byteError(off int, mode string, b byte, r rune) error {
	err := &ParseError{
		Line:   t.line,
		Column: off - t.noff,
	}
	switch mode {
	case nullMap:
		err.Message = "expected null"
	case trueMap:
		err.Message = "expected true"
	case falseMap:
		err.Message = "expected false"
	case afterMap:
		err.Message = fmt.Sprintf("expected a comma or close, not '%c'", r)
	case key1Map:
		err.Message = fmt.Sprintf("expected a string start or object close, not '%c'", r)
	case keyMap:
		err.Message = fmt.Sprintf("expected a string start, not '%c'", r)
	case colonMap:
		err.Message = fmt.Sprintf("expected a colon, not '%c'", r)
	case negMap, zeroMap, digitMap, dotMap, fracMap, expSignMap, expZeroMap, expMap:
		err.Message = "invalid number"
	case stringMap:
		err.Message = fmt.Sprintf("invalid JSON character 0x%02x", b)
	case escMap:
		err.Message = fmt.Sprintf("invalid JSON escape character '\\%c'", r)
	case uMap:
		err.Message = fmt.Sprintf("invalid JSON unicode character '%c'", r)
	case spaceMap:
		err.Message = fmt.Sprintf("extra characters after close, '%c'", r)
	default:
		err.Message = fmt.Sprintf("unexpected character '%c'", r)
	}
	return err
}
