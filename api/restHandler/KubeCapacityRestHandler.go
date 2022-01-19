/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package restHandler

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/restHandler/common"
	"github.com/devtron-labs/devtron/util"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

type KubeCapacityRestHandler interface {
	KubeCapacityDefault(w http.ResponseWriter, r *http.Request)
	KubeCapacityPods(w http.ResponseWriter, r *http.Request)
	KubeCapacityUtilization(w http.ResponseWriter, r *http.Request)
	AvailableResources(w http.ResponseWriter, r *http.Request)
	PodsAndUtil(w http.ResponseWriter, r *http.Request)
	GetNodes(w http.ResponseWriter, r *http.Request)
	GetPodsOfNodes(w http.ResponseWriter, r *http.Request)
}
type KubeCapacityRestHandlerImpl struct {
	logger *zap.SugaredLogger
}

func NewKubeCapacityRestHandlerImpl(logger *zap.SugaredLogger) *KubeCapacityRestHandlerImpl {
	return &KubeCapacityRestHandlerImpl{
		logger: logger,
	}
}

const (
	KubeCapacity            = "kube-capacity --output json"
	KubeCapacityPods        = "kube-capacity --pods --output json"
	KubeCapacityUtilization = "kube-capacity --util --output json"
	AvailableResources      = "kube-capacity --available --output json"
	PodsAndUtil             = "kube-capacity --pods --util --output json"
)

func (impl KubeCapacityRestHandlerImpl) KubeCapacityDefault(w http.ResponseWriter, r *http.Request) {
	res, err := util.KubeCapacity(KubeCapacity)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}

	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (impl KubeCapacityRestHandlerImpl) KubeCapacityPods(w http.ResponseWriter, r *http.Request) {
	res, err := util.KubeCapacity(KubeCapacityPods)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (impl KubeCapacityRestHandlerImpl) KubeCapacityUtilization(w http.ResponseWriter, r *http.Request) {
	res, err := util.KubeCapacity(KubeCapacityUtilization)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

func (impl KubeCapacityRestHandlerImpl) AvailableResources(w http.ResponseWriter, r *http.Request) {
	res, err := util.KubeCapacity(AvailableResources)
	if err != nil {
		impl.logger.Errorw("err in execute command, PodsAndUtil", "err: ", err)
		common.WriteJsonResp(w, err, nil, http.StatusInternalServerError)
	}
	common.WriteJsonResp(w, nil, res, http.StatusOK)
}

type listPrinter struct {
	cm     *clusterMetric
	sortBy string
}

type nodespods struct{
	podlist *corev1.PodList
	nodelist *corev1.NodeList
	cm json.RawMessage

}

type getNodesValues struct{
	Name 				string
	NameSpace			string
	ClusterName			string
	Labels          	map[string]string
	Annotations     	map[string]string
	CreationTimestamp	Time
	Capacity 			ResourceList
	Allocatable 		ResourceList
	InternalIP			string
	ExternalIP			string
}

type Time struct {
	time.Time `protobuf:"-"`
}

type ResourceList map[ResourceName]resource.Quantity
type ResourceName string

func (impl KubeCapacityRestHandlerImpl) GetNodes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	appId, err := strconv.Atoi(vars["Id"])
	clientset, err := NewClientSet()
	if err != nil {
		fmt.Printf("Error connecting to Kubernetes: %v\n", err)
		os.Exit(1)
	}
	nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: ""})
	if err != nil {
		fmt.Printf("Error listing Nodes: %v\n", err)
		os.Exit(2)
	}
	nodes := nodeList.Items[appId]
	nodeValues := &getNodesValues {
		Name:              nodes.Name,
		NameSpace:         nodes.Namespace,
		ClusterName:       nodes.ClusterName,
		Annotations:       nodes.Annotations,
		Labels:            nodes.Labels,
		CreationTimestamp: Time(nodes.CreationTimestamp),
		InternalIP:        nodes.Status.Addresses[0].Address,
		ExternalIP:        nodes.Status.Addresses[1].Address,
	}
	common.WriteJsonResp(w, err, nodeValues, http.StatusOK)

}
func (impl KubeCapacityRestHandlerImpl) GetPodsOfNodes(w http.ResponseWriter, r *http.Request) {

	clientset, err := NewClientSet()
	if err != nil {
		fmt.Printf("Error connecting to Kubernetes: %v\n", err)
		os.Exit(1)
	}
	podList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: ""})
	if err != nil {
		fmt.Printf("Error listing Pods: %v\n", err)
		os.Exit(3)
	}
	var x []interface{}
	for _, nodes := range podList.Items {
		if nodes.Spec.NodeName == "ip-172-31-181-73.us-east-2.compute.internal" {
			nodeValues := &getNodesValues{
				Name:              nodes.Name,
				NameSpace:         nodes.Namespace,
				ClusterName:       nodes.ClusterName,
				Annotations:       nodes.Annotations,
				Labels:            nodes.Labels,
				CreationTimestamp: Time(nodes.CreationTimestamp),
			}

			x = append(x, nodeValues)
		}
	}
	common.WriteJsonResp(w, err, x, http.StatusOK)

}

