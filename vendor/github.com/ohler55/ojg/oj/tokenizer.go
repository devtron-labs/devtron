// Copyright (c) 2021, Peter Ohler, All rights reserved.

package oj

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"unicode/utf8"

	"github.com/ohler55/ojg/gen"
)

const (
	objectStart = '{'
	arrayStart  = '['
)

// Tokenizer is used to tokenize a JSON document.
type Tokenizer struct {
	tracker
	tmp       []byte // used for numbers and strings
	runeBytes []byte
	starts    []byte
	handler   TokenHandler
	ri        int // read index for null, false, and true
	mi        int
	num       gen.Number
	rn        rune
	mode      string
	nextMode  string
}

// TokenizeString the provided JSON and call the handler functions for each
// token in the JSON.
func TokenizeString(data string, handler TokenHandler) error {
	t := Tokenizer{}
	return t.Parse([]byte(data), handler)
}

// Tokenize the provided JSON and call the TokenHandler functions for each
// token in the JSON.
func Tokenize(data []byte, handler TokenHandler) error {
	t := Tokenizer{}
	return t.Parse(data, handler)
}

// TokenizeLoad JSON from a io.Reader and call the TokenHandler functions for
// each token in the JSON.
func TokenizeLoad(r io.Reader, handler TokenHandler) error {
	t := Tokenizer{}
	return t.Load(r, handler)
}

// Parse the JSON and call the handler functions for each token in the JSON.
func (t *Tokenizer) Parse(buf []byte, handler TokenHandler) (err error) {
	t.handler = handler
	if t.starts == nil {
		t.tmp = make([]byte, 0, tmpInitSize)
		t.starts = make([]byte, 0, 16)
	} else {
		t.tmp = t.tmp[:0]
		t.starts = t.starts[:0]
	}
	t.noff = -1
	t.line = 1
	t.mode = valueMap
	t.mi = 0
	// Skip BOM if present.
	if 3 < len(buf) && buf[0] == 0xEF {
		if buf[1] == 0xBB && buf[2] == 0xBF {
			err = t.tokenizeBuffer(buf[3:], true)
		} else {
			err = fmt.Errorf("expected BOM at 1:3")
		}
	} else {
		err = t.tokenizeBuffer(buf, true)
	}
	return
}

// Load aand parse the JSON and call the handler functions for each token in
// the JSON.
func (t *Tokenizer) Load(r io.Reader, handler TokenHandler) (err error) {
	t.handler = handler
	if t.starts == nil {
		t.tmp = make([]byte, 0, tmpInitSize)
		t.starts = make([]byte, 0, 16)
	} else {
		t.tmp = t.tmp[:0]
		t.starts = t.starts[:0]
	}
	t.noff = -1
	t.line = 1
	t.mi = 0
	buf := make([]byte, readBufSize)
	eof := false
	var cnt int
	cnt, err = r.Read(buf)
	buf = buf[:cnt]
	t.mode = valueMap
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return
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
			err = t.tokenizeBuffer(buf[skip:], eof)
		} else {
			err = t.tokenizeBuffer(buf, eof)
		}
		if err != nil {
			return
		}
		skip = 0
		if eof {
			break
		}
		buf = buf[:cap(buf)]
		cnt, err = r.Read(buf)
		buf = buf[:cnt]
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return
			}
			eof = true
		}
	}
	return
}

