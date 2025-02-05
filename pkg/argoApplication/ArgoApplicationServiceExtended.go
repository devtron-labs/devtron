package argoApplication

import (
	"context"
	"encoding/json"
	"fmt"
	application2 "github.com/argoproj/argo-cd/v2/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	openapi "github.com/devtron-labs/devtron/api/helm-app/openapiClient"
	"github.com/devtron-labs/devtron/client/argocdServer"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	"github.com/devtron-labs/devtron/util"
	v12 "k8s.io/api/apps/v1"
	"net/http"
	"strings"
	"time"
)

type ArgoApplicationServiceExtendedImpl struct {
	*ArgoApplicationServiceImpl
	acdClientWrapper argocdServer.ArgoClientWrapperService
}

func NewArgoApplicationServiceExtendedServiceImpl(argoApplicationServiceImpl *ArgoApplicationServiceImpl,
	acdClientWrapper argocdServer.ArgoClientWrapperService) *ArgoApplicationServiceExtendedImpl {
	return &ArgoApplicationServiceExtendedImpl{
		ArgoApplicationServiceImpl: argoApplicationServiceImpl,
		acdClientWrapper:           acdClientWrapper,
	}
}

func (c *ArgoApplicationServiceExtendedImpl) ListApplications(clusterIds []int) ([]*bean.ArgoApplicationListDto, error) {
	return c.ArgoApplicationServiceImpl.ListApplications(clusterIds)
}

func (c *ArgoApplicationServiceExtendedImpl) HibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	return c.ArgoApplicationServiceImpl.HibernateArgoApplication(ctx, app, hibernateRequest)
}

func (c *ArgoApplicationServiceExtendedImpl) UnHibernateArgoApplication(ctx context.Context, app *bean.ArgoAppIdentifier, hibernateRequest *openapi.HibernateRequest) ([]*openapi.HibernateStatus, error) {
	return c.ArgoApplicationServiceImpl.UnHibernateArgoApplication(ctx, app, hibernateRequest)
}

func (c *ArgoApplicationServiceExtendedImpl) getArgoAppStatusMetaData(application *v1alpha1.Application,
	resp *v1alpha1.ApplicationTree) ([]v1alpha1.ApplicationCondition, map[string]string, string, string) {
	conditions := make([]v1alpha1.ApplicationCondition, 0)
	resourcesSyncResultMap := make(map[string]string)
	status := "Unknown"
	hash := ""
	if application != nil {
		// https://github.com/argoproj/argo-cd/issues/11234 workaround
		updateNodeHealthStatus(resp, application)
		argoApplicationStatus := application.Status
		status = string(argoApplicationStatus.Health.Status)
		hash = argoApplicationStatus.Sync.Revision
		conditions = argoApplicationStatus.Conditions
		for _, condition := range conditions {
			if condition.Type != v1alpha1.ApplicationConditionSharedResourceWarning {
				status = "Degraded"
			}
		}
		if argoApplicationStatus.OperationState != nil && argoApplicationStatus.OperationState.SyncResult != nil {
			resourcesSyncResults := argoApplicationStatus.OperationState.SyncResult.Resources
			for _, resourcesSyncResult := range resourcesSyncResults {
				if resourcesSyncResult == nil {
					continue
				}
				resourceIdentifier := fmt.Sprintf("%s/%s", resourcesSyncResult.Kind, resourcesSyncResult.Name)
				resourcesSyncResultMap[resourceIdentifier] = resourcesSyncResult.Message
			}
		}
		if status == "" {
			status = "Unknown"
		}
	}
	return conditions, resourcesSyncResultMap, status, hash
}

