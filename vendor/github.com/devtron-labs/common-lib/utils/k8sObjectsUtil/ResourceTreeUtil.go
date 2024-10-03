package k8sObjectsUtil

import (
	"encoding/binary"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"github.com/devtron-labs/common-lib/utils/k8s/health"
	"hash"
	"hash/fnv"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/rand"
)

func GetNodeFromResource(manifest *unstructured.Unstructured, resourceReference v1.ObjectReference, ownerRefs []metav1.OwnerReference, requestedNamespace string) (*commonBean.ResourceNode, error) {
	gvk := manifest.GroupVersionKind()
	_namespace := manifest.GetNamespace()
	if _namespace == "" {
		_namespace = requestedNamespace
	}
	ports := GetPorts(manifest, gvk)
	resourceRef := BuildResourceRef(gvk, *manifest, _namespace)

	creationTimeStamp := ""
	val, found, err := unstructured.NestedString(manifest.Object, "metadata", "creationTimestamp")
	if found && err == nil {
		creationTimeStamp = val
	}
	node := &commonBean.ResourceNode{
		ResourceRef:     resourceRef,
		ResourceVersion: manifest.GetResourceVersion(),
		NetworkingInfo: &commonBean.ResourceNetworkingInfo{
			Labels: manifest.GetLabels(),
		},
		CreatedAt: creationTimeStamp,
		Port:      ports,
	}
	node.IsHook, node.HookType = GetHookMetadata(manifest)

	// set health of Node
	SetHealthStatusForNode(node, manifest, gvk)

	// hibernate set starts
	if len(ownerRefs) == 0 {

		// set CanBeHibernated
		SetHibernationRules(node, &node.Manifest)
	}
	// hibernate set ends

	if IsPod(gvk.Kind, gvk.Group) {
		infoItems, _ := PopulatePodInfo(manifest)
		node.Info = infoItems
	}
	AddSelectiveInfoInResourceNode(node, gvk, manifest.Object)

	return node, nil
}

func GetPorts(manifest *unstructured.Unstructured, gvk schema.GroupVersionKind) []int64 {
	ports := make([]int64, 0)
	if gvk.Kind == commonBean.ServiceKind {
		ports = append(ports, getPortsFromService(manifest)...)
	}
	if gvk.Kind == commonBean.EndPointsSlice {
		ports = append(ports, getPortsFromEndPointsSlice(manifest)...)
	}
	if gvk.Kind == commonBean.EndpointsKind {
		ports = append(ports, getPortsFromEndpointsKind(manifest)...)
	}
	return ports
}

func getPortsFromService(manifest *unstructured.Unstructured) []int64 {
	var ports []int64
	if manifest.Object["spec"] != nil {
		spec := manifest.Object["spec"].(map[string]interface{})
		if spec["ports"] != nil {
			portList := spec["ports"].([]interface{})
			for _, portItem := range portList {
				if portItem.(map[string]interface{}) != nil {
					_portNumber := portItem.(map[string]interface{})["port"]
					portNumber := _portNumber.(int64)
					if portNumber != 0 {
						ports = append(ports, portNumber)
					}
				}
			}
		}
	}
	return ports
}

func getPortsFromEndPointsSlice(manifest *unstructured.Unstructured) []int64 {
	var ports []int64
	if manifest.Object["ports"] != nil {
		endPointsSlicePorts := manifest.Object["ports"].([]interface{})
		for _, val := range endPointsSlicePorts {
			_portNumber := val.(map[string]interface{})["port"]
			portNumber := _portNumber.(int64)
			if portNumber != 0 {
				ports = append(ports, portNumber)
			}
		}
	}
	return ports
}

