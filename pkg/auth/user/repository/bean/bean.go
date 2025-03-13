package bean

type RoleModelFieldsDto struct {
	Entity,
	Team,
	App,
	Env,
	AccessType,
	Cluster,
	Namespace,
	Group,
	Kind,
	Resource string
	Action    string
	OldValues bool
	Workflow  string
}
