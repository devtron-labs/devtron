package blob_storage

import (
	"fmt"
	"os"
	"os/exec"
)

func setAWSEnvironmentVariables(s3Config *AwsS3BaseConfig, command *exec.Cmd) {
	if s3Config.AccessKey != "" && s3Config.Passkey != "" {
		command.Env = append(os.Environ(),
			fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", s3Config.AccessKey),
			fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", s3Config.Passkey),
		)
	}
}