func (c *ArgoApplicationServiceExtendedImpl) getApplicationObjectWithK8sClient(ctxt context.Context,
	acdQueryRequest *bean.AcdClientQueryRequest) (*v1alpha1.Application, error) {
	var appNamespace, applicationName string
	if acdQueryRequest.Query.AppNamespace == nil {
		appNamespace = *acdQueryRequest.Query.AppNamespace
	}
	if acdQueryRequest.Query.ApplicationName == nil {
		applicationName = *acdQueryRequest.Query.ApplicationName
	}
	application, err := c.acdClientWrapper.GetArgoAppByNameWithK8sClient(ctxt, acdQueryRequest.ClusterId, appNamespace, applicationName)
	if err != nil {
		c.logger.Errorw("error in fetching application", "acdQueryRequest", acdQueryRequest, "err", err)
		return nil, err
	}
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		c.logger.Errorw("error in marshalling application", "acdQueryRequest", acdQueryRequest, "err", err)
		return nil, err
	}
	var argoApplicationSpec v1alpha1.Application
	if err = json.Unmarshal(applicationJSON, &argoApplicationSpec); err != nil {
		c.logger.Errorw("error in unmarshalling application", "acdQueryRequest", acdQueryRequest, "err", err)
		return nil, err
	}
	return &argoApplicationSpec, nil
}

func (c *ArgoApplicationServiceExtendedImpl) getApplicationObjectWithAcdClient(ctxt context.Context,
	asc application2.ApplicationServiceClient, acdQueryRequest *bean.AcdClientQueryRequest) *v1alpha1.Application {
	appQuery := application2.ApplicationQuery{Name: acdQueryRequest.Query.ApplicationName, AppNamespace: acdQueryRequest.Query.AppNamespace}
	app, _ := asc.Watch(ctxt, &appQuery)
	if app != nil {
		appResp, err := app.Recv()
		if err == nil {
			return &appResp.Application
		}
	}
	return nil
}

func (c *ArgoApplicationServiceExtendedImpl) getApplicationObject(ctxt context.Context,
	asc application2.ApplicationServiceClient, acdQueryRequest *bean.AcdClientQueryRequest) *v1alpha1.Application {
	if acdQueryRequest.Mode.IsDeclarative() {
		argoApplicationSpec, err := c.getApplicationObjectWithK8sClient(ctxt, acdQueryRequest)
		if err != nil {
			c.logger.Errorw("error in fetching application", "acdQueryRequest", acdQueryRequest, "err", err)
			return nil
		}
		return argoApplicationSpec
	} else {
		return c.getApplicationObjectWithAcdClient(ctxt, asc, acdQueryRequest)
	}
}

func (c *ArgoApplicationServiceExtendedImpl) ResourceTree(ctx context.Context, acdQueryRequest *bean.AcdClientQueryRequest) (*argoApplication.ResourceTreeResponse, error) {
	if acdQueryRequest == nil || acdQueryRequest.Query == nil {
		return nil, util2.NewApiError(http.StatusInternalServerError, "something went wrong!", "invalid argo application query request")
	}
	//all the apps deployed via argo are fetching status from here
	newCtx, cancel := context.WithTimeout(ctx, argoApplication.TimeoutSlow)
	defer cancel()

	asc, conn, err := c.acdClientWrapper.GetArgoClient(ctx)
	if err != nil {
		c.logger.Errorw("Error in GetArgoClient", "err", err)
		return nil, err
	}
	defer util.Close(conn, c.logger)
	c.logger.Debugw("GRPC_GET_RESOURCETREE", "req", acdQueryRequest)
	resp, err := asc.ResourceTree(newCtx, acdQueryRequest.Query)
	if err != nil {
		c.logger.Errorw("GRPC_GET_RESOURCETREE", "req", acdQueryRequest, "err", err)
		return nil, err
	}
	responses := c.parseResult(resp, acdQueryRequest.Query, newCtx, asc, err)
	podMetadata, newReplicaSets := c.buildPodMetadata(resp, responses)

	conditions, resourcesSyncResultMap, status, hash := c.getArgoAppStatusMetaData(c.getApplicationObject(ctx, asc, acdQueryRequest), resp)
	return &argoApplication.ResourceTreeResponse{
		ApplicationTree:          resp,
		PodMetadata:              podMetadata,
		NewGenerationReplicaSets: newReplicaSets,
		Conditions:               conditions,
		ResourcesSyncResultMap:   resourcesSyncResultMap,
		Status:                   status,
		RevisionHash:             hash,
	}, err
}

