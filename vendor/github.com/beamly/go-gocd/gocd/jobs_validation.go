package gocd

import "errors"

// ValidateExec checks that the specified values for the Task struct are correct for a cli exec task
func (t *TaskAttributes) ValidateExec() error {
	if len(t.RunIf) == 0 {
		return errors.New("'run_if' must not be empty")
	}
	if t.Command == "" {
		return errors.New("'command' must not be empty")
	}
	if len(t.Arguments) == 0 {
		return errors.New("'arguments' must not be empty")
	}
	if t.WorkingDirectory == "" {
		return errors.New("'working_directory' must not empty")
	}

	return nil
}

// ValidateAnt checks that the specified values for the Task struct are correct for a an Ant task
func (t *TaskAttributes) ValidateAnt() error {
	// Ensure valid attributes are set.
	if len(t.RunIf) == 0 {
		return errors.New("'run_if' must not be empty")
	}
	if t.BuildFile == "" {
		return errors.New("'build_file' must not be empty")
	}
	if t.Target == "" {
		return errors.New("'target' must not be empty")
	}
	if t.WorkingDirectory == "" {
		return errors.New("'working_directory' must not empty")
	}

	// Ensure extra attributes are not set
	// @TODO Rewrite this as a generator.
	//if t.Command != "" {
	//	return errors.New("'Command' can not be set for 'Ant' tasks")
	//}
	//if t.Arguments != nil || len(t.Arguments) > 0 {
	//	return errors.New("'Command' can not be set for 'Ant' tasks")
	//}
	//if t.BuildFile != "" {
	//	return errors.New("'BuildFile' can not be set for 'Ant' tasks")
	//}
	//if t.Target != "" {
	//	return errors.New("'Target' can not be set for 'Ant' tasks")
	//}
	//if t.NantPath != "" {
	//	return errors.New("'NantPath' can not be set for 'Ant' tasks")
	//}
	//if t.Pipeline != "" {
	//	return errors.New("'Pipeline' can not be set for 'Ant' tasks")
	//}
	//if t.Stage != "" {
	//	return errors.New("'Stage' can not be set for 'Ant' tasks")
	//}
	//if t.Job != "" {
	//	return errors.New("'Job' can not be set for 'Ant' tasks")
	//}
	//if t.Source != "" {
	//	return errors.New("'Source' can not be set for 'Ant' tasks")
	//}
	//if t.IsSourceAFile != "" {
	//	return errors.New("'IsSourceAFile' can not be set for 'Ant' tasks")
	//}
	//if t.Destination != "" {
	//	return errors.New("'Destination' can not be set for 'Ant' tasks")
	//}
	//if t.PluginConfiguration != nil {
	//	return errors.New("'PluginConfiguration' can not be set for 'Ant' tasks")
	//}
	//if t.Configuration != nil {
	//	return errors.New("'Configuration' can not be set for 'Ant' tasks")
	//}

	return nil
}

// ValidateNant checks that the specified values for the Task struct are correct for a a Nant task
//func (t *TaskAttributes) ValidateNant() error {
//	return errors.New("Not Implemented")
//}

// ValidateRake checks that the specified values for the Task struct are correct for a a Rake task
//func (t *TaskAttributes) ValidateRake() error {
//	return errors.New("Not Implemented")
//}

// ValidateRake checks that the specified values for the Task struct are correct for a a Rake task
//func (t *TaskAttributes) ValidateFetch() error {
//	return errors.New("Not Implemented")
//}

// ValidatePluggableTask checks that the specified values for the Task struct are correct for a a Plugin task
//func (t *TaskAttributes) ValidatePluggableTask() error {
//	return errors.New("Not Implemented")
//}
