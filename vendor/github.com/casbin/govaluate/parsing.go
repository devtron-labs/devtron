package govaluate

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
)

var (
	averageTokens = 1
	samplesMu     = sync.Mutex{}
	samples       = make([]int, 0, 10)
)

func parseTokens(expression string, functions map[string]ExpressionFunction) ([]ExpressionToken, error) {
	samplesMu.Lock()
	ret := make([]ExpressionToken, 0, averageTokens)
	samplesMu.Unlock()
	var token ExpressionToken
	var stream *lexerStream
	var state lexerState
	var err error
	var found bool

	stream = newLexerStream(expression)
	state = validLexerStates[0]

	for stream.canRead() {

		token, err, found = readToken(stream, state, functions)

		if err != nil {
			return ret, err
		}

		if !found {
			break
		}

		state, err = getLexerStateForToken(token.Kind)
		if err != nil {
			return ret, err
		}

		// append this valid token
		ret = append(ret, token)
	}
	stream.close()
	samplesMu.Lock()
	if len(samples) == cap(samples) {
		copy(samples, samples[1:])
		samples[len(samples)-1] = len(ret)
	} else {
		samples = append(samples, len(ret))
	}
	total := 0
	for _, val := range samples {
		total += val
	}
	averageTokens = total / len(samples)
	samplesMu.Unlock()

	err = checkBalance(ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func readToken(stream *lexerStream, state lexerState, functions map[string]ExpressionFunction) (ExpressionToken, error, bool) {

	var function ExpressionFunction
	var ret ExpressionToken
	var tokenValue interface{}
	var tokenTime time.Time
	var tokenString string
	var kind TokenKind
	var character rune
	var found bool
	var completed bool
	var err error

	// numeric is 0-9, or . or 0x followed by digits
	// string starts with '
	// variable is alphanumeric, always starts with a letter
	// bracket always means variable
	// symbols are anything non-alphanumeric
	// all others read into a buffer until they reach the end of the stream
	for stream.canRead() {

		character = stream.readCharacter()

		if unicode.IsSpace(character) {
			continue
		}

		// numeric constant
		if isNumeric(character) {

			if stream.canRead() && character == '0' {
				character = stream.readCharacter()

				if stream.canRead() && character == 'x' {
					tokenString, _ = readUntilFalse(stream, false, true, true, isHexDigit)
					tokenValueInt, err := strconv.ParseUint(tokenString, 16, 64)

					if err != nil {
						errorMsg := fmt.Sprintf("Unable to parse hex value '%v' to uint64\n", tokenString)
						return ExpressionToken{}, errors.New(errorMsg), false
					}

					kind = NUMERIC
					tokenValue = float64(tokenValueInt)
					break
				} else {
					stream.rewind(1)
				}
			}

			tokenString = readTokenUntilFalse(stream, isNumeric)
			tokenValue, err = strconv.ParseFloat(tokenString, 64)

			if err != nil {
				errorMsg := fmt.Sprintf("Unable to parse numeric value '%v' to float64\n", tokenString)
				return ExpressionToken{}, errors.New(errorMsg), false
			}
			kind = NUMERIC
			break
		}

		// comma, separator
		if character == ',' {

			tokenValue = ","
			kind = SEPARATOR
			break
		}

		// escaped variable
		if character == '[' {

			tokenValue, completed = readUntilFalse(stream, true, false, true, isNotClosingBracket)
			kind = VARIABLE

			if !completed {
				return ExpressionToken{}, errors.New("unclosed parameter bracket"), false
			}

			// above method normally rewinds us to the closing bracket, which we want to skip.
			stream.rewind(-1)
			break
		}

		// regular variable - or function?
		if unicode.IsLetter(character) {

			tokenString = readTokenUntilFalse(stream, isVariableName)
			switch tokenString {
			case "true":
				kind = BOOLEAN
				tokenValue = true
			case "false":
				kind = BOOLEAN
				tokenValue = false
			case "in":
				fallthrough
			case "IN":
				// force lower case for consistency
				tokenValue = "in"
				kind = COMPARATOR
			default:
				// This causes an alloc, avoid it if we can
				tokenValue = tokenString
				kind = VARIABLE
			}

			// function?
			function, found = functions[tokenString]
			if found {
				kind = FUNCTION
				tokenValue = function
			}

			// accessor?
			accessorIndex := strings.Index(tokenString, ".")
			if accessorIndex > 0 {

				// check that it doesn't end with a hanging period
				if tokenString[len(tokenString)-1] == '.' {
					errorMsg := fmt.Sprintf("Hanging accessor on token '%s'", tokenString)
					return ExpressionToken{}, errors.New(errorMsg), false
				}

				kind = ACCESSOR
				splits := strings.Split(tokenString, ".")
				tokenValue = splits
			}
			break
		}

		if !isNotQuote(character) {
			tokenValue, completed = readUntilFalse(stream, true, false, true, isNotQuote)

			if !completed {
				return ExpressionToken{}, errors.New("unclosed string literal"), false
			}

			// advance the stream one position, since reading until false assumes the terminator is a real token
			stream.rewind(-1)

			// check to see if this can be parsed as a time.
			tokenTime, found = tryParseTime(tokenValue.(string))
			if found {
				kind = TIME
				tokenValue = tokenTime
			} else {
				kind = STRING
			}
			break
		}

		if character == '(' {
			tokenValue = character
			kind = CLAUSE
			break
		}

		if character == ')' {
			tokenValue = character
			kind = CLAUSE_CLOSE
			break
		}

		// must be a known symbol
		tokenString = readTokenUntilFalse(stream, isNotAlphanumeric)
		tokenValue = tokenString

		// quick hack for the case where "-" can mean "prefixed negation" or "minus", which are used
		// very differently.
		if state.canTransitionTo(PREFIX) {
			_, found = prefixSymbols[tokenString]
			if found {

				kind = PREFIX
				break
			}
		}
		_, found = modifierSymbols[tokenString]
		if found {

			kind = MODIFIER
			break
		}

		_, found = logicalSymbols[tokenString]
		if found {

			kind = LOGICALOP
			break
		}

		_, found = comparatorSymbols[tokenString]
		if found {

			kind = COMPARATOR
			break
		}

		_, found = ternarySymbols[tokenString]
		if found {

			kind = TERNARY
			break
		}

		errorMessage := fmt.Sprintf("Invalid token: '%s'", tokenString)
		return ret, errors.New(errorMessage), false
	}

	ret.Kind = kind
	ret.Value = tokenValue

	return ret, nil, (kind != UNKNOWN)
}

func readTokenUntilFalse(stream *lexerStream, condition func(rune) bool) string {

	var ret string

	stream.rewind(1)
	ret, _ = readUntilFalse(stream, false, true, true, condition)
	return ret
}

var tokenBufferPool = sync.Pool{
	New: func() interface{} {
		return &bytes.Buffer{}
	},
}

/*
Returns the string that was read until the given [condition] was false, or whitespace was broken.
Returns false if the stream ended before whitespace was broken or condition was met.
*/
func readUntilFalse(stream *lexerStream, includeWhitespace bool, breakWhitespace bool, allowEscaping bool, condition func(rune) bool) (string, bool) {

	tokenBuffer := tokenBufferPool.Get().(*bytes.Buffer)
	tokenBuffer.Reset()
	var character rune

	startPosition := stream.strPosition
	reuseString := true
	trimString := false
	conditioned := false

	for stream.canRead() {

		character = stream.readCharacter()
		if character > utf8.RuneSelf {
			// International runes, we can't just grab from the string in this case
			reuseString = false
		}

		// Use backslashes to escape anything
		if allowEscaping && character == '\\' {
			reuseString = false
			character = stream.readCharacter()
			tokenBuffer.WriteString(string(character))
			continue
		}

		if unicode.IsSpace(character) {
			if breakWhitespace && tokenBuffer.Len() > 0 {
				conditioned = true
				trimString = true
				break
			}
			if !includeWhitespace {
				reuseString = false
				continue
			}
		}

		if condition(character) {
			tokenBuffer.WriteString(string(character))
		} else {
			conditioned = true
			stream.rewind(1)
			break
		}
	}

	// This reduces allocations by just reusing parts of the original source string if applicable
	if reuseString {
		tokenBuffer.Reset()
		tokenBufferPool.Put(tokenBuffer)
		ret := stream.sourceString[startPosition:stream.strPosition]
		if trimString {
			ret = ret[:len(ret)-1]
		}
		return ret, conditioned
	}

	ret := tokenBuffer.String()
	tokenBuffer.Reset()
	tokenBufferPool.Put(tokenBuffer)
	return ret, conditioned
}

/*
Checks to see if any optimizations can be performed on the given [tokens], which form a complete, valid expression.
The returns slice will represent the optimized (or unmodified) list of tokens to use.
*/
func optimizeTokens(tokens []ExpressionToken) ([]ExpressionToken, error) {

	var token ExpressionToken
	var symbol OperatorSymbol
	var err error
	var index int

	for index, token = range tokens {

		// if we find a regex operator, and the right-hand value is a constant, precompile and replace with a pattern.
		if token.Kind != COMPARATOR {
			continue
		}

		symbol = comparatorSymbols[token.Value.(string)]
		if symbol != REQ && symbol != NREQ {
			continue
		}

		index++
		token = tokens[index]
		if token.Kind == STRING {

			token.Kind = PATTERN
			token.Value, err = regexp.Compile(token.Value.(string))

			if err != nil {
				return tokens, err
			}

			tokens[index] = token
		}
	}
	return tokens, nil
}

/*
Checks the balance of tokens which have multiple parts, such as parenthesis.
*/
func checkBalance(tokens []ExpressionToken) error {

	var stream *tokenStream
	var token ExpressionToken
	var parens int

	stream = newTokenStream(tokens)

	for stream.hasNext() {

		token = stream.next()
		if token.Kind == CLAUSE {
			parens++
			continue
		}
		if token.Kind == CLAUSE_CLOSE {
			parens--
			continue
		}
	}

	stream.close()

	if parens != 0 {
		return errors.New("unbalanced parenthesis")
	}
	return nil
}

func isHexDigit(character rune) bool {

	character = unicode.ToLower(character)

	return unicode.IsDigit(character) ||
		character == 'a' ||
		character == 'b' ||
		character == 'c' ||
		character == 'd' ||
		character == 'e' ||
		character == 'f'
}

func isNumeric(character rune) bool {

	return unicode.IsDigit(character) || character == '.'
}

func isNotQuote(character rune) bool {

	return character != '\'' && character != '"'
}

func isNotAlphanumeric(character rune) bool {

	return !unicode.IsDigit(character) &&
		!unicode.IsLetter(character) &&
		character != '(' &&
		character != ')' &&
		character != '[' &&
		character != ']' && // starting to feel like there needs to be an `isOperation` func (#59)
		isNotQuote(character)
}

func isVariableName(character rune) bool {

	return unicode.IsLetter(character) ||
		unicode.IsDigit(character) ||
		character == '_' ||
		character == '.'
}

func isNotClosingBracket(character rune) bool {

	return character != ']'
}

type timeFormat struct {
	format    string
	minLength int
	maxLength int
}

/*
Attempts to parse the [candidate] as a Time.
Tries a series of standardized date formats, returns the Time if one applies,
otherwise returns false through the second return.
*/
func tryParseTime(candidate string) (time.Time, bool) {

	var ret time.Time
	var found bool

	if !strings.Contains(candidate, ":") && !strings.Contains(candidate, "-") {
		// The blow formats either have a : or a - in them. If the string contains neither it cannot be a time string
		return time.Now(), false
	}

	timeFormats := [...]timeFormat{
		{time.ANSIC, len(time.ANSIC) - 1, len(time.ANSIC)},
		{time.UnixDate, len(time.UnixDate) - 1, len(time.ANSIC)},
		{time.RubyDate, len(time.RubyDate), len(time.RubyDate)},
		{time.Kitchen, len(time.Kitchen), len(time.Kitchen) + 1},
		{time.RFC3339, len(time.RFC3339), len(time.RFC3339)},
		{time.RFC3339Nano, len(time.RFC3339Nano), len(time.RFC3339Nano)},
		{"2006-01-02", 10, 10},                         // RFC 3339
		{"2006-01-02 15:04", 16, 16},                   // RFC 3339 with minutes
		{"2006-01-02 15:04:05", 19, 19},                // RFC 3339 with seconds
		{"2006-01-02 15:04:05-07:00", 25, 25},          // RFC 3339 with seconds and timezone
		{"2006-01-02T15Z0700", 18, 18},                 // ISO8601 with hour
		{"2006-01-02T15:04Z0700", 21, 21},              // ISO8601 with minutes
		{"2006-01-02T15:04:05Z0700", 24, 24},           // ISO8601 with seconds
		{"2006-01-02T15:04:05.999999999Z0700", 34, 34}, // ISO8601 with nanoseconds
	}

	for _, format := range timeFormats {
		// Avoid trying to parse formats it could not be to reduce allocation of time parse errors
		if len(candidate) < format.minLength || len(candidate) > format.maxLength {
			continue
		}
		ret, found = tryParseExactTime(candidate, format.format)
		if found {
			return ret, true
		}
	}

	return time.Now(), false
}

func tryParseExactTime(candidate string, format string) (time.Time, bool) {

	var ret time.Time
	var err error

	ret, err = time.ParseInLocation(format, candidate, time.Local)
	if err != nil {
		return time.Now(), false
	}

	return ret, true
}

func getFirstRune(candidate string) rune {

	for _, character := range candidate {
		return character
	}

	return 0
}