func (impl *ArgoApplicationServiceImpl) parseResult(resp *v1alpha1.ApplicationTree, query *application2.ResourcesQuery, ctx context.Context, asc application2.ApplicationServiceClient, err error) []*argoApplication.Result {
	var responses = make([]*argoApplication.Result, 0)
	qCount := 0
	response := make(chan argoApplication.Result)
	if err != nil || resp == nil || len(resp.Nodes) == 0 {
		return responses
	}
	needPods := false
	queryNodes := make([]v1alpha1.ResourceNode, 0)
	podParents := make(map[string]v1alpha1.ResourceNode)
	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			for _, pr := range node.ParentRefs {
				podParents[pr.Name] = node
			}
		}
	}
	for _, node := range resp.Nodes {
		if node.Kind == "Rollout" || node.Kind == "Deployment" || node.Kind == "StatefulSet" || node.Kind == "DaemonSet" {
			queryNodes = append(queryNodes, node)
		}
		if node.Kind == "ReplicaSet" {
			if _, ok := podParents[node.Name]; ok {
				queryNodes = append(queryNodes, node)
			}
		}
		if node.Kind == "StatefulSet" || node.Kind == "DaemonSet" || node.Kind == "Workflow" {
			needPods = true
		}

		if node.Kind == "CronJob" || node.Kind == "Job" {
			queryNodes = append(queryNodes, node)
			needPods = true
		}
	}

	impl.logger.Debugw("needPods", "pods", needPods)

	for _, node := range podParents {
		queryNodes = append(queryNodes, node)
	}

	if needPods {
		for _, node := range resp.Nodes {
			if node.Kind == "Pod" {
				queryNodes = append(queryNodes, node)
			}
		}
	}

	relevantCR := make(map[string]bool)
	for _, node := range resp.Nodes {
		prefix := ""
		if len(node.ParentRefs) > 0 {
			for _, p := range node.ParentRefs {
				if p.Kind == "DaemonSet" {
					prefix = p.Name
				}
			}
		}
		if node.Kind == "Pod" && node.NetworkingInfo != nil && node.NetworkingInfo.Labels != nil {
			relevantCR[prefix+"-"+node.NetworkingInfo.Labels["controller-revision-hash"]] = true
		}
	}

	for _, node := range resp.Nodes {
		if node.Kind == "ControllerRevision" {
			if ok := relevantCR[node.Name]; ok {
				queryNodes = append(queryNodes, node)
			}
		}
	}

	for _, node := range queryNodes {
		rQuery := transform(node, query.ApplicationName)
		qCount++
		go func(request application2.ApplicationResourceRequest) {
			ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()
			startTime := time.Now()
			res, err := asc.GetResource(ctx, &request)
			if err != nil {
				impl.logger.Errorw("GRPC_GET_RESOURCE", "data", request, "timeTaken", time.Since(startTime), "err", err)
			} else {
				impl.logger.Debugw("GRPC_GET_RESOURCE", "data", request, "timeTaken", time.Since(startTime))
			}
			if res != nil || err != nil {
				response <- argoApplication.Result{Response: res, Error: err, Request: &request}
			} else {
				response <- argoApplication.Result{Response: nil, Error: fmt.Errorf("connection closed by client"), Request: &request}
			}
		}(*rQuery)
	}

	if qCount == 0 {
		return responses
	}

	rCount := 0
	for {
		select {
		case msg, ok := <-response:
			if ok {
				if msg.Error == nil {
					responses = append(responses, &msg)
				}
			}
			rCount++
		}
		if qCount == rCount {
			break
		}
	}
	return responses
}

