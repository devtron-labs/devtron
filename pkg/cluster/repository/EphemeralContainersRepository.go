package repository

import "time"

type EphemeralContainer struct {
	tableName           struct{} `sql:"ephemeral_container" pg:",discard_unknown_columns"`
	Id                  int      `sql:"id,pk"`
	Name                string   `sql:"name"`
	ClusterID           int      `sql:"cluster_id"`
	Namespace           string   `sql:"namespace"`
	PodName             string   `sql:"pod_name"`
	TargetContainer     string   `sql:"target_container"`
	Config              string   `sql:"config"`
	IsExternallyCreated bool     `sql:"is_externally_created"`
}

type EphemeralContainerAction struct {
	tableName            struct{}  `sql:"ephemeral_container_actions" pg:",discard_unknown_columns"`
	Id                   int       `sql:"id,pk"`
	EphemeralContainerID int       `sql:"ephemeral_container_id"`
	ActionType           int       `sql:"action_type"`
	PerformedBy          int       `sql:"performed_by"`
	PerformedAt          time.Time `sql:"performed_at"`
}
