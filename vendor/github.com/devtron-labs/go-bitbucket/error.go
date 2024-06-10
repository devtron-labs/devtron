package bitbucket

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type BitbucketError struct {
	Message string
	Fields  map[string][]string
}

func DecodeError(e map[string]interface{}) error {
	var bitbucketError BitbucketError
	err := mapstructure.Decode(e["error"], &bitbucketError)
	if err != nil {
		return err
	}

	return errors.New(bitbucketError.Message)
}

// UnexpectedResponseStatusError represents an unexpected status code
// returned from the API, along with the body, if it could be read. If the body
// could not be read, the body contains the error message trying to read it.
type UnexpectedResponseStatusError struct {
	Status string
	Body   []byte
}

func (e *UnexpectedResponseStatusError) Error() string {
	return e.Status
}

// ErrorWithBody returns an error with the given status and body.
func (e *UnexpectedResponseStatusError) ErrorWithBody() error {
	return fmt.Errorf("unexpected status %s, body: %s", e.Status, string(e.Body))
}
