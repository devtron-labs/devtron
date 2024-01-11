// Copyright (c) 2020, Peter Ohler, All rights reserved.

package oj

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

const stackMinSize = 32 // for container stack { or [

// Validator is a reusable JSON validator. It can be reused for multiple
// validations or parsings which allows buffer reuse for a performance
// advantage.
type Validator struct {
	tracker

	// This and the Parser use the same basic code but without the
	// building. It is a copy since adding the conditionals needed to avoid
	// building results add 15 to 20% overhead. An additional improvement could
	// be made by not tracking line and column but that would make it
	// validation much less useful.
	stack    []byte // { or [
	ri       int    // read index for null, false, and true
	mode     string
	nextMode string

	// OnlyOne returns an error if more than one JSON is in the string or
	// stream.
	OnlyOne bool
}

// Validate a JSON encoded byte slice.
func (p *Validator) Validate(buf []byte) (err error) {
	if cap(p.stack) < stackMinSize {
		p.stack = make([]byte, 0, stackMinSize)
	} else {
		p.stack = p.stack[:0]
	}
	p.noff = -1
	p.line = 1
	p.mode = valueMap
	// Skip BOM if present.
	if 3 < len(buf) && buf[0] == 0xEF {
		if buf[1] == 0xBB && buf[2] == 0xBF {
			err = p.validateBuffer(buf[3:], true)
		} else {
			err = fmt.Errorf("expected BOM at 1:3")
		}
	} else {
		err = p.validateBuffer(buf, true)
	}
	return
}

// ValidateReader a JSON stream. An error is returned if not valid JSON.
func (p *Validator) ValidateReader(r io.Reader) error {
	if cap(p.stack) < stackMinSize {
		p.stack = make([]byte, 0, stackMinSize)
	} else {
		p.stack = p.stack[:0]
	}
	p.noff = -1
	p.line = 1
	p.mode = valueMap
	buf := make([]byte, readBufSize)
	eof := false
	cnt, err := r.Read(buf)
	buf = buf[:cnt]
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}
		eof = true
	}
	var skip int
	// Skip BOM if present.
	if 3 < len(buf) && buf[0] == 0xEF && buf[1] == 0xBB && buf[2] == 0xBF {
		skip = 3
	}
	for {
		if 0 < skip {
			err = p.validateBuffer(buf[skip:], eof)
		} else {
			err = p.validateBuffer(buf, eof)
		}
		skip = 0
		if err != nil {
			return err
		}
		p.noff -= len(buf)
		if eof {
			break
		}
		buf = buf[:cap(buf)]
		cnt, err := r.Read(buf)
		buf = buf[:cnt]
		if err != nil {
			if err != io.EOF {
				return err
			}
			eof = true
		}
	}
	return nil
}

