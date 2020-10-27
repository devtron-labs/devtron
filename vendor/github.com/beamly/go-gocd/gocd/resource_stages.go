package gocd

import (
	"encoding/json"
	"errors"
)

// JSONString returns a string of this stage as a JSON object.
func (s *Stage) JSONString() (string, error) {
	err := s.Validate()
	if err != nil {
		return "", err
	}

	if s.Approval.Authorization == nil {
		s.Approval.Authorization = &Authorization{}
	}

	if s.Approval.Authorization.Users == nil {
		s.Approval.Authorization.Users = []string{}
	}
	if s.Approval.Authorization.Roles == nil {
		s.Approval.Authorization.Roles = []string{}
	}

	bdy, err := json.MarshalIndent(s, "", "  ")
	return string(bdy), err
}

// Validate ensures the attributes attached to this structure are ready for submission to the GoCD API.
func (s *Stage) Validate() error {
	if s.Name == "" {
		return errors.New("`gocd.Stage.Name` is empty")
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
