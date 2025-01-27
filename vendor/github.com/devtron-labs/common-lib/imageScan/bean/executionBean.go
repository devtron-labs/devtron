package bean

type ScanExecutionMedium string

const (
	InHouse  ScanExecutionMedium = "in-house" // this contains all methods of hitting image scanner directly via rest or rpc
	External ScanExecutionMedium = "external" // if a scan tool is registered via api and execution is via plugin steps
)

func (e ScanExecutionMedium) IsScanExecutionMediumInHouse() bool {
	return e == InHouse
}

func (e ScanExecutionMedium) IsScanMediumExternal() bool {
	return e == External
}
