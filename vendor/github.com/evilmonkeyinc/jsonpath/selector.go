package jsonpath

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/evilmonkeyinc/jsonpath/option"
	"github.com/evilmonkeyinc/jsonpath/script"
	"github.com/evilmonkeyinc/jsonpath/token"
)

// Selector represents a compiled JSONPath selector
// and exposes functions to query JSON data and objects.
type Selector struct {
	Options  *option.QueryOptions
	engine   script.Engine
	tokens   []token.Token
	selector string
}

// String returns the compiled selector string representation
func (query *Selector) String() string {
	jsonPath := ""
	for _, token := range query.tokens {
		jsonPath += fmt.Sprintf("%s", token)
	}
	return jsonPath
}

// Query will return the result of the JSONPath query applied against the specified JSON data.
func (query *Selector) Query(root interface{}) (interface{}, error) {
	if len(query.tokens) == 0 {
		return nil, getInvalidJSONPathSelector(query.selector)
	}

	tokens := make([]token.Token, 0)
	if len(query.tokens) > 1 {
		tokens = query.tokens[1:]
	}

	found, err := query.tokens[0].Apply(root, root, tokens)
	if err != nil {
		return nil, err
	}
	return found, nil
}

// QueryString will return the result of the JSONPath query applied against the specified JSON data.
func (query *Selector) QueryString(jsonData string) (interface{}, error) {
	jsonData = strings.TrimSpace(jsonData)
	if jsonData == "" {
		return nil, getInvalidJSONData(errDataIsUnexpectedTypeOrNil)
	}

	var root interface{}

	if strings.HasPrefix(jsonData, "{") && strings.HasSuffix(jsonData, "}") {
		// object
		root = make(map[string]interface{})
		if err := json.Unmarshal([]byte(jsonData), &root); err != nil {
			return nil, getInvalidJSONData(err)
		}
	} else if strings.HasPrefix(jsonData, "[") && strings.HasSuffix(jsonData, "]") {
		// array
		root = make([]interface{}, 0)
		if err := json.Unmarshal([]byte(jsonData), &root); err != nil {
			return nil, getInvalidJSONData(err)
		}
	} else if len(jsonData) > 2 && strings.HasPrefix(jsonData, "\"") && strings.HasPrefix(jsonData, "\"") {
		// string
		root = jsonData[1 : len(jsonData)-1]
	} else if strings.ToLower(jsonData) == "true" {
		// bool true
		root = true
	} else if strings.ToLower(jsonData) == "false" {
		// bool false
		root = false
	} else if val, err := strconv.ParseInt(jsonData, 10, 64); err == nil {
		// integer
		root = val
	} else if val, err := strconv.ParseFloat(jsonData, 64); err == nil {
		// float
		root = val
	} else {
		return nil, getInvalidJSONData(errDataIsUnexpectedTypeOrNil)
	}

	return query.Query(root)
}
