package CiPipeline

type CiFailReason string

func (r CiFailReason) String() string {
	return string(r)
}

const CiFailed CiFailReason = "CI Failed: exit code 1"

const CiStageFailErrorCode = 2
