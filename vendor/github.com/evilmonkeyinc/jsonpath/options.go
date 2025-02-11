package jsonpath

import (
	"github.com/evilmonkeyinc/jsonpath/option"
	"github.com/evilmonkeyinc/jsonpath/script"
)

// OptionFunction function that can be used as a compile or query option
type OptionFunction func(selector *Selector) error

// Apply the option to the Selector
func (fn OptionFunction) Apply(selector *Selector) error {
	return fn(selector)
}

// Option allows to set compile and query options when calling Compile
type Option interface {
	// Apply the option to the Selector
	Apply(selector *Selector) error
}

// ScriptEngine allows you to set a custom script Engine for the JSONPath selector
func ScriptEngine(engine script.Engine) Option {
	return OptionFunction(func(selector *Selector) error {
		if selector.engine == nil {
			selector.engine = engine
		}
		return nil
	})
}

// QueryOptions allows you to set the query options for the JSONPath selector
func QueryOptions(options *option.QueryOptions) Option {
	return OptionFunction(func(selector *Selector) error {
		if selector.Options == nil {
			selector.Options = options
			return nil
		}
		return errOptionAlreadySet
	})
}