func (impl KubeCapacityRestHandlerImpl) PodsAndUtil(w http.ResponseWriter, r *http.Request) {
	clientset, err := NewClientSet()
	if err != nil {
		fmt.Printf("Error connecting to Kubernetes: %v\n", err)
		os.Exit(1)
	}
	podList, nodeList := getPodsAndNodes(clientset)

	cm := buildClusterMetric(podList, nodeList)
	for i, v := range cm.nodeMetrics {
		fmt.Println(i)
		fmt.Println(v)
		fmt.Println("cpu")
		fmt.Println(v.cpu.limit)
	}
	lp := &listPrinter{
		cm:     &cm,
		sortBy: "",
	}

	listOutput := lp.buildListClusterMetrics()
	jsonRaw, err := json.MarshalIndent(listOutput, "", "  ")
	var dat json.RawMessage
	err = json.Unmarshal(jsonRaw,&dat)
	if err != nil{
		common.WriteJsonResp(w, err, dat, http.StatusOK)
	}
	nodepod := &nodespods{
		nodelist: nodeList,
		podlist: podList,
		cm: dat,
	}
	fmt.Println()
	sortedNodeMetrics := make([]*nodeMetric, len(cm.nodeMetrics))
	i := 0
	for name := range cm.nodeMetrics {
		sortedNodeMetrics[i] = cm.nodeMetrics[name]
		i++
	}
	fmt.Println(nodepod)
	common.WriteJsonResp(w, nil, podList, http.StatusOK)
}

// NewClientSet returns a new Kubernetes clientset
func NewClientSet() (*kubernetes.Clientset, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func getKubeConfig() (*rest.Config, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{CurrentContext: ""},
	).ClientConfig()
}

func getPodsAndNodes(clientset kubernetes.Interface) (*corev1.PodList, *corev1.NodeList) {
	nodeList, err := clientset.CoreV1().Nodes().List(metav1.ListOptions{LabelSelector: ""})
	if err != nil {
		fmt.Printf("Error listing Nodes: %v\n", err)
		os.Exit(2)
	}
	podList, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{LabelSelector: ""})
	if err != nil {
		fmt.Printf("Error listing Pods: %v\n", err)
		os.Exit(3)
	}
	return podList, nodeList
}

type clusterMetric struct {
	cpu         *resourceMetric
	memory      *resourceMetric
	nodeMetrics map[string]*nodeMetric
}

type resourceMetric struct {
	resourceType string
	allocatable  resource.Quantity
	utilization  resource.Quantity
	request      resource.Quantity
	limit        resource.Quantity
}

type nodeMetric struct {
	name       string
	cpu        *resourceMetric
	memory     *resourceMetric
	podMetrics map[string]*podMetric
}

type podMetric struct {
	name             string
	namespace        string
	cpu              *resourceMetric
	memory           *resourceMetric
	containerMetrics map[string]*containerMetric
}

