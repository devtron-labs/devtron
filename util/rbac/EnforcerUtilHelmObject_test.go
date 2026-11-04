/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package rbac

import (
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/app"
	environmentRepository "github.com/devtron-labs/devtron/pkg/cluster/environment/repository"
	clusterRepository "github.com/devtron-labs/devtron/pkg/cluster/repository"
	teamRepo "github.com/devtron-labs/devtron/pkg/team/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
)

// stubAppRepo overrides only the lookups used while building helm rbac objects; every other method of the
// interface is left nil and will panic if a code path unexpectedly reaches it.
type stubAppRepo struct {
	app.AppRepository
	byName map[string]*app.App
	byId   map[int]*app.App
}

func (s *stubAppRepo) FindAppAndProjectByAppName(appName string) (*app.App, error) {
	if a, ok := s.byName[appName]; ok {
		return a, nil
	}
	// mirrors the real repository, which returns an empty (non nil) app alongside ErrNoRows
	return &app.App{}, pg.ErrNoRows
}

func (s *stubAppRepo) FindAppAndProjectByAppId(appId int) (*app.App, error) {
	if a, ok := s.byId[appId]; ok {
		return a, nil
	}
	return &app.App{}, pg.ErrNoRows
}

type stubEnvRepo struct {
	environmentRepository.EnvironmentRepository
	env *environmentRepository.Environment
}

func (s *stubEnvRepo) FindById(id int) (*environmentRepository.Environment, error) {
	if s.env == nil || s.env.Id != id {
		return nil, pg.ErrNoRows
	}
	return s.env, nil
}

const (
	testClusterName = "default_cluster"
	testNamespace   = "devtron-demo"
	testEnvId       = 1
	testClusterId   = 1
)

// linkedExternalApp is how an external helm app looks in the app table once it is linked to chart store:
// app_name carries the unique identifier while display_name carries the release name.
func linkedExternalApp(id int, releaseName string, teamId int, teamName string) *app.App {
	return &app.App{
		Id:          id,
		AppName:     releaseName + "-" + testNamespace + "-1",
		DisplayName: releaseName,
		TeamId:      teamId,
		Team:        teamRepo.Team{Name: teamName},
	}
}

func newTestEnforcerUtil(envIdentifier string) *EnforcerUtilImpl {
	nginx := linkedExternalApp(1, "nginx", 2, "myproject")
	// linked with no project assigned -> team_id stays 0 and the joined team row is empty
	redis := linkedExternalApp(2, "redis", 0, "")
	// a regular chart store app, which has no display name at all
	grafana := &app.App{Id: 3, AppName: "grafana", TeamId: 2, Team: teamRepo.Team{Name: "myproject"}}

	appRepo := &stubAppRepo{
		byName: map[string]*app.App{
			nginx.AppName:   nginx,
			redis.AppName:   redis,
			grafana.AppName: grafana,
		},
		byId: map[int]*app.App{1: nginx, 2: redis, 3: grafana},
	}
	envRepo := &stubEnvRepo{env: &environmentRepository.Environment{
		Id:                    testEnvId,
		Name:                  testNamespace,
		ClusterId:             testClusterId,
		Namespace:             testNamespace,
		EnvironmentIdentifier: envIdentifier,
		Cluster:               &clusterRepository.Cluster{Id: testClusterId, ClusterName: testClusterName},
	}}
	return &EnforcerUtilImpl{logger: zap.NewNop().Sugar(), appRepo: appRepo, environmentRepository: envRepo}
}