func (impl *ArgoApplicationServiceImpl) buildPodMetadata(resp *v1alpha1.ApplicationTree, responses []*argoApplication.Result) (podMetaData []*argoApplication.PodMetadata, newReplicaSets []string) {
	rolloutManifests := make([]map[string]interface{}, 0)
	statefulSetManifest := make(map[string]interface{})
	deploymentManifests := make([]map[string]interface{}, 0)
	daemonSetManifest := make(map[string]interface{})
	replicaSetManifests := make([]map[string]interface{}, 0)
	podManifests := make([]map[string]interface{}, 0)
	controllerRevisionManifests := make([]map[string]interface{}, 0)
	jobsManifest := make(map[string]interface{})
	var parentWorkflow []string
	for _, response := range responses {
		if response != nil && response.Response != nil {
			kind := *response.Request.Kind
			manifestFromResponse := *response.Response.Manifest
			if kind == "Rollout" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					impl.logger.Error(err)
				} else {
					rolloutManifests = append(rolloutManifests, manifest)
				}
			} else if kind == "Deployment" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					impl.logger.Error(err)
				} else {
					deploymentManifests = append(deploymentManifests, manifest)
				}
			} else if kind == "StatefulSet" {
				err := json.Unmarshal([]byte(manifestFromResponse), &statefulSetManifest)
				if err != nil {
					impl.logger.Error(err)
				}
			} else if kind == "DaemonSet" {
				err := json.Unmarshal([]byte(manifestFromResponse), &daemonSetManifest)
				if err != nil {
					impl.logger.Error(err)
				}
			} else if kind == "ReplicaSet" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					impl.logger.Error(err)
				}
				replicaSetManifests = append(replicaSetManifests, manifest)
			} else if kind == "Pod" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					impl.logger.Error(err)
				}
				podManifests = append(podManifests, manifest)
			} else if kind == "ControllerRevision" {
				manifest := make(map[string]interface{})
				err := json.Unmarshal([]byte(manifestFromResponse), &manifest)
				if err != nil {
					impl.logger.Error(err)
				}
				controllerRevisionManifests = append(controllerRevisionManifests, manifest)
			} else if kind == "Job" {
				err := json.Unmarshal([]byte(manifestFromResponse), &jobsManifest)
				if err != nil {
					impl.logger.Error(err)
				}
			}
		}
	}
	newPodNames := make(map[string]bool, 0)
	// for rollout we compare pod hash
	for _, rolloutManifest := range rolloutManifests {
		if _, ok := rolloutManifest["kind"]; ok {
			newReplicaSet := getRolloutNewReplicaSetName(rolloutManifest, replicaSetManifests)
			if len(newReplicaSet) > 0 {
				newReplicaSets = append(newReplicaSets, newReplicaSet)
			}
		}
	}

	for _, deploymentManifest := range deploymentManifests {
		if _, ok := deploymentManifest["kind"]; ok {
			newReplicaSet := getDeploymentNewReplicaSetName(deploymentManifest, replicaSetManifests)
			if len(newReplicaSet) > 0 {
				newReplicaSets = append(newReplicaSets, newReplicaSet)
			}
		}
	}

	if _, ok := statefulSetManifest["kind"]; ok {
		newPodNames = getStatefulSetNewPods(statefulSetManifest, podManifests)
	}

	if _, ok := daemonSetManifest["kind"]; ok {
		newPodNames = getDaemonSetNewPods(daemonSetManifest, podManifests, controllerRevisionManifests)
	}

	if _, ok := jobsManifest["kind"]; ok {
		newPodNames = getJobsNewPods(jobsManifest, podManifests)
	}

	for _, node := range resp.Nodes {
		if node.Kind == "Workflow" {
			parentWorkflow = append(parentWorkflow, node.Name)
		}
	}

	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			if contains(parentWorkflow, node.Name) {
				newPodNames[node.Name] = true
			}
		}
	}

	//duplicatePodToReplicasetMapping can contain following data {Pod1: RS1, Pod2: RS1, Pod4: RS2, Pod5:RS2}, it contains pod
	//to replica set mapping, where key is podName and value is its respective replicaset name, multiple keys(podName) can have
	//single value (replicasetName)
	duplicatePodToReplicasetMapping := make(map[string]string)
	if len(newReplicaSets) > 0 {
		results, duplicateMapping := buildPodMetadataFromReplicaSet(resp, newReplicaSets, replicaSetManifests)
		for _, meta := range results {
			podMetaData = append(podMetaData, meta)
		}
		duplicatePodToReplicasetMapping = duplicateMapping
	}

	if newPodNames != nil {
		podsMetadataFromPods := buildPodMetadataFromPod(resp, podManifests, newPodNames)
		podMetaData = updateMetadataOfDuplicatePods(podsMetadataFromPods, duplicatePodToReplicasetMapping, podMetaData)
	}
	return
}

