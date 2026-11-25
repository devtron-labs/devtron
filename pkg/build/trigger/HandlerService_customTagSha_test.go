package trigger

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devtron-labs/devtron/client/gitSensor"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	shaRelease12 = "2c8e3eef0d1f4a6b8c9e7f1234567890abcdef12"
	shaRelease13 = "9a41b7c3e5d2f0a1b2c3d4e5f60718293a4b5c6d"
	shaMain      = "5f1c2d3e4a5b6c7d8e9f00112233445566778899"
)

type tagCall struct {
	pipelineMaterialId int
	tag                string
}

type fakeTagResolver struct {
	commits map[string]string
	calls   []tagCall
}

func (f *fakeTagResolver) resolve(pipelineMaterialId int, tag string) string {
	f.calls = append(f.calls, tagCall{pipelineMaterialId, tag})
	return f.commits[tag]
}

func webhookCommit(actionType string, data map[string]string) pipelineConfig.GitCommit {
	return pipelineConfig.GitCommit{WebhookData: pipelineConfig.WebhookData{Id: 7, EventActionType: actionType, Data: data}}
}

func TestGetCommitShaForCustomTag(t *testing.T) {
	materials := []*pipelineConfig.CiPipelineMaterial{nil, {Id: 11, GitMaterialId: 1}, {Id: 12, GitMaterialId: 2}}
	tagEvent := func(tag string) pipelineConfig.GitCommit {
		return webhookCommit("non-merged", map[string]string{bean.WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME: tag, "tag name": tag})
	}
	tests := []struct {
		name                    string
		commitHashes            map[int]pipelineConfig.GitCommit
		dockerfileGitMaterialId int
		want                    string
		wantCalls               []tagCall
	}{
		{
			name:                    "manual or branch build uses the commit",
			commitHashes:            map[int]pipelineConfig.GitCommit{11: {Commit: shaMain}},
			dockerfileGitMaterialId: 1,
			want:                    "5f1c2d3e",
		},
		{
			name:                    "tag webhook resolves the tag to its commit",
			commitHashes:            map[int]pipelineConfig.GitCommit{11: tagEvent("release/1.2")},
			dockerfileGitMaterialId: 1,
			want:                    "2c8e3eef",
			wantCalls:               []tagCall{{11, "release/1.2"}},
		},
		{
			name:                    "tags sharing an 8 char prefix get their own commits",
			commitHashes:            map[int]pipelineConfig.GitCommit{11: tagEvent("release/1.3")},
			dockerfileGitMaterialId: 1,
			want:                    "9a41b7c3",
			wantCalls:               []tagCall{{11, "release/1.3"}},
		},
		{
			name:                    "unresolvable tag gives empty sha so the missing sha error fires",
			commitHashes:            map[int]pipelineConfig.GitCommit{11: tagEvent("does-not-exist")},
			dockerfileGitMaterialId: 1,
			want:                    "",
			wantCalls:               []tagCall{{11, "does-not-exist"}},
		},
		{
			name: "PR merged webhook keeps the target checkout commit",
			commitHashes: map[int]pipelineConfig.GitCommit{11: webhookCommit(bean.WEBHOOK_EVENT_MERGED_ACTION_TYPE,
				map[string]string{bean.WEBHOOK_SELECTOR_TARGET_CHECKOUT_NAME: shaMain, bean.WEBHOOK_SELECTOR_SOURCE_CHECKOUT_NAME: shaRelease13})},
			dockerfileGitMaterialId: 1,
			want:                    "5f1c2d3e",
		},
		{
			name:                    "webhook without target checkout gives empty sha",
			commitHashes:            map[int]pipelineConfig.GitCommit{11: webhookCommit("non-merged", map[string]string{"source checkout": shaMain})},
			dockerfileGitMaterialId: 1,
			want:                    "",
		},
		{
			name:                    "uses the material holding the Dockerfile",
			commitHashes:            map[int]pipelineConfig.GitCommit{11: {Commit: shaMain}, 12: tagEvent("release/1.2")},
			dockerfileGitMaterialId: 2,
			want:                    "2c8e3eef",
			wantCalls:               []tagCall{{12, "release/1.2"}},
		},
		{
			name:                    "no material for the Dockerfile gives empty sha",
			commitHashes:            map[int]pipelineConfig.GitCommit{11: tagEvent("release/1.2")},
			dockerfileGitMaterialId: 3,
			want:                    "",
		},
		{
			name:                    "material without commit data gives empty sha",
			commitHashes:            map[int]pipelineConfig.GitCommit{},
			dockerfileGitMaterialId: 1,
			want:                    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &fakeTagResolver{commits: map[string]string{"release/1.2": shaRelease12, "release/1.3": shaRelease13}}
			got := getCommitShaForCustomTag(materials, tt.commitHashes, tt.dockerfileGitMaterialId, resolver.resolve)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantCalls, resolver.calls)
		})
	}
}