func getPortsFromEndpointsKind(manifest *unstructured.Unstructured) []int64 {
	var ports []int64
	if manifest.Object["subsets"] != nil {
		subsets := manifest.Object["subsets"].([]interface{})
		for _, subset := range subsets {
			subsetObj := subset.(map[string]interface{})
			if subsetObj != nil {
				portsIfs := subsetObj["ports"].([]interface{})
				for _, portsIf := range portsIfs {
					portsIfObj := portsIf.(map[string]interface{})
					if portsIfObj != nil {
						port := portsIfObj["port"].(int64)
						ports = append(ports, port)
					}
				}
			}
		}
	}
	return ports
}

func BuildResourceRef(gvk schema.GroupVersionKind, manifest unstructured.Unstructured, namespace string) *commonBean.ResourceRef {
	resourceRef := &commonBean.ResourceRef{
		Group:     gvk.Group,
		Version:   gvk.Version,
		Kind:      gvk.Kind,
		Namespace: namespace,
		Name:      manifest.GetName(),
		UID:       string(manifest.GetUID()),
		Manifest:  manifest,
	}
	return resourceRef
}

func GetHookMetadata(manifest *unstructured.Unstructured) (bool, string) {
	annotations, found, _ := unstructured.NestedStringMap(manifest.Object, "metadata", "annotations")
	if found {
		if hookType, ok := annotations[commonBean.HelmHookAnnotation]; ok {
			return true, hookType
		}
	}
	return false, ""
}

func SetHealthStatusForNode(res *commonBean.ResourceNode, un *unstructured.Unstructured, gvk schema.GroupVersionKind) {
	if IsService(gvk) && un.GetName() == commonBean.DEVTRON_SERVICE_NAME && IsDevtronApp(res.NetworkingInfo.Labels) {
		res.Health = &commonBean.HealthStatus{
			Status: commonBean.HealthStatusHealthy,
		}
	} else {
		if healthCheck := health.GetHealthCheckFunc(gvk); healthCheck != nil {
			health, err := healthCheck(un)
			if err != nil {
				res.Health = &commonBean.HealthStatus{
					Status:  commonBean.HealthStatusUnknown,
					Message: err.Error(),
				}
			} else if health != nil {
				res.Health = &commonBean.HealthStatus{
					Status:  string(health.Status),
					Message: health.Message,
				}
			}
		}
	}
}

func SetHibernationRules(res *commonBean.ResourceNode, un *unstructured.Unstructured) {
	if un.GetOwnerReferences() == nil {
		// set CanBeHibernated
		replicas, found, _ := unstructured.NestedInt64(un.UnstructuredContent(), "spec", "replicas")
		if found {
			res.CanBeHibernated = true
		}

		// set IsHibernated
		annotations := un.GetAnnotations()
		if annotations != nil {
			if val, ok := annotations[commonBean.HibernateReplicaAnnotation]; ok {
				if val != "0" && replicas == 0 {
					res.IsHibernated = true
				}
			}
		}
	}
}