func TestGetHelmObjectByAppNameAndEnvId(t *testing.T) {
	envIdentifier := testClusterName + "__" + testNamespace
	tests := []struct {
		name        string
		appName     string
		wantObject  string
		wantObject2 string
	}{
		{
			name: "external app linked to chart store is resolved by its display name",
			// this is the regression: the listing APIs hand over the release name, but app_name in db holds
			// the unique identifier, so the lookup used to miss and the object came back as "//"
			appName:    "nginx",
			wantObject: "myproject/" + envIdentifier + "/nginx",
		},
		{
			name:       "external app linked without a project falls under the unassigned project",
			appName:    "redis",
			wantObject: teamRepo.UNASSIGNED_PROJECT + "/" + envIdentifier + "/redis",
		},
		{
			name:       "regular chart store app keeps working via the plain app name",
			appName:    "grafana",
			wantObject: "myproject/" + envIdentifier + "/grafana",
		},
		{
			name: "caller passing the unique identifier still gets the display name based object",
			// some handlers hand over installedApp.App.AppName directly
			appName:    "nginx-" + testNamespace + "-1",
			wantObject: "myproject/" + envIdentifier + "/nginx",
		},
		{
			name:       "unknown app yields an empty object",
			appName:    "does-not-exist",
			wantObject: "//",
		},
	}
	impl := newTestEnforcerUtil(envIdentifier)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			object, object2 := impl.GetHelmObjectByAppNameAndEnvId(tt.appName, testEnvId)
			if object != tt.wantObject {
				t.Errorf("object = %q, want %q", object, tt.wantObject)
			}
			if object2 != tt.wantObject2 {
				t.Errorf("object2 = %q, want %q", object2, tt.wantObject2)
			}
		})
	}
}

// When the environment identifier is not clusterName__namespace, both the migrated and the futuristic
// permission objects are returned, and each has to carry the display name.
func TestGetHelmObjectByAppNameAndEnvId_FuturisticPermissionObject(t *testing.T) {
	impl := newTestEnforcerUtil("devtron-demo")
	object, object2 := impl.GetHelmObjectByAppNameAndEnvId("nginx", testEnvId)
	if want := "myproject/devtron-demo/nginx"; object != want {
		t.Errorf("object = %q, want %q", object, want)
	}
	if want := "myproject/" + testClusterName + "__" + testNamespace + "/nginx"; object2 != want {
		t.Errorf("object2 = %q, want %q", object2, want)
	}
}

// GetHelmObjectByAppNameAndEnvId lowercases its objects, so a mixed case release name still has to line up
// with the policy that was stored for it.
func TestGetHelmObjectByAppNameAndEnvId_LowerCased(t *testing.T) {
	envIdentifier := testClusterName + "__" + testNamespace
	mixed := &app.App{
		Id:          4,
		AppName:     "MyRelease-" + testNamespace + "-1",
		DisplayName: "MyRelease",
		TeamId:      2,
		Team:        teamRepo.Team{Name: "MyProject"},
	}
	impl := newTestEnforcerUtil(envIdentifier)
	impl.appRepo.(*stubAppRepo).byName[mixed.AppName] = mixed

	object, _ := impl.GetHelmObjectByAppNameAndEnvId("MyRelease", testEnvId)
	if want := "myproject/" + envIdentifier + "/myrelease"; object != want {
		t.Errorf("object = %q, want %q", object, want)
	}
}

func TestGetHelmObject(t *testing.T) {
	envIdentifier := testClusterName + "__" + testNamespace
	impl := newTestEnforcerUtil(envIdentifier)

	// app id lookup always succeeded, but the object used to be built from app_name (the unique identifier)
	object, _ := impl.GetHelmObject(1, testEnvId)
	if want := "myproject/" + envIdentifier + "/nginx"; object != want {
		t.Errorf("object = %q, want %q", object, want)
	}

	object, _ = impl.GetHelmObject(2, testEnvId)
	if want := teamRepo.UNASSIGNED_PROJECT + "/" + envIdentifier + "/redis"; object != want {
		t.Errorf("object = %q, want %q", object, want)
	}

	object, _ = impl.GetHelmObject(3, testEnvId)
	if want := "myproject/" + envIdentifier + "/grafana"; object != want {
		t.Errorf("object = %q, want %q", object, want)
	}
}

func TestGetAppNameForRbac(t *testing.T) {
	external := &app.App{AppName: "nginx-devtron-demo-1", DisplayName: "nginx"}
	if got := external.GetAppNameForRbac(); got != "nginx" {
		t.Errorf("GetAppNameForRbac() = %q, want %q", got, "nginx")
	}
	regular := &app.App{AppName: "grafana"}
	if got := regular.GetAppNameForRbac(); got != "grafana" {
		t.Errorf("GetAppNameForRbac() = %q, want %q", got, "grafana")
	}
}
