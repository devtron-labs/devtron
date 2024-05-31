/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package v1

const (
	Clone     Operation = "clone"
	Create    Operation = "create"
	Delete    Operation = "delete"
	Edit      Operation = "edit"
	Append    Operation = "append"
	Undefined Operation = "undefined"

	Automatic Trigger = "automatic"
	Manual    Trigger = "manual"

	Secret    string = "secret"
	ConfigMap string = "configmap"

	BranchFixed string = "BranchFixed"
	BranchRegex string = "BranchRegex"
	TagAny      string = "TagAny"
	Webhook     string = "Webhook"

	OperationUndefinedError     string = "undefined operator for %s"
	OperationUnimplementedError string = "unimplemented operator %s for %s"
	UnsupportedVersion          string = "unsupported version %s for %s"

	SourceNotSame         string = "source not same for %s and %s"
	DestinationNotSame    string = "destination not same for %s and %s"
	SourceNotUnique       string = "source not unique for %s"
	DestinationNotUnique  string = "destination not unique for %s"
	SourceDestinationSame string = "source and destination cannot be same for %s clone"

	NameEmpty        string = "name of %s cannot be empty"
	EnvironmentEmpty string = "environment cannot be empty for %s %s"
	DataEmpty        string = "data cannot be empty for %s %s"

	StagesMissing string = "task size cannot be zero"
	ScriptMissing string = "script is mandatory"
)
