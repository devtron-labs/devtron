/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package blob_storage

import (
	"fmt"
	"os"
	"os/exec"
)

const (
	WhenSupported = "when_supported"
	WhenRequired  = "when_required"
)

func setAWSEnvironmentVariables(s3Config *AwsS3BaseConfig, command *exec.Cmd) {
	if s3Config.AccessKey != "" && s3Config.Passkey != "" {
		command.Env = os.Environ()
		command.Env = append(command.Env,
			fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3Config.AccessKey),
			fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Config.Passkey),
		)
	}
	if s3Config.EndpointUrl != "" {
		command.Env = append(command.Env,
			// The below is required for https://github.com/aws/aws-cli/issues/9214
			// This is only required for secure endpoints only https://github.com/boto/boto3/issues/4398#issuecomment-2712259341
			fmt.Sprintf("AWS_REQUEST_CHECKSUM_CALCULATION=%s", WhenRequired),
			fmt.Sprintf("AWS_RESPONSE_CHECKSUM_VALIDATION=%s", WhenRequired),
		)
	}
}
