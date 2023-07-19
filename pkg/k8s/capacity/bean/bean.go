package bean

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

const (
	LabelNodeRolePrefix = "node-role.kubernetes.io/"
	NodeLabelRole       = "kubernetes.io/role"
	Kibibyte            = 1024
	Mebibyte            = 1024 * 1024
	Gibibyte            = 1024 * 1024 * 1024
	kilobyte            = 1000
	Megabyte            = 1000 * 1000
	Gigabyte            = 1000 * 1000 * 1000
)

// below const set is used for pod filters
const (
	daemonSetFatal       = "DaemonSet-managed Pods (use --ignore-daemonsets to ignore)"
	daemonSetWarning     = "ignoring DaemonSet-managed Pods"
	localStorageFatal    = "Pods with local storage (use --delete-emptydir-data to override)"
	localStorageWarning  = "deleting Pods with local storage"
	unmanagedFatal       = "Pods declare no controller (use --force to override)"
	unmanagedWarning     = "deleting Pods that declare no controller"
	AWSNodeGroupLabel    = "alpha.eksctl.io/nodegroup-name"
	AzureNodeGroupLabel  = "kubernetes.azure.com/agentpool"
	GcpNodeGroupLabel    = "cloud.google.com/gke-nodepool"
	KopsNodeGroupLabel   = "kops.k8s.io/instancegroup"
	AWSEKSNodeGroupLabel = "eks.amazonaws.com/nodegroup"
)

// TODO: add any new nodeGrouplabel in this array
var NodeGroupLabels = [5]string{AWSNodeGroupLabel, AzureNodeGroupLabel, GcpNodeGroupLabel, KopsNodeGroupLabel, AWSEKSNodeGroupLabel}

// below const set is used for pod delete status
const (
	// PodDeleteStatusTypeOkay is "Okay"
	PodDeleteStatusTypeOkay = "Okay"
	// PodDeleteStatusTypeSkip is "Skip"
	PodDeleteStatusTypeSkip = "Skip"
	// PodDeleteStatusTypeWarning is "Warning"
	PodDeleteStatusTypeWarning = "Warning"
	// PodDeleteStatusTypeError is "Error"
	PodDeleteStatusTypeError = "Error"
)

type ClusterCapacityDetail struct {
	Id                int                                   `json:"id,omitempty"`
	Name              string                                `json:"name,omitempty"`
	ErrorInConnection string                                `json:"errorInNodeListing,omitempty"`
	NodeCount         int                                   `json:"nodeCount,omitempty"`
	NodeDetails       []NodeDetails                         `json:"nodeDetails"`
	NodeErrors        map[corev1.NodeConditionType][]string `json:"nodeErrors"`
	NodeK8sVersions   []string                              `json:"nodeK8sVersions"`
	ServerVersion     string                                `json:"serverVersion,omitempty"`
	Cpu               *ResourceDetailObject                 `json:"cpu"`
	Memory            *ResourceDetailObject                 `json:"memory"`
	IsVirtualCluster  bool                                  `json:"isVirtualCluster"`
}

type NodeCapacityDetail struct {
	Name          string                              `json:"name"`
	Version       string                              `json:"version,omitempty"`
	Kind          string                              `json:"kind,omitempty"`
	Roles         []string                            `json:"roles"`
	K8sVersion    string                              `json:"k8sVersion"`
	Cpu           *ResourceDetailObject               `json:"cpu,omitempty"`
	Memory        *ResourceDetailObject               `json:"memory,omitempty"`
	Age           string                              `json:"age,omitempty"`
	Status        string                              `json:"status,omitempty"`
	PodCount      int                                 `json:"podCount,omitempty"`
	Errors        map[corev1.NodeConditionType]string `json:"errors"`
	InternalIp    string                              `json:"internalIp"`
	ExternalIp    string                              `json:"externalIp"`
	Unschedulable bool                                `json:"unschedulable"`
	CreatedAt     string                              `json:"createdAt"`
	Labels        []*LabelAnnotationTaintObject       `json:"labels,omitempty"`
	Annotations   []*LabelAnnotationTaintObject       `json:"annotations,omitempty"`
	Taints        []*LabelAnnotationTaintObject       `json:"taints,omitempty"`
	Conditions    []*NodeConditionObject              `json:"conditions,omitempty"`
	Resources     []*ResourceDetailObject             `json:"resources,omitempty"`
	Pods          []*PodCapacityDetail                `json:"pods,omitempty"`
	Manifest      unstructured.Unstructured           `json:"manifest,omitempty"`
	ClusterName   string                              `json:"clusterName,omitempty"`
	NodeGroup     string                              `json:"nodeGroup"`
}

type PodCapacityDetail struct {
	Name      string                `json:"name"`
	Namespace string                `json:"namespace"`
	Cpu       *ResourceDetailObject `json:"cpu"`
	Memory    *ResourceDetailObject `json:"memory"`
	Age       string                `json:"age"`
	CreatedAt string                `json:"createdAt"`
}