type fakeGitSensorClient struct {
	gitSensor.Client
	commit *gitSensor.GitCommit
	err    error
	req    *gitSensor.CommitMetadataRequest
}

func (f *fakeGitSensorClient) GetCommitMetadata(_ context.Context, req *gitSensor.CommitMetadataRequest) (*gitSensor.GitCommit, error) {
	f.req = req
	return f.commit, f.err
}

func TestGetCommitForWebhookTag(t *testing.T) {
	tests := []struct {
		name   string
		commit *gitSensor.GitCommit
		err    error
		want   string
	}{
		{name: "tag resolved", commit: &gitSensor.GitCommit{Commit: shaRelease12}, want: shaRelease12},
		{name: "git-sensor error", err: errors.New("reference not found"), want: ""},
		{name: "no commit returned", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &fakeGitSensorClient{commit: tt.commit, err: tt.err}
			impl := &HandlerServiceImpl{gitSensorClient: client, Logger: zap.NewNop().Sugar()}
			assert.Equal(t, tt.want, impl.getCommitForWebhookTag(11, "release/1.2"))
			assert.Equal(t, &gitSensor.CommitMetadataRequest{PipelineMaterialId: 11, GitTag: "release/1.2"}, client.req)
		})
	}
}

// git-sensor REST client against a fake commit-metadata endpoint, so the wire format is covered too
func TestGetCommitForWebhookTagWithRestClient(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		want     string
	}{
		{name: "tag resolved", status: http.StatusOK, response: `{"code":200,"status":"OK","result":{"Commit":"` + shaRelease12 + `"}}`, want: shaRelease12},
		{name: "tag not found", status: http.StatusBadRequest, response: `{"code":400,"status":"Bad Request","errors":[{"code":"000","internalMessage":"object not found","userMessage":"object not found"}]}`, want: ""},
		{name: "null result", status: http.StatusOK, response: `{"code":200,"status":"OK","result":null}`, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			var gotBody map[string]interface{}
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.Method + " " + r.URL.Path
				_ = json.NewDecoder(r.Body).Decode(&gotBody)
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.response))
			}))
			defer server.Close()
			client, err := gitSensor.NewGitSensorClient(zap.NewNop().Sugar(), &gitSensor.ClientConfig{Url: server.URL, Protocol: "REST"})
			require.NoError(t, err)
			impl := &HandlerServiceImpl{gitSensorClient: client, Logger: zap.NewNop().Sugar()}

			assert.Equal(t, tt.want, impl.getCommitForWebhookTag(11, "release/1.2"))
			assert.Equal(t, "POST /commit-metadata", gotPath)
			assert.Equal(t, map[string]interface{}{"pipelineMaterialId": float64(11), "gitTag": "release/1.2", "gitHash": "", "branchName": ""}, gotBody)
		})
	}
}

func TestHasShaVariable(t *testing.T) {
	assert.True(t, hasShaVariable("dev-{sha}-{x}"))
	assert.True(t, hasShaVariable("dev-{SHA}-{x}"))
	assert.False(t, hasShaVariable("v1.{x}"))
	assert.False(t, hasShaVariable("sha-{x}"))
}
