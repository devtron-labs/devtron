package gocd

import (
	"encoding/json"
	"errors"
)

// JSONString returns a string of this stage as a JSON object.
func (s *StageInstance) JSONString() (string, error) {
	err := s.Validate()
	if err != nil {
		return "", err
	}
	bdy, err := json.MarshalIndent(s, "", "  ")
	return string(bdy), err
}

// Validate ensures the attributes attached to this structure are ready for submission to the GoCD API.
func (s *StageInstance) Validate() error {
	if s.Name == "" {
		return errors.New("`gocd.StageInstance.Name` is empty")
	}

	if len(s.Jobs) == 0 {
		return errors.New("At least one `Job` must be specified")
	}

	for _, job := range s.Jobs {
		if err := job.Validate(); err != nil {
			return err
		}
	}

	return nil
}
