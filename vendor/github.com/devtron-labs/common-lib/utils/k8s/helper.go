/*
 * Copyright (c) 2024. Devtron Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package k8s

import (
	"context"
	"encoding/json"
	errors "errors"
	"fmt"
	"github.com/devtron-labs/common-lib/utils/k8s/commonBean"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	v14 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	v13 "k8s.io/api/policy/v1"
	v1beta12 "k8s.io/api/policy/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func ServerResourceForGroupVersionKind(discoveryClient discovery.DiscoveryInterface, gvk schema.GroupVersionKind) (*metav1.APIResource, error) {
	resources, err := discoveryClient.ServerResourcesForGroupVersion(gvk.GroupVersion().String())
	if err != nil {
		return nil, err
	}
	for _, r := range resources.APIResources {
		if r.Kind == gvk.Kind {
			return &r, nil
		}
	}
	return nil, k8sErrors.NewNotFound(schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind}, "")
}

func isServiceAccountTokenSecret(un *unstructured.Unstructured) (bool, metav1.OwnerReference) {
	ref := metav1.OwnerReference{
		APIVersion: "v1",
		Kind:       commonBean.ServiceAccountKind,
	}

	if typeVal, ok, err := unstructured.NestedString(un.Object, "type"); !ok || err != nil || typeVal != "kubernetes.io/service-account-token" {
		return false, ref
	}

	annotations := un.GetAnnotations()
	if annotations == nil {
		return false, ref
	}

	id, okId := annotations["kubernetes.io/service-account.uid"]
	name, okName := annotations["kubernetes.io/service-account.name"]
	if okId && okName {
		ref.Name = name
		ref.UID = types.UID(id)
	}
	return ref.Name != "" && ref.UID != "", ref
}

func ResolveResourceReferences(un *unstructured.Unstructured) ([]metav1.OwnerReference, func(ResourceKey) bool) {
	var isInferredParentOf func(_ ResourceKey) bool
	ownerRefs := un.GetOwnerReferences()
	gvk := un.GroupVersionKind()

	switch {

	// Special case for endpoint. Remove after https://github.com/kubernetes/kubernetes/issues/28483 is fixed
	case gvk.Group == "" && gvk.Kind == commonBean.EndpointsKind && len(un.GetOwnerReferences()) == 0:
		ownerRefs = append(ownerRefs, metav1.OwnerReference{
			Name:       un.GetName(),
			Kind:       commonBean.ServiceKind,
			APIVersion: "v1",
		})

	// Special case for Operator Lifecycle Manager ClusterServiceVersion:
	case un.GroupVersionKind().Group == "operators.coreos.com" && un.GetKind() == "ClusterServiceVersion":
		if un.GetAnnotations()["olm.operatorGroup"] != "" {
			ownerRefs = append(ownerRefs, metav1.OwnerReference{
				Name:       un.GetAnnotations()["olm.operatorGroup"],
				Kind:       "OperatorGroup",
				APIVersion: "operators.coreos.com/v1",
			})
		}

	// Edge case: consider auto-created service account tokens as a child of service account objects
	case un.GetKind() == commonBean.SecretKind && un.GroupVersionKind().Group == "":
		if yes, ref := isServiceAccountTokenSecret(un); yes {
			ownerRefs = append(ownerRefs, ref)
		}

	case (un.GroupVersionKind().Group == "apps" || un.GroupVersionKind().Group == "extensions") && un.GetKind() == commonBean.StatefulSetKind:
		if refs, err := isStatefulSetChild(un); err != nil {
			fmt.Println("error")
		} else {
			isInferredParentOf = refs
		}
	}

	return ownerRefs, isInferredParentOf
}

func isStatefulSetChild(un *unstructured.Unstructured) (func(ResourceKey) bool, error) {
	sts := v14.StatefulSet{}
	data, err := json.Marshal(un)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &sts)
	if err != nil {
		return nil, err
	}

	templates := sts.Spec.VolumeClaimTemplates
	return func(key ResourceKey) bool {
		if key.Kind == commonBean.PersistentVolumeClaimKind && key.GroupKind().Group == "" {
			for _, templ := range templates {
				if strings.HasPrefix(key.Name, fmt.Sprintf("%s-%s-", templ.Name, un.GetName())) {
					return true
				}
			}
		}
		return false
	}, nil
}

func CheckIfValidLabel(labelKey string, labelValue string) error {
	labelKey = strings.TrimSpace(labelKey)
	labelValue = strings.TrimSpace(labelValue)

	errs := validation.IsQualifiedName(labelKey)
	if len(labelKey) == 0 || len(errs) > 0 {
		return errors.New(fmt.Sprintf("Validation error - label key - %s is not satisfying the label key criteria", labelKey))
	}

	errs = validation.IsValidLabelValue(labelValue)
	if len(labelValue) == 0 || len(errs) > 0 {
		return errors.New(fmt.Sprintf("Validation error - label value - %s is not satisfying the label value criteria for label key - %s", labelValue, labelKey))
	}
	return nil
}

// DeletePod will delete the given pod, or return an error if it couldn't
func DeletePod(pod v1.Pod, k8sClientSet *kubernetes.Clientset, deleteOptions metav1.DeleteOptions) error {
	return k8sClientSet.CoreV1().Pods(pod.Namespace).Delete(context.Background(), pod.Name, deleteOptions)
}

// EvictPod will evict the given pod, or return an error if it couldn't
func EvictPod(pod v1.Pod, k8sClientSet *kubernetes.Clientset, evictionGroupVersion schema.GroupVersion, deleteOptions metav1.DeleteOptions) error {
	switch evictionGroupVersion {
	case v13.SchemeGroupVersion:
		// send policy/v1 if the server supports it
		eviction := &v13.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &deleteOptions,
		}
		return k8sClientSet.PolicyV1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)

	default:
		// otherwise, fall back to policy/v1beta1, supported by all servers that support the eviction subresource
		eviction := &v1beta12.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
			DeleteOptions: &deleteOptions,
		}
		return k8sClientSet.PolicyV1beta1().Evictions(eviction.Namespace).Evict(context.TODO(), eviction)
	}
}

// CheckEvictionSupport uses Discovery API to find out if the server support
// eviction subresource If support, it will return its groupVersion; Otherwise,
// it will return an empty GroupVersion
func CheckEvictionSupport(clientset kubernetes.Interface) (schema.GroupVersion, error) {
	discoveryClient := clientset.Discovery()

	// version info available in subresources since v1.8.0 in https://github.com/kubernetes/kubernetes/pull/49971
	resourceList, err := discoveryClient.ServerResourcesForGroupVersion("v1")
	if err != nil {
		return schema.GroupVersion{}, err
	}
	for _, resource := range resourceList.APIResources {
		if resource.Name == commonBean.EvictionSubresource && resource.Kind == commonBean.EvictionKind &&
			len(resource.Group) > 0 && len(resource.Version) > 0 {
			return schema.GroupVersion{Group: resource.Group, Version: resource.Version}, nil
		}
	}
	return schema.GroupVersion{}, nil
}

func UpdateNodeUnschedulableProperty(desiredUnschedulable bool, node *v1.Node, k8sClientSet *kubernetes.Clientset) (*v1.Node, error) {
	node.Spec.Unschedulable = desiredUnschedulable
	node, err := k8sClientSet.CoreV1().Nodes().Update(context.Background(), node, metav1.UpdateOptions{})
	return node, err
}

func OverrideK8sHttpClientWithTracer(restConfig *rest.Config) (*http.Client, error) {
	httpClientFor, err := rest.HTTPClientFor(restConfig)
	if err != nil {
		fmt.Println("error occurred while overriding k8s client", "reason", err)
		return nil, err
	}
	httpClientFor.Transport = otelhttp.NewTransport(httpClientFor.Transport)
	return httpClientFor, nil
}

func GetServerNameFromServerUrl(serverURL string) (string, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", err
	}

	host := u.Host
	if strings.Contains(host, ":") {
		host, _, err = net.SplitHostPort(u.Host)
		if err != nil {
			return "", err
		}
	}
	return host, nil
}