type containerMetric struct {
	name   string
	cpu    *resourceMetric
	memory *resourceMetric
}

func buildClusterMetric(podList *corev1.PodList, nodeList *corev1.NodeList) clusterMetric {
	cm := clusterMetric{
		cpu:         &resourceMetric{resourceType: "cpu"},
		memory:      &resourceMetric{resourceType: "memory"},
		nodeMetrics: map[string]*nodeMetric{},
	}

	for _, node := range nodeList.Items {
		cm.nodeMetrics[node.Name] = &nodeMetric{
			name: node.Name,
			cpu: &resourceMetric{
				resourceType: "cpu",
				allocatable:  node.Status.Allocatable["cpu"],
			},
			memory: &resourceMetric{
				resourceType: "memory",
				allocatable:  node.Status.Allocatable["memory"],
			},
			podMetrics: map[string]*podMetric{},
		}
	}

	for _, pod := range podList.Items {
		if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed {
			cm.addPodMetric(&pod)
		}
	}
	for _, node := range nodeList.Items {
		nm := cm.nodeMetrics[node.Name]
		fmt.Println(nm)
		if nm != nil {
			cm.addNodeMetric(cm.nodeMetrics[node.Name])
		}
	}

	return cm
}

type listClusterMetrics struct {
	Nodes         []*listNodeMetric  `json:"nodes"`
	ClusterTotals *listClusterTotals `json:"clusterTotals"`
}

type listClusterTotals struct {
	CPU    *listResourceOutput `json:"cpu"`
	Memory *listResourceOutput `json:"memory"`
}

type listNodeMetric struct {
	Name   string              `json:"name"`
	CPU    *listResourceOutput `json:"cpu,omitempty"`
	Memory *listResourceOutput `json:"memory,omitempty"`
	Pods   []*listPod          `json:"pods,omitempty"`
}
type listPod struct {
	Name       string              `json:"name"`
	Namespace  string              `json:"namespace"`
	CPU        *listResourceOutput `json:"cpu"`
	Memory     *listResourceOutput `json:"memory"`
	Containers []listContainer     `json:"containers,omitempty"`
}
type listContainer struct {
	Name   string              `json:"name"`
	CPU    *listResourceOutput `json:"cpu"`
	Memory *listResourceOutput `json:"memory"`
}

type listResourceOutput struct {
	Requests       string `json:"requests"`
	RequestsPct    string `json:"requestsPercent"`
	Limits         string `json:"limits"`
	LimitsPct      string `json:"limitsPercent"`
	Utilization    string `json:"utilization,omitempty"`
	UtilizationPct string `json:"utilizationPercent,omitempty"`
}

func (lp *listPrinter) buildListClusterMetrics() listClusterMetrics {
	var response listClusterMetrics

	response.ClusterTotals = &listClusterTotals{
		CPU:    lp.buildListResourceOutput(lp.cm.cpu),
		Memory: lp.buildListResourceOutput(lp.cm.memory),
	}

	for _, nodeMetric := range lp.cm.getSortedNodeMetrics(lp.sortBy) {
		var node listNodeMetric
		node.Name = nodeMetric.name
		node.CPU = lp.buildListResourceOutput(nodeMetric.cpu)
		node.Memory = lp.buildListResourceOutput(nodeMetric.memory)

		for _, podMetric := range nodeMetric.getSortedPodMetrics(lp.sortBy) {
			var pod listPod
			pod.Name = podMetric.name
			pod.Namespace = podMetric.namespace
			pod.CPU = lp.buildListResourceOutput(podMetric.cpu)
			pod.Memory = lp.buildListResourceOutput(podMetric.memory)

			for _, containerMetric := range podMetric.getSortedContainerMetrics(lp.sortBy) {
				pod.Containers = append(pod.Containers, listContainer{
					Name:   containerMetric.name,
					Memory: lp.buildListResourceOutput(containerMetric.memory),
					CPU:    lp.buildListResourceOutput(containerMetric.cpu),
				})
			}
			node.Pods = append(node.Pods, &pod)
		}
		response.Nodes = append(response.Nodes, &node)
	}

	return response
}
func (lp *listPrinter) buildListResourceOutput(item *resourceMetric) *listResourceOutput {
	valueCalculator := item.valueFunction()
	percentCalculator := item.percentFunction()

	out := listResourceOutput{
		Requests:    valueCalculator(item.request),
		RequestsPct: percentCalculator(item.request),
		Limits:      valueCalculator(item.limit),
		LimitsPct:   percentCalculator(item.limit),
	}

	out.Utilization = valueCalculator(item.utilization)
	out.UtilizationPct = percentCalculator(item.utilization)
	return &out
}

