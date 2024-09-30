package workFlow

import "path"

type CiFailReason string

type CdFailReason string

func (r CiFailReason) String() string {
	return string(r)
}

func (r CdFailReason) String() string {
	return string(r)
}

const (
	PreCiFailed  CiFailReason = "Pre-CI task failed: %s"
	PostCiFailed CiFailReason = "Post-CI task failed: %s"
	BuildFailed  CiFailReason = "Docker build failed"
	PushFailed   CiFailReason = "Docker push failed"
	ScanFailed   CiFailReason = "Image scan failed"
	CiFailed     CiFailReason = "CI failed"

	CdStageTaskFailed CdFailReason = "%s task failed: %s"
	CdStageFailed     CdFailReason = "%s failed. Reason: %s"
)

const (
	TerminalLogDir  = "/dev"
	TerminalLogFile = "/termination-log"
)

func GetTerminalLogFilePath() string {
	return path.Join(TerminalLogDir, TerminalLogFile)
}

const (
	DefaultErrorCode     = 1
	AbortErrorCode       = 143
	CiStageFailErrorCode = 2
)
