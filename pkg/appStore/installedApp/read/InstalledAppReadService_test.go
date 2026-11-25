package read

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type fakeInstalledAppRepository struct {
	repository.InstalledAppRepository
	model     *repository.InstalledApps
	err       error
	gotAppArg string
}

func (f *fakeInstalledAppRepository) GetInstalledAppByGitOpsAppName(acdAppName string) (*repository.InstalledApps, error) {
	f.gotAppArg = acdAppName
	return f.model, f.err
}

func newReadServiceWithLogBuffer(repo repository.InstalledAppRepository) (*InstalledAppReadServiceImpl, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(buf), zapcore.DebugLevel)
	return NewInstalledAppReadServiceImpl(NewInstalledAppReadServiceEAImpl(zap.New(core).Sugar(), repo)), buf
}

func errorLogs(t *testing.T, buf *bytes.Buffer) []map[string]interface{} {
	var out []map[string]interface{}
	sc := bufio.NewScanner(bytes.NewReader(buf.Bytes()))
	for sc.Scan() {
		entry := map[string]interface{}{}
		require.NoError(t, json.Unmarshal(sc.Bytes(), &entry))
		if entry["level"] == "error" {
			out = append(out, entry)
		}
	}
	return out
}

func TestGetInstalledAppByGitOpsAppName(t *testing.T) {
	found := &repository.InstalledApps{Id: 11, AppId: 22, EnvironmentId: 33, Active: true, DeploymentAppType: "argo_cd", GitOpsRepoName: "repo"}
	dbErr := errors.New("dial tcp 127.0.0.1:5432: connect: connection refused")
	tests := []struct {
		name         string
		model        *repository.InstalledApps
		repoErr      error
		wantId       int
		wantErr      error
		wantErrorLog bool
	}{
		{name: "managed app found", model: found, wantId: 11},
		{name: "unmanaged app, no rows", model: &repository.InstalledApps{}, repoErr: pg.ErrNoRows, wantErr: pg.ErrNoRows},
		{name: "unmanaged app, wrapped no rows", model: &repository.InstalledApps{}, repoErr: fmt.Errorf("lookup: %w", pg.ErrNoRows), wantErr: pg.ErrNoRows},
		{name: "db connection failure", model: &repository.InstalledApps{}, repoErr: dbErr, wantErr: dbErr, wantErrorLog: true},
		{name: "multiple rows", model: &repository.InstalledApps{}, repoErr: pg.ErrMultiRows, wantErr: pg.ErrMultiRows, wantErrorLog: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeInstalledAppRepository{model: tt.model, err: tt.repoErr}
			svc, logs := newReadServiceWithLogBuffer(repo)

			got, err := svc.GetInstalledAppByGitOpsAppName("guestbook-dev")

			assert.Equal(t, "guestbook-dev", repo.gotAppArg)
			if tt.wantErr == nil {
				require.NoError(t, err)
				require.NotNil(t, got)
				assert.Equal(t, tt.wantId, got.Id)
				assert.Equal(t, 22, got.AppId)
				assert.Equal(t, 33, got.EnvironmentId)
				assert.Equal(t, "argo_cd", got.DeploymentAppType)
			} else {
				// error is returned unchanged so callers keep their ErrNoRows handling
				assert.Nil(t, got)
				assert.Equal(t, tt.repoErr, err)
				assert.ErrorIs(t, err, tt.wantErr)
			}
			errs := errorLogs(t, logs)
			if tt.wantErrorLog {
				require.Len(t, errs, 1)
				assert.Equal(t, "error while fetching installed app by GitOps app name", errs[0]["msg"])
				assert.Equal(t, "guestbook-dev", errs[0]["acdAppName"])
				assert.Equal(t, tt.repoErr.Error(), errs[0]["error"])
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
