/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
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
