package chartGroup

// / bean for v2
type ChartGroupInstallRequest struct {
	ProjectId                     int                              `json:"projectId"  validate:"required,number"`
	ChartGroupInstallChartRequest []*ChartGroupInstallChartRequest `json:"charts" validate:"dive,required"`
	ChartGroupId                  int                              `json:"chartGroupId"` //optional
	UserId                        int32                            `json:"-"`
}

type ChartGroupInstallChartRequest struct {
	AppName            string `json:"appName,omitempty"  validate:"name-component,max=100" `
	EnvironmentId      int    `json:"environmentId,omitempty" validate:"required,number" `
	AppStoreVersion    int    `json:"appStoreVersion,omitempty,notnull" validate:"required,number" `
	ValuesOverrideYaml string `json:"valuesOverrideYaml,omitempty"` //optional
	ReferenceValueId   int    `json:"referenceValueId, omitempty" validate:"required,number"`
	ReferenceValueKind string `json:"referenceValueKind, omitempty" validate:"oneof=DEFAULT TEMPLATE DEPLOYED"`
	ChartGroupEntryId  int    `json:"chartGroupEntryId"` //optional
}
type ChartGroupInstallMetadata struct {
	AppName       string `json:"appName"`
	EnvironmentId int    `json:"environmentId"`
	TriggerStatus string `json:"triggerStatus"`
	Reason        string `json:"reason"`
}
type ChartGroupInstallAppRes struct {
	ChartGroupInstallMetadata []ChartGroupInstallMetadata `json:"chartGroupInstallMetadata"`
	Summary                   string                      `json:"summary"`
}
type TriggerStatus string
type Reason string

const FAILED_TO_TRIGGER = "%d/%d failed to trigger"
const (
	StatusFailed       TriggerStatus = "failed"
	StatusSuccess      TriggerStatus = "success"
	ReasonNotAuthorize Reason        = "not authorized"
	ReasonTriggered    Reason        = "triggered"
)