func (t *Tokenizer) tokenizeBuffer(buf []byte, last bool) error {
	var b byte
	var i int
	var off int
	depth := len(t.starts)
	for off = 0; off < len(buf); off++ {
		b = buf[off]
		switch t.mode[b] {
		case skipNewline:
			t.line++
			t.noff = off
			for i, b = range buf[off+1:] {
				if spaceMap[b] != skipChar {
					break
				}
			}
			off += i
			continue
		case colonColon:
			t.mode = valueMap
			continue
		case skipChar: // skip and continue
			continue
		case strOk:
			t.tmp = append(t.tmp, b)
		case keyQuote:
			start := off + 1
			if len(buf) <= start {
				t.tmp = t.tmp[:0]
				t.mode = stringMap
				t.nextMode = colonMap
				continue
			}
			for i, b = range buf[off+1:] {
				if stringMap[b] != strOk {
					break
				}
			}
			off += i
			if b == '"' {
				off++
				t.handler.Key(string(buf[start:off]))
				t.mode = colonMap
			} else {
				t.tmp = t.tmp[:0]
				t.tmp = append(t.tmp, buf[start:off+1]...)
				t.mode = stringMap
				t.nextMode = colonMap
			}
			continue
		case afterComma:
			if 0 < len(t.starts) && t.starts[len(t.starts)-1] == '{' {
				t.mode = keyMap
			} else {
				t.mode = commaMap
			}
			continue
		case valQuote:
			start := off + 1
			if len(buf) <= start {
				t.tmp = t.tmp[:0]
				t.mode = stringMap
				t.nextMode = afterMap
				continue
			}
			for i, b = range buf[off+1:] {
				if stringMap[b] != strOk {
					break
				}
			}
			off += i
			if b == '"' {
				off++
				t.handler.String(string(buf[start:off]))
				t.mode = afterMap
			} else {
				t.tmp = t.tmp[:0]
				t.tmp = append(t.tmp, buf[start:off+1]...)
				t.mode = stringMap
				t.nextMode = afterMap
				continue
			}
		case numComma:
			t.handleNum()
			if 0 < len(t.starts) && t.starts[len(t.starts)-1] == '{' {
				t.mode = keyMap
			} else {
				t.mode = commaMap
			}
		case strSlash:
			t.mode = escMap
			continue
		case escOk:
			t.tmp = append(t.tmp, escByteMap[b])
			t.mode = stringMap
			continue
		case openObject:
			t.starts = append(t.starts, objectStart)
			t.handler.ObjectStart()
			t.mode = key1Map
			depth++
			continue
		case closeObject:
			depth--
			if depth < 0 || t.starts[depth] != objectStart {
				return t.newError(off, "unexpected object close")
			}
			if 256 < len(t.mode) && t.mode[256] == 'n' {
				t.handleNum()
			}
			t.starts = t.starts[0:depth]
			t.handler.ObjectEnd()
			t.mode = afterMap
		case val0:
			t.mode = zeroMap
			t.num.Reset()
		case valDigit:
			t.num.Reset()
			t.mode = digitMap
			t.num.I = uint64(b - '0')
			for i, b = range buf[off+1:] {
				if digitMap[b] != numDigit {
					break
				}
				t.num.I = t.num.I*10 + uint64(b-'0')
				if math.MaxInt64 < t.num.I {
					t.num.FillBig()
					break
				}
			}
			if digitMap[b] == numDigit {
				off++
			}
			off += i
		case valNeg:
			t.mode = negMap
			t.num.Reset()
			t.num.Neg = true
			continue
		case escU:
			t.mode = uMap
			t.rn = 0
			t.ri = 0
			continue
		case openArray:
			t.starts = append(t.starts, arrayStart)
			t.handler.ArrayStart()
			t.mode = valueMap
			depth++
			continue
		case closeArray:
			depth--
			if depth < 0 || t.starts[depth] != arrayStart {
				return t.newError(off, "unexpected array close")
			}
			// Only modes with a close array are value, after, and numbers
			// which are all over 256 long.
			if t.mode[256] == 'n' {
				t.handleNum()
			}
			t.starts = t.starts[:len(t.starts)-1]
			t.handler.ArrayEnd()
			t.mode = afterMap
		case valNull:
			if off+4 <= len(buf) && string(buf[off:off+4]) == "null" {
				off += 3
				t.mode = afterMap
				t.handler.Null()
			} else {
				t.mode = nullMap
				t.ri = 0
			}
		case valTrue:
			if off+4 <= len(buf) && string(buf[off:off+4]) == "true" {
				off += 3
				t.mode = afterMap
				t.handler.Bool(true)
			} else {
				t.mode = trueMap
				t.ri = 0
			}
		case valFalse:
			if off+5 <= len(buf) && string(buf[off:off+5]) == "false" {
				off += 4
				t.mode = afterMap
				t.handler.Bool(false)
			} else {
				t.mode = falseMap
				t.ri = 0
			}
		case numDot:
			if 0 < len(t.num.BigBuf) {
				t.num.BigBuf = append(t.num.BigBuf, b)
				t.mode = dotMap
				continue
			}
			for i, b = range buf[off+1:] {
				if digitMap[b] != numDigit {
					break
				}
				t.num.Frac = t.num.Frac*10 + uint64(b-'0')
				t.num.Div *= 10.0
				if math.MaxInt64 < t.num.Frac {
					t.num.FillBig()
					break
				}
			}
			off += i
			if digitMap[b] == numDigit {
				off++
			}
			t.mode = fracMap
		case numFrac:
			t.num.AddFrac(b)
			t.mode = fracMap
		case fracE:
			if 0 < len(t.num.BigBuf) {
				t.num.BigBuf = append(t.num.BigBuf, b)
			}
			t.mode = expSignMap
			continue
		case strQuote:
			t.mode = t.nextMode
			if t.nextMode == colonMap {
				t.handler.Key(string(t.tmp))
			} else {
				t.handler.String(string(t.tmp))
			}
		case numZero:
			t.mode = zeroMap
		case numDigit:
			t.num.AddDigit(b)
		case negDigit:
			t.num.AddDigit(b)
			t.mode = digitMap
		case numSpc:
			t.handleNum()
			t.mode = afterMap
		case numNewline:
			t.handleNum()
			t.line++
			t.noff = off
			t.mode = afterMap
			for i, b = range buf[off+1:] {
				if spaceMap[b] != skipChar {
					break
				}
			}
			off += i
		case expSign:
			t.mode = expZeroMap
			if b == '-' {
				t.num.NegExp = true
			}
			continue
		case expDigit:
			t.num.AddExp(b)
			t.mode = expMap
		case uOk:
			t.ri++
			switch b {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				t.rn = t.rn<<4 | rune(b-'0')
			case 'a', 'b', 'c', 'd', 'e', 'f':
				t.rn = t.rn<<4 | rune(b-'a'+10)
			case 'A', 'B', 'C', 'D', 'E', 'F':
				t.rn = t.rn<<4 | rune(b-'A'+10)
			}
			if t.ri == 4 {
				if len(t.runeBytes) < 6 {
					t.runeBytes = make([]byte, 6)
				}
				n := utf8.EncodeRune(t.runeBytes, t.rn)
				t.tmp = append(t.tmp, t.runeBytes[:n]...)
				t.mode = stringMap
			}
			continue
		case tokenOk:
			switch {
			case t.mode['r'] == tokenOk:
				t.ri++
				if "true"[t.ri] != b {
					return t.newError(off, "expected true")
				}
				if 3 <= t.ri {
					t.handler.Bool(true)
					t.mode = afterMap
				}
			case t.mode['a'] == tokenOk:
				t.ri++
				if "false"[t.ri] != b {
					return t.newError(off, "expected false")
				}
				if 4 <= t.ri {
					t.handler.Bool(false)
					t.mode = afterMap
				}
			case t.mode['u'] == tokenOk && t.mode['l'] == tokenOk:
				t.ri++
				if "null"[t.ri] != b {
					return t.newError(off, "expected null")
				}
				if 3 <= t.ri {
					t.handler.Null()
					t.mode = afterMap
				}
			}
		case charErr:
			return t.byteError(off, t.mode, b, bytes.Runes(buf[off:])[0])
		}
		if depth == 0 && 256 < len(t.mode) && t.mode[256] == 'a' {
			t.mi = 0
			if t.OnlyOne {
				t.mode = spaceMap
			} else {
				t.mode = valueMap
			}
		}
	}
	if last {
		if len(t.mode) == 256 { // valid finishing maps are one byte longer
			return t.newError(off, "incomplete JSON")
		}
		if t.mode[256] == 'n' {
			t.handleNum()
		}
	}
	return nil
}

func (t *Tokenizer) handleNum() {
	switch tn := t.num.AsNum().(type) {
	case int64:
		t.handler.Int(tn)
	case float64:
		t.handler.Float(tn)
	case json.Number:
		t.handler.Number(string(tn))
	}
}
