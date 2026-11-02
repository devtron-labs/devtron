package argoApplication

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"

	application2 "github.com/argoproj/argo-cd/v3/pkg/apiclient/application"
	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/common-lib/async"
	"github.com/devtron-labs/common-lib/constants"
	"github.com/devtron-labs/devtron/client/argocdServer"
	argoApplication "github.com/devtron-labs/devtron/client/argocdServer/bean"
	"github.com/devtron-labs/devtron/pkg/argoApplication/bean"
	util3 "github.com/devtron-labs/devtron/pkg/util"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const testNs = "kube-system"

// fakeArgoServer serves argo's cached tree and live GETs per resource
type fakeArgoServer struct {
	application2.ApplicationServiceClient
	nodes   []v1alpha1.ResourceNode
	live    map[string]string // ns/kind/name -> manifest
	errs    map[string]error  // ns/kind/name -> live GET error
	treeErr error
}

func resKey(ns, kind, name string) string { return ns + "/" + kind + "/" + name }

func (f *fakeArgoServer) ResourceTree(ctx context.Context, in *application2.ResourcesQuery, opts ...grpc.CallOption) (*v1alpha1.ApplicationTree, error) {
	if f.treeErr != nil {
		return nil, f.treeErr
	}
	nodes := make([]v1alpha1.ResourceNode, len(f.nodes))
	copy(nodes, f.nodes)
	return &v1alpha1.ApplicationTree{Nodes: nodes}, nil
}

func (f *fakeArgoServer) GetResource(ctx context.Context, in *application2.ApplicationResourceRequest, opts ...grpc.CallOption) (*application2.ApplicationResourceResponse, error) {
	key := resKey(in.GetNamespace(), in.GetKind(), in.GetResourceName())
	if err, ok := f.errs[key]; ok {
		return nil, err
	}
	if m, ok := f.live[key]; ok {
		return &application2.ApplicationResourceResponse{Manifest: &m}, nil
	}
	// argocd-server maps k8s 404 to codes.NotFound
	return nil, status.Errorf(codes.NotFound, "error getting resource: pods %q not found", in.GetResourceName())
}

func (f *fakeArgoServer) Watch(ctx context.Context, in *application2.ApplicationQuery, opts ...grpc.CallOption) (application2.ApplicationService_WatchClient, error) {
	app := v1alpha1.Application{}
	app.Status.Health.Status = "Healthy"
	return &fakeWatch{app: app}, nil
}

type fakeWatch struct {
	application2.ApplicationService_WatchClient
	app v1alpha1.Application
}

func (w *fakeWatch) Recv() (*v1alpha1.ApplicationWatchEvent, error) {
	return &v1alpha1.ApplicationWatchEvent{Application: w.app}, nil
}

type fakeAcdWrapper struct {
	argocdServer.ArgoClientWrapperService
	asc application2.ApplicationServiceClient
}

func (w *fakeAcdWrapper) GetArgoClient(ctx context.Context) (application2.ApplicationServiceClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient("passthrough:///argocd-fake", grpc.WithTransportCredentials(insecure.NewCredentials()))
	return w.asc, conn, err
}

func newTestService(srv *fakeArgoServer) *ArgoApplicationServiceExtendedImpl {
	return &ArgoApplicationServiceExtendedImpl{
		ArgoApplicationServiceImpl: &ArgoApplicationServiceImpl{logger: zap.NewNop().Sugar()},
		aCDAuthConfig:              &util3.ACDAuthConfig{},
		acdClientWrapper:           &fakeAcdWrapper{asc: srv},
		asyncRunnable:              async.NewAsyncRunnable(zap.NewNop().Sugar(), constants.Orchestrator),
	}
}

