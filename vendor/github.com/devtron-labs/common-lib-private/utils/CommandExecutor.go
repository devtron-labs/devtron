package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

func RunCommand(cmd *exec.Cmd) error {
	var stdBuffer bytes.Buffer
	mw := io.MultiWriter(os.Stdout, &stdBuffer)
	cmd.Stdout = mw
	cmd.Stderr = mw
	if err := cmd.Run(); err != nil {
		return err
	}
	//log.Println(stdBuffer.String())
	return nil
}