func getDeploymentNewReplicaSetName(deploymentManifest map[string]interface{}, replicaSetManifests []map[string]interface{}) (newReplicaSet string) {
	d, err := json.Marshal(deploymentManifest)
	if err != nil {
		return
	}
	deployment := &v12.Deployment{}
	err = json.Unmarshal(d, deployment)
	if err != nil {
		return
	}
	dPodHash := util.ComputeHash(&deployment.Spec.Template, deployment.Status.CollisionCount)
	for _, rs := range replicaSetManifests {
		r, err := json.Marshal(rs)
		if err != nil {
			return
		}
		replicaset := &v12.ReplicaSet{}
		err = json.Unmarshal(r, replicaset)
		if err != nil {
			continue
		}
		rsCopy := replicaset.Spec.DeepCopy()
		labels := make(map[string]string)
		for k, v := range rsCopy.Template.Labels {
			if k != "pod-template-hash" {
				labels[k] = v
			}
		}
		rsCopy.Template.Labels = labels
		podHash := util.ComputeHash(&rsCopy.Template, deployment.Status.CollisionCount)
		if podHash == dPodHash {
			newReplicaSet = getResourceName(rs)
		}
	}
	return
}

func getDaemonSetNewPods(daemonSetManifest map[string]interface{}, podManifests []map[string]interface{}, controllerRevisionManifests []map[string]interface{}) (newPodNames map[string]bool) {
	d, err := json.Marshal(daemonSetManifest)
	if err != nil {
		return
	}
	daemonSet := &v12.DaemonSet{}
	err = json.Unmarshal(d, daemonSet)
	if err != nil {
		return
	}
	latestRevision := ""
	latestGen := 0
	newPodNames = make(map[string]bool, 0)
	for _, crm := range controllerRevisionManifests {
		rev := int(crm["revision"].(float64))
		if latestGen < rev {
			latestGen = rev
			latestRevision = getDaemonSetPodControllerRevisionHash(crm)
		}
	}
	for _, pod := range podManifests {
		podRevision := getDaemonSetPodControllerRevisionHash(pod)
		if latestRevision == podRevision {
			newPodNames[getResourceName(pod)] = true
		}
	}
	return
}

func getDaemonSetPodControllerRevisionHash(pod map[string]interface{}) string {
	if md, ok := pod["metadata"]; ok {
		if mdm, ok := md.(map[string]interface{}); ok {
			if l, ok := mdm["labels"]; ok {
				if lm, ok := l.(map[string]interface{}); ok {
					if h, ok := lm["controller-revision-hash"]; ok {
						if hs, ok := h.(string); ok {
							return hs
						}
					}
				}
			}
		}
	}
	return ""
}