func fetchTree(t *testing.T, srv *fakeArgoServer) *argoApplication.ResourceTreeResponse {
	t.Helper()
	appName := "efs-csi-dev"
	req := bean.NewImperativeQueryRequest(&application2.ResourcesQuery{ApplicationName: &appName}).WithArgoClusterId(1)
	resp, err := newTestService(srv).GetResourceTree(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	return resp
}

func ref(group, kind, ns, name, uid string) v1alpha1.ResourceRef {
	return v1alpha1.ResourceRef{Group: group, Version: "v1", Kind: kind, Namespace: ns, Name: name, UID: uid}
}

func treeNode(r v1alpha1.ResourceRef, labels map[string]string, parents ...v1alpha1.ResourceRef) v1alpha1.ResourceNode {
	n := v1alpha1.ResourceNode{ResourceRef: r, ParentRefs: parents}
	if labels != nil {
		n.NetworkingInfo = &v1alpha1.ResourceNetworkingInfo{Labels: labels}
	}
	return n
}

func podManifest(ns, name, hash string) string {
	return fmt.Sprintf(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":%q,"namespace":%q,"labels":{"controller-revision-hash":%q}},"spec":{"containers":[{"name":"efs-plugin"},{"name":"liveness-probe"}]}}`, name, ns, hash)
}

// daemonSetApp builds a DaemonSet tree in ns with the given pods; only livePods exist in k8s
func daemonSetApp(ns, dsName string, pods, livePods []string) *fakeArgoServer {
	hash := "7d9f8c6b5"
	ds := ref("apps", "DaemonSet", ns, dsName, "ds-"+ns+dsName)
	labels := map[string]string{"controller-revision-hash": hash}
	srv := &fakeArgoServer{
		nodes: []v1alpha1.ResourceNode{
			treeNode(ds, nil),
			treeNode(ref("apps", "ControllerRevision", ns, dsName+"-"+hash, "cr-"+ns+dsName), labels, ds),
		},
		live: map[string]string{
			resKey(ns, "DaemonSet", dsName):                   fmt.Sprintf(`{"apiVersion":"apps/v1","kind":"DaemonSet","metadata":{"name":%q}}`, dsName),
			resKey(ns, "ControllerRevision", dsName+"-"+hash): fmt.Sprintf(`{"apiVersion":"apps/v1","kind":"ControllerRevision","revision":3,"metadata":{"name":"%s-%s","labels":{"controller-revision-hash":%q}}}`, dsName, hash, hash),
		},
		errs: map[string]error{},
	}
	for _, p := range pods {
		srv.nodes = append(srv.nodes, treeNode(ref("", "Pod", ns, p, "uid-"+ns+p), labels, ds))
	}
	for _, p := range livePods {
		srv.live[resKey(ns, "Pod", p)] = podManifest(ns, p, hash)
	}
	return srv
}

func merge(a, b *fakeArgoServer) *fakeArgoServer {
	a.nodes = append(a.nodes, b.nodes...)
	for k, v := range b.live {
		a.live[k] = v
	}
	for k, v := range b.errs {
		a.errs[k] = v
	}
	return a
}

func podNames(resp *argoApplication.ResourceTreeResponse) []string {
	var names []string
	for _, n := range resp.Nodes {
		if n.Kind == "Pod" {
			names = append(names, n.Namespace+"/"+n.Name)
		}
	}
	sort.Strings(names)
	return names
}

func podMetadataNames(resp *argoApplication.ResourceTreeResponse) []string {
	var names []string
	for _, m := range resp.PodMetadata {
		names = append(names, m.Name)
	}
	sort.Strings(names)
	return names
}

func nonPodNames(resp *argoApplication.ResourceTreeResponse) []string {
	var names []string
	for _, n := range resp.Nodes {
		if n.Kind != "Pod" {
			names = append(names, n.Kind+"/"+n.Name)
		}
	}
	sort.Strings(names)
	return names
}

func assertStrings(t *testing.T, what string, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s:\n got  %v\n want %v", what, got, want)
	}
}

// pods argo still lists but k8s deleted are dropped; live, timed-out and non-pod nodes stay
func TestGetResourceTree_DaemonSetDeletedPodsRemoved(t *testing.T) {
	srv := daemonSetApp(testNs, "efs-csi-node",
		[]string{"efs-csi-node-248h2", "efs-csi-node-24b7k", "efs-csi-node-26w65", "efs-csi-node-2fnwg", "efs-csi-node-slow1"},
		[]string{"efs-csi-node-248h2", "efs-csi-node-24b7k"})
	srv.errs[resKey(testNs, "Pod", "efs-csi-node-slow1")] = status.Error(codes.DeadlineExceeded, "context deadline exceeded")
	resp := fetchTree(t, srv)

	assertStrings(t, "pods", podNames(resp), []string{testNs + "/efs-csi-node-248h2", testNs + "/efs-csi-node-24b7k", testNs + "/efs-csi-node-slow1"})
	assertStrings(t, "podMetadata", podMetadataNames(resp), []string{"efs-csi-node-248h2", "efs-csi-node-24b7k", "efs-csi-node-slow1"})
	assertStrings(t, "non-pod nodes", nonPodNames(resp), []string{"ControllerRevision/efs-csi-node-7d9f8c6b5", "DaemonSet/efs-csi-node"})
	for _, m := range resp.PodMetadata {
		if m.Name == "efs-csi-node-slow1" {
			continue
		}
		if len(m.Containers) != 2 || !m.IsNew {
			t.Errorf("live pod %s metadata changed: containers=%d isNew=%v", m.Name, len(m.Containers), m.IsNew)
		}
	}
	if resp.Status != "Healthy" {
		t.Errorf("status changed: %s", resp.Status)
	}
}

func TestGetResourceTree_AllDaemonSetPodsDeleted(t *testing.T) {
	resp := fetchTree(t, daemonSetApp(testNs, "efs-csi-node", []string{"efs-csi-node-a", "efs-csi-node-b"}, nil))
	assertStrings(t, "pods", podNames(resp), nil)
	assertStrings(t, "podMetadata", podMetadataNames(resp), nil)
	assertStrings(t, "non-pod nodes", nonPodNames(resp), []string{"ControllerRevision/efs-csi-node-7d9f8c6b5", "DaemonSet/efs-csi-node"})
}

func TestGetResourceTree_StatefulSetDeletedPodRemoved(t *testing.T) {
	sts := ref("apps", "StatefulSet", testNs, "db", "sts-1")
	labels := map[string]string{"controller-revision-hash": "db-abc"}
	srv := &fakeArgoServer{
		nodes: []v1alpha1.ResourceNode{
			treeNode(sts, nil),
			treeNode(ref("", "Pod", testNs, "db-0", "p-0"), labels, sts),
			treeNode(ref("", "Pod", testNs, "db-1", "p-1"), labels, sts),
		},
		live: map[string]string{
			resKey(testNs, "StatefulSet", "db"): `{"apiVersion":"apps/v1","kind":"StatefulSet","metadata":{"name":"db"},"status":{"updateRevision":"db-abc"}}`,
			resKey(testNs, "Pod", "db-0"):       podManifest(testNs, "db-0", "db-abc"),
		},
		errs: map[string]error{},
	}
	resp := fetchTree(t, srv)
	assertStrings(t, "pods", podNames(resp), []string{testNs + "/db-0"})
	assertStrings(t, "podMetadata", podMetadataNames(resp), []string{"db-0"})
	if !resp.PodMetadata[0].IsNew || len(resp.PodMetadata[0].Containers) != 2 {
		t.Errorf("live sts pod metadata changed: %+v", resp.PodMetadata[0])
	}
}

func TestGetResourceTree_JobDeletedPodRemoved(t *testing.T) {
	cron := ref("batch", "CronJob", testNs, "backup", "cj-1")
	job := ref("batch", "Job", testNs, "backup-1", "job-1")
	srv := &fakeArgoServer{
		nodes: []v1alpha1.ResourceNode{
			treeNode(cron, nil),
			treeNode(job, nil, cron),
			treeNode(ref("", "Pod", testNs, "backup-1-live", "p-1"), nil, job),
			treeNode(ref("", "Pod", testNs, "backup-1-gone", "p-2"), nil, job),
		},
		live: map[string]string{
			resKey(testNs, "CronJob", "backup"):    `{"apiVersion":"batch/v1","kind":"CronJob","metadata":{"name":"backup"}}`,
			resKey(testNs, "Job", "backup-1"):      `{"apiVersion":"batch/v1","kind":"Job","metadata":{"name":"backup-1"}}`,
			resKey(testNs, "Pod", "backup-1-live"): podManifest(testNs, "backup-1-live", ""),
		},
		errs: map[string]error{},
	}
	resp := fetchTree(t, srv)
	assertStrings(t, "pods", podNames(resp), []string{testNs + "/backup-1-live"})
	assertStrings(t, "non-pod nodes", nonPodNames(resp), []string{"CronJob/backup", "Job/backup-1"})
}

func deploymentApp(livePods map[string]bool) *fakeArgoServer {
	dep := ref("apps", "Deployment", testNs, "efs-csi-controller", "dep-1")
	rs := ref("apps", "ReplicaSet", testNs, "efs-csi-controller-5f6d", "rs-1")
	srv := &fakeArgoServer{
		nodes: []v1alpha1.ResourceNode{treeNode(dep, nil), treeNode(rs, nil, dep)},
		live: map[string]string{
			resKey(testNs, "Deployment", "efs-csi-controller"):      `{"apiVersion":"apps/v1","kind":"Deployment","metadata":{"name":"efs-csi-controller"},"spec":{"template":{"spec":{"containers":[{"name":"efs-plugin"}]}}}}`,
			resKey(testNs, "ReplicaSet", "efs-csi-controller-5f6d"): `{"apiVersion":"apps/v1","kind":"ReplicaSet","metadata":{"name":"efs-csi-controller-5f6d"},"spec":{"template":{"spec":{"containers":[{"name":"efs-plugin"}]}}}}`,
		},
		errs: map[string]error{},
	}
	for _, p := range []string{"ctrl-a", "ctrl-b", "ctrl-c"} {
		srv.nodes = append(srv.nodes, treeNode(ref("", "Pod", testNs, p, "uid-"+p), nil, rs))
		if livePods[p] {
			srv.live[resKey(testNs, "Pod", p)] = fmt.Sprintf(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":%q},"spec":{"containers":[{"name":"efs-plugin"}]}}`, p)
		}
	}
	return srv
}

// deployment pods: only one pod per ReplicaSet is fetched live, so only that one can be detected
func TestGetResourceTree_DeploymentFetchedPodDeletedRemoved(t *testing.T) {
	resp := fetchTree(t, deploymentApp(map[string]bool{"ctrl-a": true, "ctrl-b": true}))
	assertStrings(t, "pods", podNames(resp), []string{testNs + "/ctrl-a", testNs + "/ctrl-b"})
	assertStrings(t, "podMetadata", podMetadataNames(resp), []string{"ctrl-a", "ctrl-b"})
	assertStrings(t, "non-pod nodes", nonPodNames(resp), []string{"Deployment/efs-csi-controller", "ReplicaSet/efs-csi-controller-5f6d"})
}

func TestGetResourceTree_DeploymentUnfetchedPodUnchanged(t *testing.T) {
	resp := fetchTree(t, deploymentApp(map[string]bool{"ctrl-b": true, "ctrl-c": true}))
	assertStrings(t, "pods", podNames(resp), []string{testNs + "/ctrl-a", testNs + "/ctrl-b", testNs + "/ctrl-c"})
}

// only a real NotFound drops a pod; every other error keeps it
func TestGetResourceTree_NonNotFoundErrorsKeepPod(t *testing.T) {
	cases := map[string]error{
		"InvalidArgument":  status.Error(codes.InvalidArgument, "Pod efs-csi-node-x not found as part of application"),
		"PermissionDenied": status.Error(codes.PermissionDenied, "permission denied"),
		"DeadlineExceeded": status.Error(codes.DeadlineExceeded, "context deadline exceeded"),
		"Unavailable":      status.Error(codes.Unavailable, "connection refused"),
		"Unknown":          status.Error(codes.Unknown, `pods "efs-csi-node-x" not found`),
		"Canceled":         status.Error(codes.Canceled, "context canceled"),
		"Internal":         status.Error(codes.Internal, "internal error"),
		"plain error":      errors.New(`pods "efs-csi-node-x" not found`),
	}
	for name, err := range cases {
		t.Run(name, func(t *testing.T) {
			srv := daemonSetApp(testNs, "efs-csi-node", []string{"efs-csi-node-live", "efs-csi-node-x"}, []string{"efs-csi-node-live"})
			srv.errs[resKey(testNs, "Pod", "efs-csi-node-x")] = err
			resp := fetchTree(t, srv)
			assertStrings(t, "pods", podNames(resp), []string{testNs + "/efs-csi-node-live", testNs + "/efs-csi-node-x"})
			assertStrings(t, "podMetadata", podMetadataNames(resp), []string{"efs-csi-node-live", "efs-csi-node-x"})
		})
	}
}

// NotFound on non-pod resources never removes those nodes
func TestGetResourceTree_NonPodNotFoundKept(t *testing.T) {
	srv := daemonSetApp(testNs, "efs-csi-node", []string{"efs-csi-node-live"}, []string{"efs-csi-node-live"})
	delete(srv.live, resKey(testNs, "ControllerRevision", "efs-csi-node-7d9f8c6b5"))
	delete(srv.live, resKey(testNs, "DaemonSet", "efs-csi-node"))
	resp := fetchTree(t, srv)
	assertStrings(t, "non-pod nodes", nonPodNames(resp), []string{"ControllerRevision/efs-csi-node-7d9f8c6b5", "DaemonSet/efs-csi-node"})
	assertStrings(t, "pods", podNames(resp), []string{testNs + "/efs-csi-node-live"})
}

// a deleted pod in one namespace never hides a same-named pod in another
func TestGetResourceTree_SamePodNameOtherNamespaceKept(t *testing.T) {
	srv := merge(daemonSetApp("ns-a", "agent", []string{"agent-x"}, nil),
		daemonSetApp("ns-b", "agent", []string{"agent-x"}, []string{"agent-x"}))
	resp := fetchTree(t, srv)
	assertStrings(t, "pods", podNames(resp), []string{"ns-b/agent-x"})
}

// nothing deleted: nodes come back exactly as argo sent them
func TestGetResourceTree_NoDeletedPodsTreeUnchanged(t *testing.T) {
	srv := merge(daemonSetApp(testNs, "efs-csi-node", []string{"efs-csi-node-a", "efs-csi-node-b"}, []string{"efs-csi-node-a", "efs-csi-node-b"}),
		deploymentApp(map[string]bool{"ctrl-a": true, "ctrl-b": true, "ctrl-c": true}))
	resp := fetchTree(t, srv)
	if !reflect.DeepEqual(resp.Nodes, srv.nodes) {
		t.Errorf("nodes changed for a tree with no deleted pods")
	}
	assertStrings(t, "podMetadata", podMetadataNames(resp), []string{"ctrl-a", "ctrl-b", "ctrl-c", "efs-csi-node-a", "efs-csi-node-b"})
}

func TestGetResourceTree_EmptyTree(t *testing.T) {
	resp := fetchTree(t, &fakeArgoServer{live: map[string]string{}, errs: map[string]error{}})
	if len(resp.Nodes) != 0 || len(resp.PodMetadata) != 0 {
		t.Errorf("expected empty tree, got %d nodes, %d podMetadata", len(resp.Nodes), len(resp.PodMetadata))
	}
}

func TestGetResourceTree_ResourceTreeErrorReturned(t *testing.T) {
	srv := &fakeArgoServer{treeErr: status.Error(codes.NotFound, "application not found")}
	appName := "efs-csi-dev"
	req := bean.NewImperativeQueryRequest(&application2.ResourcesQuery{ApplicationName: &appName}).WithArgoClusterId(1)
	_, err := newTestService(srv).GetResourceTree(context.Background(), req)
	if status.Code(err) != codes.NotFound {
		t.Errorf("expected argo error to pass through, got %v", err)
	}
}
