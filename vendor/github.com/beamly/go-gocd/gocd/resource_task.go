package gocd

import "errors"

// Validate each of the possible task types.
func (t *Task) Validate() error {
	switch t.Type {
	case "":
		return errors.New("Missing `gocd.TaskAttribute` type")
	case "exec":
		return t.Attributes.ValidateExec()
	case "ant":
		return t.Attributes.ValidateAnt()
	default:
		return errors.New("Unexpected `gocd.Task.Attribute` types")
	}
}