type ResourceDetailObject struct {
	ResourceName      string `json:"name,omitempty"`
	Capacity          string `json:"capacity,omitempty"`
	Allocatable       string `json:"allocatable,omitempty"`
	Usage             string `json:"usage,omitempty"`
	Request           string `json:"request,omitempty"`
	Limit             string `json:"limit,omitempty"`
	UsagePercentage   string `json:"usagePercentage,omitempty"`
	RequestPercentage string `json:"requestPercentage,omitempty"`
	LimitPercentage   string `json:"limitPercentage,omitempty"`
	//below fields to be used at FE for sorting
	CapacityInBytes    int64 `json:"capacityInBytes,omitempty"`
	AllocatableInBytes int64 `json:"allocatableInBytes,omitempty"`
	UsageInBytes       int64 `json:"usageInBytes,omitempty"`
	RequestInBytes     int64 `json:"requestInBytes,omitempty"`
	LimitInBytes       int64 `json:"limitInBytes,omitempty"`
}

type LabelAnnotationTaintObject struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect,omitempty"`
}

type NodeConditionObject struct {
	Type      string `json:"type"`
	HaveIssue bool   `json:"haveIssue"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
}

type NodeUpdateRequestDto struct {
	ClusterId        int               `json:"clusterId"`
	Name             string            `json:"name"`
	ManifestPatch    string            `json:"manifestPatch"`
	Version          string            `json:"version"`
	Kind             string            `json:"kind"`
	Taints           []corev1.Taint    `json:"taints"`
	NodeCordonHelper *NodeCordonHelper `json:"nodeCordonOptions"`
	NodeDrainHelper  *NodeDrainHelper  `json:"nodeDrainOptions"`
}

type NodeCordonHelper struct {
	UnschedulableDesired bool `json:"unschedulableDesired"`
}

type NodeDrainHelper struct {
	Force              bool `json:"force"`
	DeleteEmptyDirData bool `json:"deleteEmptyDirData"`
	// GracePeriodSeconds is how long to wait for a pod to terminate.
	// IMPORTANT: 0 means "delete immediately"; set to a negative value
	// to use the pod's terminationGracePeriodSeconds.
	GracePeriodSeconds  int  `json:"gracePeriodSeconds"`
	IgnoreAllDaemonSets bool `json:"ignoreAllDaemonSets"`
	// DisableEviction forces drain to use delete rather than evict
	DisableEviction bool `json:"disableEviction"`
	K8sClientSet    *kubernetes.Clientset
}

type NodeDetails struct {
	NodeName  string                        `json:"nodeName"`
	NodeGroup string                        `json:"nodeGroup"`
	Taints    []*LabelAnnotationTaintObject `json:"taints"`
}

// PodDelete informs filtering logic whether a pod should be deleted or not
type PodDelete struct {
	Pod    corev1.Pod
	Status PodDeleteStatus
}

// PodDeleteList is a wrapper around []PodDelete
type PodDeleteList struct {
	items []PodDelete
}

// Pods returns a list of all pods marked for deletion after filtering.
func (l *PodDeleteList) Pods() []corev1.Pod {
	pods := []corev1.Pod{}
	for _, i := range l.items {
		if i.Status.Delete {
			pods = append(pods, i.Pod)
		}
	}
	return pods
}

func (l *PodDeleteList) Errors() []error {
	failedPods := make(map[string][]string)
	for _, i := range l.items {
		if i.Status.Reason == PodDeleteStatusTypeError {
			msg := i.Status.Message
			if msg == "" {
				msg = "unexpected error"
			}
			failedPods[msg] = append(failedPods[msg], fmt.Sprintf("%s/%s", i.Pod.Namespace, i.Pod.Name))
		}
	}
	errs := make([]error, 0, len(failedPods))
	for msg, pods := range failedPods {
		errs = append(errs, fmt.Errorf("cannot delete %s: %s", msg, strings.Join(pods, ", ")))
	}
	return errs
}

// PodDeleteStatus informs filters if a pod should be deleted
type PodDeleteStatus struct {
	Delete  bool
	Reason  string
	Message string
}

// PodFilter takes a pod and returns a PodDeleteStatus
type PodFilter func(corev1.Pod) PodDeleteStatus

func FilterPods(podList *corev1.PodList, filters []PodFilter) *PodDeleteList {
	pods := []PodDelete{}
	for _, pod := range podList.Items {
		var status PodDeleteStatus
		for _, filter := range filters {
			status = filter(pod)
			if !status.Delete {
				// short-circuit as soon as pod is filtered out
				// at that point, there is no reason to run pod
				// through any additional filters
				break
			}
		}
		// Add the pod to PodDeleteList no matter what PodDeleteStatus is,
		// those pods whose PodDeleteStatus is false like DaemonSet will
		// be catched by list.errors()
		pod.Kind = "Pod"
		pod.APIVersion = "v1"
		pods = append(pods, PodDelete{
			Pod:    pod,
			Status: status,
		})
	}
	list := &PodDeleteList{items: pods}
	return list
}

func (f *NodeDrainHelper) MakeFilters() []PodFilter {
	baseFilters := []PodFilter{
		f.skipDeletedFilter,
		f.daemonSetFilter,
		f.mirrorPodFilter,
		f.localStorageFilter,
		f.unreplicatedFilter,
	}
	return baseFilters
}

func (f *NodeDrainHelper) mirrorPodFilter(pod corev1.Pod) PodDeleteStatus {
	if _, found := pod.ObjectMeta.Annotations[corev1.MirrorPodAnnotationKey]; found {
		return MakePodDeleteStatusSkip()
	}
	return MakePodDeleteStatusOkay()
}

func (f *NodeDrainHelper) localStorageFilter(pod corev1.Pod) PodDeleteStatus {
	if !hasLocalStorage(pod) {
		return MakePodDeleteStatusOkay()
	}
	// Any finished pod can be removed.
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return MakePodDeleteStatusOkay()
	}
	if !f.DeleteEmptyDirData {
		return MakePodDeleteStatusWithError(localStorageFatal)
	}

	// TODO: this warning gets dropped by subsequent filters;
	// consider accounting for multiple warning conditions or at least
	// preserving the last warning message.
	return MakePodDeleteStatusWithWarning(true, localStorageWarning)
}

func (f *NodeDrainHelper) unreplicatedFilter(pod corev1.Pod) PodDeleteStatus {
	// any finished pod can be removed
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return MakePodDeleteStatusOkay()
	}

	controllerRef := v1.GetControllerOf(&pod)
	if controllerRef != nil {
		return MakePodDeleteStatusOkay()
	}
	if f.Force {
		return MakePodDeleteStatusWithWarning(true, unmanagedWarning)
	}
	return MakePodDeleteStatusWithError(unmanagedFatal)
}

func (f *NodeDrainHelper) daemonSetFilter(pod corev1.Pod) PodDeleteStatus {
	// Note that we return false in cases where the pod is DaemonSet managed,
	// regardless of flags.
	//
	// The exception is for pods that are orphaned (the referencing
	// management resource - including DaemonSet - is not found).
	// Such pods will be deleted if --force is used.
	controllerRef := v1.GetControllerOf(&pod)
	if controllerRef == nil || controllerRef.Kind != v1.SchemeGroupVersion.WithKind("DaemonSet").Kind {
		return MakePodDeleteStatusOkay()
	}
	// Any finished pod can be removed.
	if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
		return MakePodDeleteStatusOkay()
	}

	if _, err := f.K8sClientSet.AppsV1().DaemonSets(pod.Namespace).Get(context.TODO(), controllerRef.Name, v1.GetOptions{}); err != nil {
		// remove orphaned pods with a warning if --force is used
		if apierrors.IsNotFound(err) && f.Force {
			return MakePodDeleteStatusWithWarning(true, err.Error())
		}

		return MakePodDeleteStatusWithError(err.Error())
	}

	if !f.IgnoreAllDaemonSets {
		return MakePodDeleteStatusWithError(daemonSetFatal)
	}
	return MakePodDeleteStatusWithWarning(false, daemonSetWarning)
}

func (f *NodeDrainHelper) skipDeletedFilter(pod corev1.Pod) PodDeleteStatus {
	//hardcoded value=0 because this flag is not supported on UI yet
	//but is a base filter on kubectl side so including this in our filter set
	if shouldSkipPod(pod, 0) {
		return MakePodDeleteStatusSkip()
	}
	return MakePodDeleteStatusOkay()
}

func hasLocalStorage(pod corev1.Pod) bool {
	for _, volume := range pod.Spec.Volumes {
		if volume.EmptyDir != nil {
			return true
		}
	}

	return false
}

func shouldSkipPod(pod corev1.Pod, skipDeletedTimeoutSeconds int) bool {
	return skipDeletedTimeoutSeconds > 0 &&
		!pod.ObjectMeta.DeletionTimestamp.IsZero() &&
		int(time.Now().Sub(pod.ObjectMeta.GetDeletionTimestamp().Time).Seconds()) > skipDeletedTimeoutSeconds
}

// MakePodDeleteStatusOkay is a helper method to return the corresponding PodDeleteStatus
func MakePodDeleteStatusOkay() PodDeleteStatus {
	return PodDeleteStatus{
		Delete: true,
		Reason: PodDeleteStatusTypeOkay,
	}
}

// MakePodDeleteStatusSkip is a helper method to return the corresponding PodDeleteStatus
func MakePodDeleteStatusSkip() PodDeleteStatus {
	return PodDeleteStatus{
		Delete: false,
		Reason: PodDeleteStatusTypeSkip,
	}
}

// MakePodDeleteStatusWithWarning is a helper method to return the corresponding PodDeleteStatus
func MakePodDeleteStatusWithWarning(delete bool, message string) PodDeleteStatus {
	return PodDeleteStatus{
		Delete:  delete,
		Reason:  PodDeleteStatusTypeWarning,
		Message: message,
	}
}

// MakePodDeleteStatusWithError is a helper method to return the corresponding PodDeleteStatus
func MakePodDeleteStatusWithError(message string) PodDeleteStatus {
	return PodDeleteStatus{
		Delete:  false,
		Reason:  PodDeleteStatusTypeError,
		Message: message,
	}
}
