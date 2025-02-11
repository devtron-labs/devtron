package jsonpath

import (
	"github.com/evilmonkeyinc/jsonpath/script/standard"
	"github.com/evilmonkeyinc/jsonpath/token"
)

// Compile will compile the JSONPath selector
func Compile(selector string, options ...Option) (*Selector, error) {
	jsonPath := &Selector{
		selector: selector,
	}

	for _, option := range options {
		if err := option.Apply(jsonPath); err != nil {
			return nil, err
		}
	}

	// Set defaults if options were not used
	if jsonPath.engine == nil {
		jsonPath.engine = new(standard.ScriptEngine)
	}

	tokenStrings, err := token.Tokenize(jsonPath.selector)
	if err != nil {
		return nil, getInvalidJSONPathSelectorWithReason(selector, err)
	}

	tokens := make([]token.Token, len(tokenStrings))
	for idx, tokenString := range tokenStrings {
		token, err := token.Parse(tokenString, jsonPath.engine, jsonPath.Options)
		if err != nil {
			return nil, getInvalidJSONPathSelectorWithReason(selector, err)
		}
		tokens[idx] = token
	}
	jsonPath.tokens = tokens

	return jsonPath, nil
}

// Query will return the result of the JSONPath selector applied against the specified JSON data.
func Query(selector string, jsonData interface{}, options ...Option) (interface{}, error) {
	jsonPath, err := Compile(selector, options...)
	if err != nil {
		return nil, getInvalidJSONPathSelectorWithReason(selector, err)
	}
	return jsonPath.Query(jsonData)
}

// QueryString will return the result of the JSONPath selector applied against the specified JSON data.
func QueryString(selector string, jsonData string, options ...Option) (interface{}, error) {
	jsonPath, err := Compile(selector, options...)
	if err != nil {
		return nil, getInvalidJSONPathSelectorWithReason(selector, err)
	}
	return jsonPath.QueryString(jsonData)
}