func (nm *nodeMetric) getSortedPodMetrics(sortBy string) []*podMetric {
	sortedPodMetrics := make([]*podMetric, len(nm.podMetrics))

	i := 0
	for name := range nm.podMetrics {
		sortedPodMetrics[i] = nm.podMetrics[name]
		i++
	}

	sort.Slice(sortedPodMetrics, func(i, j int) bool {
		m1 := sortedPodMetrics[i]
		m2 := sortedPodMetrics[j]

		switch sortBy {
		case "cpu.util":
			return m2.cpu.utilization.MilliValue() < m1.cpu.utilization.MilliValue()
		case "cpu.limit":
			return m2.cpu.limit.MilliValue() < m1.cpu.limit.MilliValue()
		case "cpu.request":
			return m2.cpu.request.MilliValue() < m1.cpu.request.MilliValue()
		case "mem.util":
			return m2.memory.utilization.Value() < m1.memory.utilization.Value()
		case "mem.limit":
			return m2.memory.limit.Value() < m1.memory.limit.Value()
		case "mem.request":
			return m2.memory.request.Value() < m1.memory.request.Value()
		default:
			return m1.name < m2.name
		}
	})

	return sortedPodMetrics
}

func (pm *podMetric) getSortedContainerMetrics(sortBy string) []*containerMetric {
	sortedContainerMetrics := make([]*containerMetric, len(pm.containerMetrics))

	i := 0
	for name := range pm.containerMetrics {
		sortedContainerMetrics[i] = pm.containerMetrics[name]
		i++
	}

	sort.Slice(sortedContainerMetrics, func(i, j int) bool {
		m1 := sortedContainerMetrics[i]
		m2 := sortedContainerMetrics[j]

		switch sortBy {
		case "cpu.util":
			return m2.cpu.utilization.MilliValue() < m1.cpu.utilization.MilliValue()
		case "cpu.limit":
			return m2.cpu.limit.MilliValue() < m1.cpu.limit.MilliValue()
		case "cpu.request":
			return m2.cpu.request.MilliValue() < m1.cpu.request.MilliValue()
		case "mem.util":
			return m2.memory.utilization.Value() < m1.memory.utilization.Value()
		case "mem.limit":
			return m2.memory.limit.Value() < m1.memory.limit.Value()
		case "mem.request":
			return m2.memory.request.Value() < m1.memory.request.Value()
		default:
			return m1.name < m2.name
		}
	})

	return sortedContainerMetrics
}

func (cm *clusterMetric) getSortedNodeMetrics(sortBy string) []*nodeMetric {
	sortedNodeMetrics := make([]*nodeMetric, len(cm.nodeMetrics))

	i := 0
	for name := range cm.nodeMetrics {
		sortedNodeMetrics[i] = cm.nodeMetrics[name]
		i++
	}

	sort.Slice(sortedNodeMetrics, func(i, j int) bool {
		m1 := sortedNodeMetrics[i]
		m2 := sortedNodeMetrics[j]

		switch sortBy {
		case "cpu.util":
			return m2.cpu.utilization.MilliValue() < m1.cpu.utilization.MilliValue()
		case "cpu.limit":
			return m2.cpu.limit.MilliValue() < m1.cpu.limit.MilliValue()
		case "cpu.request":
			return m2.cpu.request.MilliValue() < m1.cpu.request.MilliValue()
		case "mem.util":
			return m2.memory.utilization.Value() < m1.memory.utilization.Value()
		case "mem.limit":
			return m2.memory.limit.Value() < m1.memory.limit.Value()
		case "mem.request":
			return m2.memory.request.Value() < m1.memory.request.Value()
		default:
			return m1.name < m2.name
		}
	})

	return sortedNodeMetrics
}
func (rm resourceMetric) valueFunction() (f func(r resource.Quantity) string) {
	switch rm.resourceType {
	case "cpu":
		f = func(r resource.Quantity) string {
			return fmt.Sprintf("%s", r.String())
		}
	case "memory":
		f = func(r resource.Quantity) string {
			return fmt.Sprintf("%s", r.String())
		}
	}
	return f
}