func getStatefulSetNewPods(statefulSetManifest map[string]interface{}, podManifests []map[string]interface{}) (newPodNames map[string]bool) {
	newPodNames = make(map[string]bool, 0)
	updateRevision := getStatefulSetUpdateRevision(statefulSetManifest)
	for _, pod := range podManifests {
		podRevision := getStatefulSetPodControllerRevisionHash(pod)
		if updateRevision == podRevision {
			newPodNames[getResourceName(pod)] = true
		}
	}
	return
}

func getStatefulSetUpdateRevision(statefulSet map[string]interface{}) string {
	if s, ok := statefulSet["status"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if cph, ok := sm["updateRevision"]; ok {
				if cphs, ok := cph.(string); ok {
					return cphs
				}
			}
		}
	}
	return ""
}

func getStatefulSetPodControllerRevisionHash(pod map[string]interface{}) string {
	if md, ok := pod["metadata"]; ok {
		if mdm, ok := md.(map[string]interface{}); ok {
			if l, ok := mdm["labels"]; ok {
				if lm, ok := l.(map[string]interface{}); ok {
					if h, ok := lm["controller-revision-hash"]; ok {
						if hs, ok := h.(string); ok {
							return hs
						}
					}
				}
			}
		}
	}
	return ""
}

func getRolloutNewReplicaSetName(rManifest map[string]interface{}, replicaSetManifests []map[string]interface{}) (newReplicaSet string) {
	rPodHash := getRolloutPodHash(rManifest)
	for _, rs := range replicaSetManifests {
		podHash := getRolloutPodTemplateHash(rs)
		if podHash == rPodHash {
			newReplicaSet = getResourceName(rs)
		}
	}
	return newReplicaSet
}

func getRolloutPodHash(rollout map[string]interface{}) string {
	if s, ok := rollout["status"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if cph, ok := sm["currentPodHash"]; ok {
				if cphs, ok := cph.(string); ok {
					return cphs
				}
			}
		}
	}
	return ""
}

func getRolloutPodTemplateHash(pod map[string]interface{}) string {
	if md, ok := pod["metadata"]; ok {
		if mdm, ok := md.(map[string]interface{}); ok {
			if l, ok := mdm["labels"]; ok {
				if lm, ok := l.(map[string]interface{}); ok {
					if h, ok := lm["rollouts-pod-template-hash"]; ok {
						if hs, ok := h.(string); ok {
							return hs
						}
					}
				}
			}
		}
	}
	return ""
}

func buildPodMetadataFromPod(resp *v1alpha1.ApplicationTree, podManifests []map[string]interface{}, newPodNames map[string]bool) (podMetadata []*argoApplication.PodMetadata) {
	containerMapping := make(map[string][]*string)
	initContainerMapping := make(map[string][]*string)
	for _, pod := range podManifests {
		containerMapping[getResourceName(pod)] = getPodContainers(pod)
	}

	for _, pod := range podManifests {
		initContainerMapping[getResourceName(pod)] = getPodInitContainers(pod)
	}

	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			isNew := newPodNames[node.Name]
			metadata := argoApplication.PodMetadata{Name: node.Name, UID: node.UID, Containers: containerMapping[node.Name], InitContainers: initContainerMapping[node.Name], IsNew: isNew}
			podMetadata = append(podMetadata, &metadata)
		}
	}
	return
}

func contains(elems []string, v string) bool {
	for _, s := range elems {
		if strings.HasPrefix(v, s) {
			return true
		}
	}
	return false
}

func getPodContainers(resource map[string]interface{}) []*string {
	containers := make([]*string, 0)
	if s, ok := resource["spec"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if c, ok := sm["containers"]; ok {
				if cas, ok := c.([]interface{}); ok {
					for _, ca := range cas {
						if cam, ok := ca.(map[string]interface{}); ok {
							if n, ok := cam["name"]; ok {
								if cn, ok := n.(string); ok {
									containers = append(containers, &cn)
								}
							}
						}
					}
				}
			}
		}
	}
	return containers
}

