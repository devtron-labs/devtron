package token

import (
	"strconv"
	"strings"

	"github.com/evilmonkeyinc/jsonpath/option"
	"github.com/evilmonkeyinc/jsonpath/script"
)

/** Feature request
support double quotes in keys?
**/

// Token represents a component of a JSON Path selector
type Token interface {
	Apply(root, current interface{}, next []Token) (interface{}, error)
	String() string
	Type() string
}

// Tokenize converts a JSON Path selector to a collection of parsable tokens
func Tokenize(selector string) ([]string, error) {
	if selector == "" {
		return nil, getUnexpectedTokenError("", 0)
	}

	tokens := []string{}
	tokenString := ""

	quoteCount := 0
	openScriptBracket := 0
	closeScriptBracket := 0
	openSubscriptBracket := 0
	closeSubscriptBracket := 0

	for idx, rne := range selector {

		if tokenString == "" {
			quoteCount = 0
			openScriptBracket = 0
			closeScriptBracket = 0
		}
		switch rne {
		case '\'':
			quoteCount++
			break
		case '(':
			openScriptBracket++
			break
		case ')':
			closeScriptBracket++
			break
		case '[':
			openSubscriptBracket++
			break
		case ']':
			closeSubscriptBracket++
			break
		}

		tokenString += string(rne)

		if idx == 0 {
			if tokenString != "$" && tokenString != "@" {
				return nil, getUnexpectedTokenError(string(rne), idx)
			}

			if len(selector) > 1 {
				if next := selector[1]; next != '.' && next != '[' {
					return nil, getUnexpectedTokenError(string(next), idx+1)
				}
			}

			tokens = append(tokens, tokenString[:])
			tokenString = ""
			continue
		}

		if tokenString == "." {
			continue
		}

		if tokenString == ".." {
			// recursive operator
			tokens = append(tokens, tokenString[:])
			tokenString = ""
			continue
		}

		if rne == '[' {
			if tokenString == "[" || tokenString == ".[" {
				// open bracket and at start of token
				continue
			}

			// open bracket in middle of token, new subscript
			if strings.Count(tokenString, "[") > 1 {
				// this is not the only opening bracket, subscript in subscript
				continue
			} else {
				// subscript should be own token
				if tokenString[0] == '.' {
					tokenString = tokenString[1 : len(tokenString)-1]
				} else {
					tokenString = tokenString[:len(tokenString)-1]
				}

				tokens = append(tokens, tokenString[:])

				tokenString = "["
				continue
			}
		}

		if strings.Contains(tokenString, "[") {
			if quoteCount%2 == 1 || openScriptBracket != closeScriptBracket {
				// inside expression or quotes
				continue
			}

			if rne == ']' && openSubscriptBracket == closeSubscriptBracket {
				if tokenString[0] == '.' {
					tokenString = tokenString[1:]
				} else {
					tokenString = tokenString[:]
				}

				tokens = append(tokens, tokenString[:])

				tokenString = ""
				continue
			}
		} else if rne == '.' {
			if quoteCount%2 == 1 || openScriptBracket != closeScriptBracket {
				// inside expression or quotes
				continue
			}

			if tokenString[0] == '.' {
				tokenString = tokenString[1 : len(tokenString)-1]
			} else {
				tokenString = tokenString[:len(tokenString)-1]
			}

			tokens = append(tokens, tokenString[:])

			tokenString = "."
			continue
		}
	}

	// parse the last token
	if len(tokenString) > 0 {
		if tokenString[0] == '.' {
			tokenString = tokenString[1:]
		}

		tokens = append(tokens, tokenString[:])
	}

	return tokens, nil
}

