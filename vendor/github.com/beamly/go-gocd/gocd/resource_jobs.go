package gocd

import (
	"encoding/json"
	"errors"
	"strconv"
)

// JSONString returns a string of this stage as a JSON object.
func (j *Job) JSONString() (body string, err error) {
	err = j.Validate()
	if err != nil {
		return
	}

	bdy, err := json.MarshalIndent(j, "", "  ")
	body = string(bdy)

	return
}

// Validate a job structure has non-nil values on correct attributes
func (j *Job) Validate() (err error) {
	if j.Name == "" {
		err = errors.New("`gocd.Jobs.Name` is empty")
	}
	return
}

// UnmarshalJSON and handle "never", "null", and integers.
func (tf *TimeoutField) UnmarshalJSON(b []byte) (err error) {
	// The target `valint` value has -1 so that we have a value which does not match
	// the expected value (probably) and we can ensure the methods is working and not just putting
	// the default in (which is `0`).
	valInt := -1

	value := string(b)

	if value == `"never"` || value == `"null"` || value == "never" || value == "null" {
		valInt = 0
	} else {
		valInt, err = strconv.Atoi(value)

	}
	*tf = TimeoutField(valInt)

	return
}

// MarshalJSON of TimeoutField into a string
func (tf TimeoutField) MarshalJSON() (b []byte, err error) {
	return []byte(strconv.Itoa(int(tf))), nil
}