func getPodInitContainers(resource map[string]interface{}) []*string {
	containers := make([]*string, 0)
	if s, ok := resource["spec"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if c, ok := sm["initContainers"]; ok {
				if cas, ok := c.([]interface{}); ok {
					for _, ca := range cas {
						if cam, ok := ca.(map[string]interface{}); ok {
							if n, ok := cam["name"]; ok {
								if cn, ok := n.(string); ok {
									containers = append(containers, &cn)
								}
							}
						}
					}
				}
			}
		}
	}
	return containers
}

func buildPodMetadataFromReplicaSet(resp *v1alpha1.ApplicationTree, newReplicaSets []string, replicaSetManifests []map[string]interface{}) ([]*argoApplication.PodMetadata, map[string]string) {
	replicaSets := make(map[string]map[string]interface{})
	podToReplicasetMapping := make(map[string]string)
	var podMetadata []*argoApplication.PodMetadata
	for _, replicaSet := range replicaSetManifests {
		replicaSets[getResourceName(replicaSet)] = replicaSet
	}
	for _, node := range resp.Nodes {
		if node.Kind == "Pod" {
			parentName := ""
			for _, p := range node.ParentRefs {
				if p.Kind == "ReplicaSet" {
					parentName = p.Name
				}
			}
			if parentName != "" {
				isNew := false
				for _, newReplicaSet := range newReplicaSets {
					if parentName == newReplicaSet {
						isNew = true
						break
					}
				}
				replicaSet := replicaSets[parentName]
				containers, intContainers := getReplicaSetContainers(replicaSet)
				podToReplicasetMapping[node.Name] = parentName
				metadata := argoApplication.PodMetadata{Name: node.Name, UID: node.UID, Containers: containers, InitContainers: intContainers, IsNew: isNew}
				podMetadata = append(podMetadata, &metadata)
			}
		}
	}
	return podMetadata, podToReplicasetMapping
}

