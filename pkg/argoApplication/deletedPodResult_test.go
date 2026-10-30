package argoApplication

import (
	"errors"
	"reflect"
	"testing"

	application2 "github.com/argoproj/argo-cd/v3/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func resourceResult(group, kind, ns, name string, err error) *argoApplication.Result {
	return &argoApplication.Result{
		Request: &application2.ApplicationResourceRequest{Group: &group, Kind: &kind, Namespace: &ns, ResourceName: &name},
		Error:   err,
	}
}

func notFound(name string) error {
	return status.Errorf(codes.NotFound, "error getting resource: pods %q not found", name)
}

func TestIsDeletedPodResult(t *testing.T) {
	cases := []struct {
		name   string
		result *argoApplication.Result
		want   bool
	}{
		{"nil result", nil, false},
		{"no error", resourceResult("", "Pod", "ns", "p", nil), false},
		{"nil request", &argoApplication.Result{Error: notFound("p")}, false},
		{"pod NotFound", resourceResult("", "Pod", "ns", "p", notFound("p")), true},
		{"pod kind in another group", resourceResult("metrics.k8s.io", "Pod", "ns", "p", notFound("p")), false},
		{"replicaset NotFound", resourceResult("apps", "ReplicaSet", "ns", "rs", notFound("rs")), false},
		{"controllerrevision NotFound", resourceResult("apps", "ControllerRevision", "ns", "cr", notFound("cr")), false},
		{"pod InvalidArgument", resourceResult("", "Pod", "ns", "p", status.Error(codes.InvalidArgument, "not found as part of application")), false},
		{"pod PermissionDenied", resourceResult("", "Pod", "ns", "p", status.Error(codes.PermissionDenied, "denied")), false},
		{"pod DeadlineExceeded", resourceResult("", "Pod", "ns", "p", status.Error(codes.DeadlineExceeded, "timeout")), false},
		{"pod Unavailable", resourceResult("", "Pod", "ns", "p", status.Error(codes.Unavailable, "unavailable")), false},
		{"pod Unknown with not found text", resourceResult("", "Pod", "ns", "p", status.Error(codes.Unknown, `pods "p" not found`)), false},
		{"pod Canceled", resourceResult("", "Pod", "ns", "p", status.Error(codes.Canceled, "canceled")), false},
		{"pod plain error", resourceResult("", "Pod", "ns", "p", errors.New(`pods "p" not found`)), false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isDeletedPodResult(c.result); got != c.want {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
}

func node(group, kind, ns, name string) v1alpha1.ResourceNode {
	return v1alpha1.ResourceNode{ResourceRef: v1alpha1.ResourceRef{Group: group, Kind: kind, Namespace: ns, Name: name}}
}

func names(nodes []v1alpha1.ResourceNode) []string {
	out := make([]string, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, n.Kind+"/"+n.Namespace+"/"+n.Name)
	}
	return out
}

func TestRemoveDeletedPodNodes(t *testing.T) {
	tree := []v1alpha1.ResourceNode{
		node("apps", "DaemonSet", "ns", "ds"),
		node("", "Pod", "ns", "live"),
		node("", "Pod", "ns", "gone"),
		node("", "Pod", "other", "gone"),
		node("", "Service", "ns", "gone"),
		node("", "Pod", "ns", "gone-2"),
	}
	cases := []struct {
		name    string
		results []*argoApplication.Result
		want    []string
	}{
		{"no results", nil, names(tree)},
		{"only successful results", []*argoApplication.Result{resourceResult("", "Pod", "ns", "live", nil)}, names(tree)},
		{"non NotFound error keeps pod", []*argoApplication.Result{resourceResult("", "Pod", "ns", "gone", status.Error(codes.DeadlineExceeded, "timeout"))}, names(tree)},
		{"NotFound pod removed, same name in other ns and non-pod kept",
			[]*argoApplication.Result{resourceResult("", "Pod", "ns", "gone", notFound("gone"))},
			[]string{"DaemonSet/ns/ds", "Pod/ns/live", "Pod/other/gone", "Service/ns/gone", "Pod/ns/gone-2"}},
		{"several removed, order kept",
			[]*argoApplication.Result{resourceResult("", "Pod", "ns", "gone-2", notFound("gone-2")), resourceResult("", "Pod", "other", "gone", notFound("gone"))},
			[]string{"DaemonSet/ns/ds", "Pod/ns/live", "Pod/ns/gone", "Service/ns/gone"}},
		{"duplicate results for one pod",
			[]*argoApplication.Result{resourceResult("", "Pod", "ns", "gone", notFound("gone")), resourceResult("", "Pod", "ns", "gone", notFound("gone"))},
			[]string{"DaemonSet/ns/ds", "Pod/ns/live", "Pod/other/gone", "Service/ns/gone", "Pod/ns/gone-2"}},
		{"NotFound on non-pod kind ignored", []*argoApplication.Result{resourceResult("apps", "DaemonSet", "ns", "ds", notFound("ds"))}, names(tree)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := names(removeDeletedPodNodes(tree, c.results)); !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %v, want %v", got, c.want)
			}
		})
	}
	if got := removeDeletedPodNodes(nil, []*argoApplication.Result{resourceResult("", "Pod", "ns", "gone", notFound("gone"))}); len(got) != 0 {
		t.Errorf("nil nodes: got %v", got)
	}
}