func PopulatePodInfo(un *unstructured.Unstructured) ([]commonBean.InfoItem, error) {
	var infoItems []commonBean.InfoItem

	pod := v1.Pod{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, &pod)
	if err != nil {
		return nil, err
	}
	restarts := 0
	totalContainers := len(pod.Spec.Containers)
	readyContainers := 0

	reason := string(pod.Status.Phase)
	if pod.Status.Reason != "" {
		reason = pod.Status.Reason
	}

	initializing := false
	for i := range pod.Status.InitContainerStatuses {
		container := pod.Status.InitContainerStatuses[i]
		restarts += int(container.RestartCount)
		switch {
		case container.State.Terminated != nil && container.State.Terminated.ExitCode == 0:
			continue
		case container.State.Terminated != nil:
			// initialization is failed
			if len(container.State.Terminated.Reason) == 0 {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Init:Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("Init:ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else {
				reason = "Init:" + container.State.Terminated.Reason
			}
			initializing = true
		case container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing":
			reason = "Init:" + container.State.Waiting.Reason
			initializing = true
		default:
			reason = fmt.Sprintf("Init:%d/%d", i, len(pod.Spec.InitContainers))
			initializing = true
		}
		break
	}
	if !initializing {
		restarts = 0
		hasRunning := false
		for i := len(pod.Status.ContainerStatuses) - 1; i >= 0; i-- {
			container := pod.Status.ContainerStatuses[i]

			restarts += int(container.RestartCount)
			if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
				reason = container.State.Waiting.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason != "" {
				reason = container.State.Terminated.Reason
			} else if container.State.Terminated != nil && container.State.Terminated.Reason == "" {
				if container.State.Terminated.Signal != 0 {
					reason = fmt.Sprintf("Signal:%d", container.State.Terminated.Signal)
				} else {
					reason = fmt.Sprintf("ExitCode:%d", container.State.Terminated.ExitCode)
				}
			} else if container.Ready && container.State.Running != nil {
				hasRunning = true
				readyContainers++
			}
		}

		// change pod status back to "Running" if there is at least one container still reporting as "Running" status
		if reason == "Completed" && hasRunning {
			reason = "Running"
		}
	}

	// "NodeLost" = https://github.com/kubernetes/kubernetes/blob/cb8ad64243d48d9a3c26b11b2e0945c098457282/pkg/util/node/node.go#L46
	// But depending on the k8s.io/kubernetes package just for a constant
	// is not worth it.
	// See https://github.com/argoproj/argo-cd/issues/5173
	// and https://github.com/kubernetes/kubernetes/issues/90358#issuecomment-617859364
	if pod.DeletionTimestamp != nil && pod.Status.Reason == "NodeLost" {
		reason = "Unknown"
	} else if pod.DeletionTimestamp != nil {
		reason = "Terminating"
	}
	infoItems = getAllInfoItems(infoItems, reason, restarts, readyContainers, totalContainers, pod)
	return infoItems, nil
}

func getAllInfoItems(infoItems []commonBean.InfoItem, reason string, restarts int, readyContainers int, totalContainers int, pod v1.Pod) []commonBean.InfoItem {
	if reason != "" {
		infoItems = append(infoItems, commonBean.InfoItem{Name: commonBean.StatusReason, Value: reason})
	}
	infoItems = append(infoItems, commonBean.InfoItem{Name: commonBean.Node, Value: pod.Spec.NodeName})

	containerNames, initContainerNames, ephemeralContainersInfo, ephemeralContainerStatus := getContainersInfo(pod)

	infoItems = append(infoItems, commonBean.InfoItem{Name: commonBean.ContainersType, Value: fmt.Sprintf("%d/%d", readyContainers, totalContainers)})
	infoItems = append(infoItems, commonBean.InfoItem{Name: commonBean.ContainersNamesType, Value: containerNames})
	infoItems = append(infoItems, commonBean.InfoItem{Name: commonBean.InitContainersNamesType, Value: initContainerNames})
	infoItems = append(infoItems, commonBean.InfoItem{Name: commonBean.EphemeralContainersInfoType, Value: ephemeralContainersInfo})
	infoItems = append(infoItems, commonBean.InfoItem{Name: commonBean.EphemeralContainersStatusType, Value: ephemeralContainerStatus})

	if restarts > 0 {
		infoItems = append(infoItems, commonBean.InfoItem{Name: commonBean.RestartCount, Value: fmt.Sprintf("%d", restarts)})
	}
	return infoItems
}

func getContainersInfo(pod v1.Pod) ([]string, []string, []commonBean.EphemeralContainerInfo, []commonBean.EphemeralContainerStatusesInfo) {
	containerNames := make([]string, 0, len(pod.Spec.Containers))
	initContainerNames := make([]string, 0, len(pod.Spec.InitContainers))
	ephemeralContainers := make([]commonBean.EphemeralContainerInfo, 0, len(pod.Spec.EphemeralContainers))
	ephemeralContainerStatus := make([]commonBean.EphemeralContainerStatusesInfo, 0, len(pod.Status.EphemeralContainerStatuses))
	for _, container := range pod.Spec.Containers {
		containerNames = append(containerNames, container.Name)
	}
	for _, initContainer := range pod.Spec.InitContainers {
		initContainerNames = append(initContainerNames, initContainer.Name)
	}
	for _, ec := range pod.Spec.EphemeralContainers {
		ecData := commonBean.EphemeralContainerInfo{
			Name:    ec.Name,
			Command: ec.Command,
		}
		ephemeralContainers = append(ephemeralContainers, ecData)
	}
	for _, ecStatus := range pod.Status.EphemeralContainerStatuses {
		status := commonBean.EphemeralContainerStatusesInfo{
			Name:  ecStatus.Name,
			State: ecStatus.State,
		}
		ephemeralContainerStatus = append(ephemeralContainerStatus, status)
	}
	return containerNames, initContainerNames, ephemeralContainers, ephemeralContainerStatus
}

func AddSelectiveInfoInResourceNode(resourceNode *commonBean.ResourceNode, gvk schema.GroupVersionKind, obj map[string]interface{}) {
	if gvk.Kind == commonBean.StatefulSetKind {
		resourceNode.UpdateRevision = GetUpdateRevisionForStatefulSet(obj)
	}
	if gvk.Kind == commonBean.DeploymentKind {
		deployment, _ := ConvertToV1Deployment(obj)
		if deployment == nil {
			return
		}
		deploymentPodHash := ComputePodHash(&deployment.Spec.Template, deployment.Status.CollisionCount)
		resourceNode.DeploymentPodHash = deploymentPodHash
		resourceNode.DeploymentCollisionCount = deployment.Status.CollisionCount
	}
	if gvk.Kind == commonBean.K8sClusterResourceRolloutKind {
		rolloutPodHash, found, _ := unstructured.NestedString(obj, "status", "currentPodHash")
		if found {
			resourceNode.RolloutCurrentPodHash = rolloutPodHash
		}
	}
}

func ComputePodHash(template *v1.PodTemplateSpec, collisionCount *int32) string {
	podTemplateSpecHasher := fnv.New32a()
	DeepHashObject(podTemplateSpecHasher, *template)

	// Add collisionCount in the hash if it exists.
	if collisionCount != nil {
		collisionCountBytes := make([]byte, 8)
		binary.LittleEndian.PutUint32(collisionCountBytes, uint32(*collisionCount))
		_, err := podTemplateSpecHasher.Write(collisionCountBytes)
		if err != nil {
			fmt.Println(err)
		}
	}
	return rand.SafeEncodeString(fmt.Sprint(podTemplateSpecHasher.Sum32()))
}

func ConvertToV1Deployment(nodeObj map[string]interface{}) (*v1beta1.Deployment, error) {
	deploymentObj := v1beta1.Deployment{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(nodeObj, &deploymentObj)
	if err != nil {
		return nil, err
	}
	return &deploymentObj, nil
}

func GetUpdateRevisionForStatefulSet(obj map[string]interface{}) string {
	updateRevisionFromManifest, found, _ := unstructured.NestedString(obj, "status", "updateRevision")
	if found {
		return updateRevisionFromManifest
	}
	return ""
}

// DeepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	_, err := printer.Fprintf(hasher, "%#v", objectToWrite)
	if err != nil {
		fmt.Println(err)
	}
}

func BuildPodMetadata(nodes []*commonBean.ResourceNode) ([]*commonBean.PodMetadata, error) {
	podsMetadata := make([]*commonBean.PodMetadata, 0, len(nodes))
	for _, node := range nodes {

		if node.Kind != commonBean.PodKind {
			continue
		}
		// set containers,initContainers and ephemeral container names
		var containerNames []string
		var initContainerNames []string
		var ephemeralContainersInfo []commonBean.EphemeralContainerInfo
		var ephemeralContainerStatus []commonBean.EphemeralContainerStatusesInfo

		for _, nodeInfo := range node.Info {
			switch nodeInfo.Name {
			case commonBean.ContainersNamesType:
				containerNames = nodeInfo.Value.([]string)
			case commonBean.InitContainersNamesType:
				initContainerNames = nodeInfo.Value.([]string)
			case commonBean.EphemeralContainersInfoType:
				ephemeralContainersInfo = nodeInfo.Value.([]commonBean.EphemeralContainerInfo)
			case commonBean.EphemeralContainersStatusType:
				ephemeralContainerStatus = nodeInfo.Value.([]commonBean.EphemeralContainerStatusesInfo)
			default:
				continue
			}
		}

		ephemeralContainerStatusMap := make(map[string]bool)
		for _, c := range ephemeralContainerStatus {
			// c.state contains three states running,waiting and terminated
			// at any point of time only one state will be there
			if c.State.Running != nil {
				ephemeralContainerStatusMap[c.Name] = true
			}
		}
		ephemeralContainers := make([]*commonBean.EphemeralContainerData, 0, len(ephemeralContainersInfo))
		// sending only running ephemeral containers in the list
		for _, ec := range ephemeralContainersInfo {
			if _, ok := ephemeralContainerStatusMap[ec.Name]; ok {
				containerData := &commonBean.EphemeralContainerData{
					Name:       ec.Name,
					IsExternal: IsExternalEphemeralContainer(ec.Command, ec.Name),
				}
				ephemeralContainers = append(ephemeralContainers, containerData)
			}
		}

		podMetadata := &commonBean.PodMetadata{
			Name:                node.Name,
			UID:                 node.UID,
			Containers:          containerNames,
			InitContainers:      initContainerNames,
			EphemeralContainers: ephemeralContainers,
		}

		podsMetadata = append(podsMetadata, podMetadata)

	}
	return podsMetadata, nil
}

func GetExtraNodeInfoMappings(nodes []*commonBean.ResourceNode) (map[string]string, map[string]*commonBean.ExtraNodeInfo, map[string]*commonBean.ExtraNodeInfo) {
	deploymentPodHashMap := make(map[string]string)
	rolloutNameVsExtraNodeInfoMapping := make(map[string]*commonBean.ExtraNodeInfo)
	uidVsExtraNodeInfoMapping := make(map[string]*commonBean.ExtraNodeInfo)
	for _, node := range nodes {
		if node.Kind == commonBean.DeploymentKind {
			deploymentPodHashMap[node.Name] = node.DeploymentPodHash
		} else if node.Kind == commonBean.K8sClusterResourceRolloutKind {
			rolloutNameVsExtraNodeInfoMapping[node.Name] = &commonBean.ExtraNodeInfo{
				RolloutCurrentPodHash: node.RolloutCurrentPodHash,
			}
		} else if node.Kind == commonBean.StatefulSetKind || node.Kind == commonBean.DaemonSetKind {
			if _, ok := uidVsExtraNodeInfoMapping[node.UID]; !ok {
				uidVsExtraNodeInfoMapping[node.UID] = &commonBean.ExtraNodeInfo{UpdateRevision: node.UpdateRevision, ResourceNetworkingInfo: node.NetworkingInfo}
			}
		}
	}
	return deploymentPodHashMap, rolloutNameVsExtraNodeInfoMapping, uidVsExtraNodeInfoMapping
}

func IsPodNew(nodes []*commonBean.ResourceNode, node *commonBean.ResourceNode, deploymentPodHashMap map[string]string, rolloutMap map[string]*commonBean.ExtraNodeInfo,
	uidVsExtraNodeInfoMap map[string]*commonBean.ExtraNodeInfo) (bool, error) {

	isNew := false
	parentRef := node.ParentRefs[0]
	parentKind := parentRef.Kind

	// if parent is StatefulSet - then pod label controller-revision-hash should match StatefulSet's update revision
	if parentKind == commonBean.StatefulSetKind && node.NetworkingInfo != nil {
		isNew = uidVsExtraNodeInfoMap[parentRef.UID].UpdateRevision == node.NetworkingInfo.Labels["controller-revision-hash"]
	}

	// if parent is Job - then pod label controller-revision-hash should match StatefulSet's update revision
	if parentKind == commonBean.JobKind {
		// TODO - new or old logic not built in orchestrator for Job's pods. hence not implementing here. as don't know the logic :)
		isNew = true
	}

	// if parent kind is replica set then
	if parentKind == commonBean.ReplicaSetKind {
		replicaSetNode := GetMatchingNode(nodes, parentKind, parentRef.Name)

		// if parent of replicaset is deployment, compare label pod-template-hash
		if replicaSetParent := replicaSetNode.ParentRefs[0]; replicaSetNode != nil && len(replicaSetNode.ParentRefs) > 0 && replicaSetParent.Kind == commonBean.DeploymentKind {
			deploymentPodHash := deploymentPodHashMap[replicaSetParent.Name]
			replicaSetObj, err := GetReplicaSetObject(replicaSetNode)
			if err != nil {
				return isNew, err
			}
			deploymentNode := GetMatchingNode(nodes, replicaSetParent.Kind, replicaSetParent.Name)
			// TODO: why do we need deployment object for collisionCount ??
			var deploymentCollisionCount *int32
			if deploymentNode != nil && deploymentNode.DeploymentCollisionCount != nil {
				deploymentCollisionCount = deploymentNode.DeploymentCollisionCount
			} else {
				deploymentCollisionCount, err = getDeploymentCollisionCount(replicaSetParent)
				if err != nil {
					return isNew, err
				}
			}
			replicaSetPodHash := GetReplicaSetPodHash(replicaSetObj, deploymentCollisionCount)
			isNew = replicaSetPodHash == deploymentPodHash
		} else if replicaSetParent.Kind == commonBean.K8sClusterResourceRolloutKind {

			rolloutExtraInfo := rolloutMap[replicaSetParent.Name]
			rolloutPodHash := rolloutExtraInfo.RolloutCurrentPodHash
			replicasetPodHash := GetRolloutPodTemplateHash(replicaSetNode)

			isNew = rolloutPodHash == replicasetPodHash

		}

	}

	// if parent kind is DaemonSet then compare DaemonSet's Child ControllerRevision's label controller-revision-hash with pod label controller-revision-hash
	if parentKind == commonBean.DaemonSetKind {
		controllerRevisionNodes := GetMatchingNodes(nodes, "ControllerRevision")
		for _, controllerRevisionNode := range controllerRevisionNodes {
			if len(controllerRevisionNode.ParentRefs) > 0 && controllerRevisionNode.ParentRefs[0].Kind == parentKind &&
				controllerRevisionNode.ParentRefs[0].Name == parentRef.Name && uidVsExtraNodeInfoMap[parentRef.UID].ResourceNetworkingInfo != nil &&
				node.NetworkingInfo != nil {

				isNew = uidVsExtraNodeInfoMap[parentRef.UID].ResourceNetworkingInfo.Labels["controller-revision-hash"] == node.NetworkingInfo.Labels["controller-revision-hash"]
			}
		}
	}
	return isNew, nil
}

func GetRolloutPodTemplateHash(replicasetNode *commonBean.ResourceNode) string {
	if rolloutPodTemplateHash, ok := replicasetNode.NetworkingInfo.Labels["rollouts-pod-template-hash"]; ok {
		return rolloutPodTemplateHash
	}
	return ""
}

func getDeploymentCollisionCount(deploymentInfo *commonBean.ResourceRef) (*int32, error) {

	var deploymentNodeObj map[string]interface{}
	var err error
	deploymentNodeObj = deploymentInfo.Manifest.Object

	deploymentObj, err := ConvertToV1Deployment(deploymentNodeObj)
	if err != nil {
		return nil, err
	}
	return deploymentObj.Status.CollisionCount, nil
}

func GetMatchingNode(nodes []*commonBean.ResourceNode, kind string, name string) *commonBean.ResourceNode {
	for _, node := range nodes {
		if node.Kind == kind && node.Name == name {
			return node
		}
	}
	return nil
}

func GetMatchingNodes(nodes []*commonBean.ResourceNode, kind string) []*commonBean.ResourceNode {
	nodesRes := make([]*commonBean.ResourceNode, 0, len(nodes))
	for _, node := range nodes {
		if node.Kind == kind {
			nodesRes = append(nodesRes, node)
		}
	}
	return nodesRes
}

func GetReplicaSetObject(replicaSetNode *commonBean.ResourceNode) (*v1beta1.ReplicaSet, error) {
	var replicaSetNodeObj map[string]interface{}
	var err error
	replicaSetNodeObj = replicaSetNode.Manifest.Object

	replicaSetObj, err := ConvertToV1ReplicaSet(replicaSetNodeObj)
	if err != nil {
		return nil, err
	}
	return replicaSetObj, nil
}

func ConvertToV1ReplicaSet(nodeObj map[string]interface{}) (*v1beta1.ReplicaSet, error) {
	replicaSetObj := v1beta1.ReplicaSet{}
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(nodeObj, &replicaSetObj)
	if err != nil {
		return nil, err
	}
	return &replicaSetObj, nil
}

func GetReplicaSetPodHash(replicasetObj *v1beta1.ReplicaSet, collisionCount *int32) string {
	labels := make(map[string]string)
	for k, v := range replicasetObj.Spec.Template.Labels {
		if k != "pod-template-hash" {
			labels[k] = v
		}
	}
	replicasetObj.Spec.Template.Labels = labels
	podHash := ComputePodHash(&replicasetObj.Spec.Template, collisionCount)
	return podHash
}

func IsDevtronApp(labels map[string]string) bool {
	isDevtronApp := false
	if val, ok := labels[commonBean.DEVTRON_APP_LABEL_KEY]; ok {
		if val == commonBean.DEVTRON_APP_LABEL_VALUE1 || val == commonBean.DEVTRON_APP_LABEL_VALUE2 {
			isDevtronApp = true
		}
	}
	return isDevtronApp
}

func IsService(gvk schema.GroupVersionKind) bool {
	return gvk.Group == "" && gvk.Kind == commonBean.ServiceKind
}

func IsPod(kind string, group string) bool {
	return kind == "Pod" && group == ""
}

func GetMatchingPodMetadataForUID(podMetadatas []*commonBean.PodMetadata, uid string) *commonBean.PodMetadata {
	if len(podMetadatas) == 0 {
		return nil
	}
	for _, podMetadata := range podMetadatas {
		if podMetadata.UID == uid {
			return podMetadata
		}
	}
	return nil
}

// app health is worst of the nodes health
// or if app status is healthy then check for hibernation status
func BuildAppHealthStatus(nodes []*commonBean.ResourceNode) *commonBean.HealthStatusCode {
	appHealthStatus := commonBean.HealthStatusHealthy
	isAppFullyHibernated := true
	var isAppPartiallyHibernated bool
	var isAnyNodeCanByHibernated bool

	for _, node := range nodes {
		if node.IsHook {
			continue
		}
		nodeHealth := node.Health
		if node.CanBeHibernated {
			isAnyNodeCanByHibernated = true
			if !node.IsHibernated {
				isAppFullyHibernated = false
			} else {
				isAppPartiallyHibernated = true
			}
		}
		if nodeHealth == nil {
			continue
		}
		if health.IsWorseStatus(health.HealthStatusCode(appHealthStatus), health.HealthStatusCode(nodeHealth.Status)) {
			appHealthStatus = nodeHealth.Status
		}
	}

	// override hibernate status on app level if status is healthy and hibernation done
	if appHealthStatus == commonBean.HealthStatusHealthy && isAnyNodeCanByHibernated {
		if isAppFullyHibernated {
			appHealthStatus = commonBean.HealthStatusHibernated
		} else if isAppPartiallyHibernated {
			appHealthStatus = commonBean.HealthStatusPartiallyHibernated
		}
	}

	return &appHealthStatus
}