// NOTE: This might not be a great place for closures due to the cyclical nature of how resourceType works. Perhaps better implemented another way.
func (rm resourceMetric) percentFunction() (f func(r resource.Quantity) string) {
	f = func(r resource.Quantity) string {
		return fmt.Sprintf("%v%%", int64(float64(r.MilliValue())/float64(rm.allocatable.MilliValue())*100))
	}
	return f
}

func PodRequestsAndLimits(pod *corev1.Pod) (reqs, limits corev1.ResourceList) {
	reqs, limits = corev1.ResourceList{}, corev1.ResourceList{}
	for _, container := range pod.Spec.Containers {
		addResourceList(reqs, container.Resources.Requests)
		addResourceList(limits, container.Resources.Limits)
	}
	// init containers define the minimum of any resource
	for _, container := range pod.Spec.InitContainers {
		maxResourceList(reqs, container.Resources.Requests)
		maxResourceList(limits, container.Resources.Limits)
	}
	return
}

func addResourceList(list, new corev1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
		} else {
			value.Add(quantity)
			list[name] = value
		}
	}
}

// maxResourceList sets list to the greater of list/newList for every resource
// either list
func maxResourceList(list, new corev1.ResourceList) {
	for name, quantity := range new {
		if value, ok := list[name]; !ok {
			list[name] = quantity.DeepCopy()
			continue
		} else {
			if quantity.Cmp(value) > 0 {
				list[name] = quantity.DeepCopy()
			}
		}
	}
}

func (cm *clusterMetric) addPodMetric(pod *corev1.Pod) {
	req, limit := PodRequestsAndLimits(pod)
	//key := fmt.Sprintf("%s-%s", pod.Namespace, pod.Name)
	nm := cm.nodeMetrics[pod.Spec.NodeName]

	pm := &podMetric{
		name:      pod.Name,
		namespace: pod.Namespace,
		cpu: &resourceMetric{
			resourceType: "cpu",
			request:      req["cpu"],
			limit:        limit["cpu"],
		},
		memory: &resourceMetric{
			resourceType: "memory",
			request:      req["memory"],
			limit:        limit["memory"],
		},
		containerMetrics: map[string]*containerMetric{},
	}

	for _, container := range pod.Spec.Containers {
		pm.containerMetrics[container.Name] = &containerMetric{
			name: container.Name,
			cpu: &resourceMetric{
				resourceType: "cpu",
				request:      container.Resources.Requests["cpu"],
				limit:        container.Resources.Limits["cpu"],
			},
			memory: &resourceMetric{
				resourceType: "memory",
				request:      container.Resources.Requests["memory"],
				limit:        container.Resources.Limits["memory"],
			},
		}
	}
	if nm != nil {

		nm.cpu.request.Add(req["cpu"])
		nm.cpu.limit.Add(limit["cpu"])
		nm.memory.request.Add(req["memory"])
		nm.memory.limit.Add(limit["memory"])
	}
}

func (cm *clusterMetric) addNodeMetric(nm *nodeMetric) {
	cm.cpu.addMetric(nm.cpu)
	cm.memory.addMetric(nm.memory)
}

func (rm *resourceMetric) addMetric(m *resourceMetric) {
	rm.allocatable.Add(m.allocatable)
	rm.utilization.Add(m.utilization)
	rm.request.Add(m.request)
	rm.limit.Add(m.limit)
}
