package bean

type UserTerminalSessionRequest struct {
	Id        int
	UserId    int32
	ClusterId int
	NodeName  string
	BaseImage string
	ShellName string
}

type UserTerminalSessionConfig struct {
	MaxSessionPerUser int
}

type UserTerminalSessionResponse struct {
	UserTerminalSessionId int
	UserId                int32
	TerminalAccessId      int
	ShellName             string
	Status                TerminalPodStatus
}

const TerminalAccessPodNameTemplate = "terminal-access-" + TerminalAccessClusterIdTemplateVar + "-" + TerminalAccessUserIdTemplateVar + "-" + TerminalAccessRandomIdVar
const TerminalAccessClusterIdTemplateVar = "${cluster_id}"
const TerminalAccessUserIdTemplateVar = "${user_id}"
const TerminalAccessRandomIdVar = "${random_id}"
const TerminalAccessPodNameVar = "${pod_name}"
const TerminalAccessPodTemplateName = "terminal-access-pod-template"

type TerminalPodStatus string

const (
	Starting   TerminalPodStatus = "Starting"
	Running    TerminalPodStatus = "Running"
	Terminated TerminalPodStatus = "Terminated"
	Error      TerminalPodStatus = "Error"
)
