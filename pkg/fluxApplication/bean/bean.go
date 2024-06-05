package bean

import (
	"github.com/devtron-labs/devtron/api/helm-app/gRPC"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	FluxKustomizationGroup   = "kustomize.toolkit.fluxcd.io"
	FluxAppKustomizationKind = "Kustomization"
	FluxKustomizationVersion = "v1"
	AllNamespaces            = ""
	FluxHelmReleaseGroup     = "helm.toolkit.fluxcd.io"
	FluxAppHelmreleaseKind   = "HelmRelease"
	FluxHelmReleaseVersion   = "v2"

	//DevtronCDNamespae            = "devtroncd"
	//ArgoLabelForManagedResources = "app.kubernetes.io/instance"
)

const (
	Destination string = "destination"
	Server      string = "server"
	STATUS      string = "status"
	INVENTORY   string = "inventory"
	ENTRIES     string = "entries"
	ID          string = "id"
	VERSION     string = "v"
)
const (
	FieldSeparator  = "_"
	ColonTranscoded = "__"
)

type FluxApplicationListDto struct {
	ClusterId   int    `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	FluxAppDto  []*FluxApplicationDto
}
type FluxApplicationDto struct {
	Name               string             `json:"appName"`
	HealthStatus       string             `json:"appStatus"`
	SyncStatus         string             `json:"syncStatus"`
	EnvironmentDetails *EnvironmentDetail `json:"environmentDetail"`
	IsKustomizeApp     bool               `json:"isKustomizeApp"`
}
type EnvironmentDetail struct {
	ClusterId   int    `json:"clusterId"`
	ClusterName string `json:"clusterName"`
	Namespace   string `json:"namespace"`
}

var GvkForKustomizationFluxApp = schema.GroupVersionKind{
	Group:   FluxKustomizationGroup,
	Kind:    FluxAppKustomizationKind,
	Version: FluxKustomizationVersion,
}

var GvkForhelmreleaseFluxApp = schema.GroupVersionKind{
	Group:   FluxHelmReleaseGroup,
	Kind:    FluxAppHelmreleaseKind,
	Version: FluxHelmReleaseVersion,
}

type FluxpplicationDetailDto struct {
	*FluxApplicationListDto
	ResourceTree *gRPC.ResourceTreeResponse `json:"resourceTree"`
	Manifest     map[string]interface{}     `json:"manifest"`
}

type FluxManagedResource struct {
	Group     string
	Kind      string
	Version   string
	Name      string
	Namespace string
}

type ObjMetadata struct {
	Namespace string
	Name      string
	GroupKind schema.GroupKind
}

type ObjectMetadataCompact struct {
	Id      string
	Version string
}