// Parse will parse a single token string and return an actionable token
func Parse(tokenString string, engine script.Engine, options *option.QueryOptions) (Token, error) {
	isScript := func(token string) bool {
		return len(token) > 2 && strings.HasPrefix(token, "(") && strings.HasSuffix(token, ")")
	}

	isKey := func(token string) bool {
		isSingleQuoted := strings.HasPrefix(token, "'") && strings.HasSuffix(token, "'")
		isDoubleQuoted := strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"")
		return len(token) > 1 && (isSingleQuoted || isDoubleQuoted)
	}

	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, getInvalidTokenEmpty()
	}

	if tokenString == "$" {
		return newRootToken(), nil
	}
	if tokenString == "@" {
		return newCurrentToken(), nil
	}
	if tokenString == "*" {
		return newWildcardToken(), nil
	}
	if tokenString == ".." {
		return newRecursiveToken(), nil
	}

	if !strings.HasPrefix(tokenString, "[") {
		if tokenString == "length" {
			return newLengthToken(), nil
		}
		return newKeyToken(tokenString), nil
	}

	if !strings.HasSuffix(tokenString, "]") {
		return nil, getInvalidTokenFormatError(tokenString)
	}
	// subscript, or child operator

	subscript := strings.TrimSpace(tokenString[1 : len(tokenString)-1])
	if subscript == "" {
		return nil, getInvalidTokenFormatError(tokenString)
	}

	if subscript == "*" {
		// range all
		return newWildcardToken(), nil
	} else if strings.HasPrefix(subscript, "?") {
		// filter
		if !strings.HasPrefix(subscript, "?(") || !strings.HasSuffix(subscript, ")") {
			return nil, getInvalidTokenFormatError(tokenString)
		}
		return newFilterToken(strings.TrimSpace(subscript[2:len(subscript)-1]), engine, options)
	}

	// from this point we have the chance of things being nested or wrapped
	// which would result in the parsing being invalid

	openBracketCount, closeBracketCount := 0, 0
	openSingleQuote := false
	openDoubleQuote := false

	inQuotes := func() bool {
		return openSingleQuote || openDoubleQuote
	}

	args := []interface{}{}

	bufferString := ""
	for idx, rne := range subscript {
		bufferString += string(rne)
		switch rne {
		case ' ':
			if !inQuotes() && openBracketCount == closeBracketCount {
				// remove whitespace
				bufferString = strings.TrimSpace(bufferString)
			}
			break
		case '(':
			if inQuotes() {
				continue
			}
			openBracketCount++
			break
		case ')':
			closeBracketCount++

			if openBracketCount == closeBracketCount {
				// if we are closing bracket, add script to args
				script := bufferString[:]
				if !isScript(script) {
					return nil, getInvalidExpressionFormatError(script)
				}
				args = append(args, script)
				bufferString = ""
			}
			break
		case '\'':
			if openDoubleQuote || openBracketCount != closeBracketCount {
				continue
			}

			// if last token is escape character, then this is not an open or close
			if len(bufferString) > 1 && bufferString[len(bufferString)-2] == '\\' {
				bufferString = bufferString[0:len(bufferString)-2] + "'"
				break
			}

			openSingleQuote = !openSingleQuote

			if openSingleQuote {
				// open quote
				if bufferString != "'" {
					return nil, getInvalidTokenFormatError(tokenString)
				}
			} else {
				// close quote
				if !isKey(bufferString) {
					return nil, getInvalidTokenFormatError(tokenString)
				}
				args = append(args, bufferString[:])
				bufferString = ""
			}
			break
		case '"':
			if openSingleQuote || openBracketCount != closeBracketCount {
				continue
			}

			// if last token is escape character, then this is not an open or close
			if len(bufferString) > 1 && bufferString[len(bufferString)-2] == '\\' {
				bufferString = bufferString[0:len(bufferString)-2] + "\""
				break
			}

			openDoubleQuote = !openDoubleQuote

			if openDoubleQuote {
				// open quote
				if bufferString != "\"" {
					return nil, getInvalidTokenFormatError(tokenString)
				}
			} else {
				// close quote
				if !isKey(bufferString) {
					return nil, getInvalidTokenFormatError(tokenString)
				}
				args = append(args, bufferString[:])
				bufferString = ""
			}
			break
		case ':':
			if inQuotes() || (openBracketCount != closeBracketCount) {
				continue
			}
			if arg := bufferString[:len(bufferString)-1]; arg != "" {
				if num, err := strconv.ParseInt(arg, 10, 64); err == nil {
					args = append(args, num)
				} else {
					return nil, getInvalidTokenFormatError(tokenString)
				}
			} else if idx == 0 {
				// if the token starts with :
				args = append(args, nil)
			}
			args = append(args, ":")

			bufferString = ""
			break
		case ',':
			if inQuotes() || (openBracketCount != closeBracketCount) {
				continue
			}

			if arg := bufferString[:len(bufferString)-1]; arg != "" {
				if num, err := strconv.ParseInt(arg, 10, 64); err == nil {
					args = append(args, num)
				} else {
					args = append(args, arg)
				}
			}
			args = append(args, ",")

			bufferString = ""
			break
		}
	}

	if bufferString != "" {
		if num, err := strconv.ParseInt(bufferString, 10, 64); err == nil {
			args = append(args, num)
		} else {
			args = append(args, bufferString[:])
		}
	}

	if len(args) == 1 {
		// key, index, or script
		arg := args[0]
		if strArg, ok := arg.(string); ok {
			if isKey(strArg) {
				return newKeyToken(strArg[1 : len(strArg)-1]), nil
			} else if isScript(strArg) {
				return newScriptToken(strArg[1:len(strArg)-1], engine, options)
			}
		} else if intArg, ok := isInteger(arg); ok {
			return newIndexToken(intArg, options), nil
		}
		return nil, getInvalidTokenFormatError(tokenString)
	}

	// range or union
	colonCount := 0
	lastWasColon := false
	commaCount := 0

	// includesKeys := false
	justArgs := []interface{}{}

	for _, arg := range args {
		switch arg {
		case ":":
			colonCount++
			if lastWasColon {
				justArgs = append(justArgs, nil)
			}
			lastWasColon = true
			continue
		case ",":
			commaCount++
			break
		default:
			justArgs = append(justArgs, arg)
			break
		}
		lastWasColon = false
	}

	args = justArgs

	if colonCount > 0 && commaCount > 0 {
		return nil, getInvalidTokenFormatError(tokenString)
	} else if commaCount > 0 {
		// Union

		// we should always have one less comma than args
		if commaCount >= len(args) {
			return nil, getInvalidTokenFormatError(tokenString)
		}
		for idx, arg := range args {
			if strArg, ok := arg.(string); ok {
				if isScript(strArg) {
					var err error
					arg, err = newExpressionToken(strArg[1:len(strArg)-1], engine, options)
					if err != nil {
						return nil, getInvalidExpressionError(err)
					}
					args[idx] = arg
					continue
				} else if isKey(strArg) {
					args[idx] = strArg[1 : len(strArg)-1]
					continue
				}
			} else if intArg, ok := isInteger(arg); ok {
				args[idx] = intArg
				continue
			}
			return nil, getInvalidTokenFormatError(tokenString)
		}

		return newUnionToken(args, options), nil
	} else if colonCount > 0 {
		// Range
		if colonCount > 2 {
			return nil, getInvalidTokenFormatError(tokenString)
		}
		if colonCount == 1 && len(args) == 1 {
			// to help support [x:] tokens
			args = append(args, nil)
		}

		var from, to, step interface{} = args[0], args[1], nil
		if len(args) > 2 {
			step = args[2]
		}

		if strFrom, ok := from.(string); ok {
			if !isScript(strFrom) {
				return nil, getInvalidExpressionFormatError(strFrom)
			}
			var err error
			from, err = newExpressionToken(strFrom[1:len(strFrom)-1], engine, options)
			if err != nil {
				return nil, getInvalidExpressionError(err)
			}
		}
		if strTo, ok := to.(string); ok {
			if !isScript(strTo) {
				return nil, getInvalidExpressionFormatError(strTo)
			}
			var err error
			to, err = newExpressionToken(strTo[1:len(strTo)-1], engine, options)
			if err != nil {
				return nil, getInvalidExpressionError(err)
			}
		}
		if strStep, ok := step.(string); ok {
			if !isScript(strStep) {
				return nil, getInvalidExpressionFormatError(strStep)
			}
			var err error
			step, err = newExpressionToken(strStep[1:len(strStep)-1], engine, options)
			if err != nil {
				return nil, getInvalidExpressionError(err)
			}
		}

		return newRangeToken(from, to, step, options), nil
	}

	return nil, getInvalidTokenFormatError(tokenString)
}