func (p *Validator) validateBuffer(buf []byte, last bool) error {
	var b byte
	var i int
	var off int
	depth := len(p.stack)
	for off = 0; off < len(buf); off++ {
		b = buf[off]
		switch p.mode[b] {
		case skipNewline:
			p.line++
			p.noff = off
			for i, b = range buf[off+1:] {
				if spaceMap[b] != skipChar {
					break
				}
			}
			off += i
			continue
		case colonColon:
			p.mode = valueMap
			continue
		case skipChar:
			continue
		case strOk:
			continue
		case keyQuote:
			i = 0
			for i, b = range buf[off+1:] {
				if stringMap[b] != strOk {
					break
				}
			}
			off += i
			if b == '"' && 0 < i {
				off++
				p.mode = colonMap
			} else {
				p.mode = stringMap
				p.nextMode = colonMap
			}
			continue
		case afterComma:
			if 0 < len(p.stack) && p.stack[len(p.stack)-1] == '{' {
				p.mode = keyMap
			} else {
				p.mode = commaMap
			}
			continue
		case valQuote:
			i = 0
			for i, b = range buf[off+1:] {
				if stringMap[b] != strOk {
					break
				}
			}
			off += i
			if b == '"' && 0 < i {
				off++
				p.mode = afterMap
			} else {
				p.mode = stringMap
				p.nextMode = afterMap
				continue
			}
		case numComma:
			if 0 < len(p.stack) && p.stack[len(p.stack)-1] == '{' {
				p.mode = keyMap
			} else {
				p.mode = commaMap
			}
		case strSlash:
			p.mode = escMap
			continue
		case escOk:
			p.mode = stringMap
			continue
		case openObject:
			p.stack = append(p.stack, '{')
			p.mode = key1Map
			depth++
			continue
		case closeObject:
			depth--
			if depth < 0 || p.stack[depth] != '{' {
				return p.newError(off, "unexpected object close")
			}
			p.stack = p.stack[0:depth]
			p.mode = afterMap
		case val0:
			p.mode = zeroMap
			continue
		case valDigit:
			p.mode = digitMap
			continue
		case valNeg:
			p.mode = negMap
			continue
		case escU:
			p.mode = uMap
			p.ri = 0
			continue
		case openArray:
			p.stack = append(p.stack, '[')
			p.mode = valueMap
			depth++
			continue
		case closeArray:
			depth--
			if depth < 0 || p.stack[depth] != '[' {
				return p.newError(off, "unexpected array close")
			}
			p.stack = p.stack[0:depth]
			p.mode = afterMap
		case valNull:
			if off+4 <= len(buf) && string(buf[off:off+4]) == "null" {
				off += 3
				p.mode = afterMap
			} else {
				p.mode = nullMap
				p.ri = 0
			}
		case valTrue:
			if off+4 <= len(buf) && string(buf[off:off+4]) == "true" {
				off += 3
				p.mode = afterMap
			} else {
				p.mode = trueMap
				p.ri = 0
			}
		case valFalse:
			if off+5 <= len(buf) && string(buf[off:off+5]) == "false" {
				off += 4
				p.mode = afterMap
			} else {
				p.mode = falseMap
				p.ri = 0
			}
		case numDot:
			p.mode = dotMap
			continue
		case numFrac:
			p.mode = fracMap
		case fracE:
			p.mode = expSignMap
			continue
		case strQuote:
			p.mode = p.nextMode
		case numZero:
			p.mode = zeroMap
		case numDigit:
			// nothing to do
		case negDigit:
			p.mode = digitMap
		case numSpc:
			p.mode = afterMap
		case numNewline:
			p.line++
			p.noff = off
			p.mode = afterMap
			i = 0
			for i, b = range buf[off+1:] {
				if spaceMap[b] != skipChar {
					break
				}
			}
			off += i
		case expSign:
			p.mode = expZeroMap
			continue
		case expDigit:
			p.mode = expMap
		case uOk:
			p.ri++
			if p.ri == 4 {
				p.mode = stringMap
			}
			continue

		case tokenOk:
			switch {
			case p.mode['r'] == tokenOk:
				p.ri++
				if "true"[p.ri] != b {
					return p.newError(off, "expected true")
				}
				if 3 <= p.ri {
					p.mode = afterMap
				}
			case p.mode['a'] == tokenOk:
				p.ri++
				if "false"[p.ri] != b {
					return p.newError(off, "expected false")
				}
				if 4 <= p.ri {
					p.mode = afterMap
				}
			case p.mode['u'] == tokenOk && p.mode['l'] == tokenOk:
				p.ri++
				if "null"[p.ri] != b {
					return p.newError(off, "expected null")
				}
				if 3 <= p.ri {
					p.mode = afterMap
				}
			}
		case charErr:
			return p.byteError(off, p.mode, b, bytes.Runes(buf[off:])[0])
		}
		if depth == 0 && 256 < len(p.mode) && p.mode[256] == 'a' {
			if p.OnlyOne {
				p.mode = spaceMap
			} else {
				p.mode = valueMap
			}
		}
	}
	if last && len(p.mode) == 256 { // valid finishing maps are one byte longer
		return p.newError(off, "incomplete JSON")
	}
	return nil
}