func getReplicaSetContainers(resource map[string]interface{}) (containers []*string, intContainers []*string) {
	if s, ok := resource["spec"]; ok {
		if sm, ok := s.(map[string]interface{}); ok {
			if t, ok := sm["template"]; ok {
				if tm, ok := t.(map[string]interface{}); ok {
					if tms, ok := tm["spec"]; ok {
						if tmsm, ok := tms.(map[string]interface{}); ok {
							if c, ok := tmsm["containers"]; ok {
								if cas, ok := c.([]interface{}); ok {
									for _, ca := range cas {
										if cam, ok := ca.(map[string]interface{}); ok {
											if n, ok := cam["name"]; ok {
												if cn, ok := n.(string); ok {
													containers = append(containers, &cn)
												}
											}
										}
									}
								}
							}
							///initContainers.name
							if c, ok := tmsm["initContainers"]; ok {
								if cas, ok := c.([]interface{}); ok {
									for _, ca := range cas {
										if cam, ok := ca.(map[string]interface{}); ok {
											if n, ok := cam["name"]; ok {
												if cn, ok := n.(string); ok {
													intContainers = append(intContainers, &cn)
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return containers, intContainers
}

func getResourceName(resource map[string]interface{}) string {
	if md, ok := resource["metadata"]; ok {
		if mdm, ok := md.(map[string]interface{}); ok {
			if h, ok := mdm["name"]; ok {
				if hs, ok := h.(string); ok {
					return hs
				}
			}
		}
	}
	return ""
}

func transform(resource v1alpha1.ResourceNode, name *string) *application2.ApplicationResourceRequest {
	resourceName := resource.Name
	kind := resource.Kind
	group := resource.Group
	version := resource.Version
	namespace := resource.Namespace
	request := &application2.ApplicationResourceRequest{
		Name:         name,
		ResourceName: &resourceName,
		Kind:         &kind,
		Group:        &group,
		Version:      &version,
		Namespace:    &namespace,
	}
	return request
}

func getJobsNewPods(jobManifest map[string]interface{}, podManifests []map[string]interface{}) (newPodNames map[string]bool) {
	newPodNames = make(map[string]bool, 0)
	for _, pod := range podManifests {
		newPodNames[getResourceName(pod)] = true
	}

	//TODO - new or old logic
	return
}

func updateMetadataOfDuplicatePods(podsMetadataFromPods []*argoApplication.PodMetadata, duplicatePodToReplicasetMapping map[string]string, podMetaData []*argoApplication.PodMetadata) []*argoApplication.PodMetadata {
	// Initialize mappings for containers
	containersPodMapping := make(map[string][]*string) // Mapping from pod name to container names
	initContainersPodMapping := make(map[string][]*string)
	// iterate over pod metadata extracted from pods' manifests
	for _, podMetadataFromPod := range podsMetadataFromPods {
		// If pod is not a duplicate
		if _, ok := duplicatePodToReplicasetMapping[podMetadataFromPod.Name]; !ok {
			// if pod is not a duplicate append pod metadata to the final result
			podMetaData = append(podMetaData, podMetadataFromPod)
		} else {
			// update init and sidecar container data into podsMetadataFromPods array's pods obj. if pod is a duplicate found in duplicatePodToReplicasetMapping,
			for _, podMetadataFromReplicaSet := range podMetaData {
				if podMetadataFromReplicaSet.Name == podMetadataFromPod.Name {
					// Update containers mapping
					if podMetadataFromPod.Containers != nil {
						containersPodMapping[podMetadataFromPod.Name] = podMetadataFromPod.Containers
						// Update containers mapping for other duplicate pods with the same replicaset
						// because we are only fetching manifest for one pod
						// and propagate to other pods having same parent
						currentPodParentName := duplicatePodToReplicasetMapping[podMetadataFromPod.Name]
						for podName, podParentName := range duplicatePodToReplicasetMapping {
							if podParentName == currentPodParentName {
								containersPodMapping[podName] = podMetadataFromPod.Containers
							}
						}
					}
					if podMetadataFromPod.InitContainers != nil {
						initContainersPodMapping[podMetadataFromPod.Name] = podMetadataFromPod.InitContainers
						currentPodParentName := duplicatePodToReplicasetMapping[podMetadataFromPod.Name]
						for podName, podParentName := range duplicatePodToReplicasetMapping {
							if podParentName == currentPodParentName {
								initContainersPodMapping[podName] = podMetadataFromPod.InitContainers
							}
						}
					}
				}
			}
		}
	}

	// Update pod metadata with containers mapping
	for _, metadata := range podMetaData {
		if containers, ok := containersPodMapping[metadata.Name]; ok {
			metadata.Containers = containers
		}
		if initContainers, ok := initContainersPodMapping[metadata.Name]; ok {
			metadata.InitContainers = initContainers
		}
	}
	// Return updated pod metadata
	return podMetaData
}

// fill the health status in node from app resources
func updateNodeHealthStatus(resp *v1alpha1.ApplicationTree, application *v1alpha1.Application) *v1alpha1.ApplicationTree {
	if resp == nil || len(resp.Nodes) == 0 || len(application.Status.Resources) == 0 {
		return resp
	}

	for index, node := range resp.Nodes {
		if node.Health != nil {
			continue
		}
		for _, resource := range application.Status.Resources {
			if node.Group != resource.Group || node.Version != resource.Version || node.Kind != resource.Kind ||
				node.Name != resource.Name || node.Namespace != resource.Namespace {
				continue
			}
			resourceHealth := resource.Health
			if resourceHealth != nil {
				node.Health = &v1alpha1.HealthStatus{
					Message: resourceHealth.Message,
					Status:  resourceHealth.Status,
				}
				// updating the element in slice
				// https://medium.com/@xcoulon/3-ways-to-update-elements-in-a-slice-d5df54c9b2f8
				resp.Nodes[index] = node
			}
			break
		}
	}
	return resp
}
