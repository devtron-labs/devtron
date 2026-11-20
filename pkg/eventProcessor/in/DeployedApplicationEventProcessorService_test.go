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

package in

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/argoproj/argo-cd/v3/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/common-lib/pubsub-lib/model"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	"github.com/devtron-labs/devtron/pkg/app"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/read"
	"github.com/devtron-labs/devtron/pkg/appStore/installedApp/repository"
	"github.com/go-pg/pg"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestHandleArgoAppUpdateWithoutPipeline(t *testing.T) {
	databaseErr := errors.New("database connection failed")
	tests := []struct {
		name         string
		installedApp *repository.InstalledApps
		lookupErr    error
		wantError    bool
	}{
		{name: "unregistered application", lookupErr: pg.ErrNoRows},
		{name: "unregistered application with wrapped error", lookupErr: fmt.Errorf("lookup: %w", pg.ErrNoRows)},
		{name: "database failure", lookupErr: databaseErr, wantError: true},
		{name: "registered chart", installedApp: &repository.InstalledApps{Id: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			encoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
			logger := zap.New(zapcore.NewCore(encoder, zapcore.AddSync(&logs), zap.ErrorLevel)).Sugar()
			appService := &argoAppStatusRecorder{}
			processor := &DeployedApplicationEventProcessorImpl{
				logger:             logger,
				appService:         appService,
				pipelineRepository: &noArgoPipelineRepository{},
				installedAppReadService: read.NewInstalledAppReadServiceImpl(read.NewInstalledAppReadServiceEAImpl(
					logger, &argoInstalledAppLookup{installedApp: tt.installedApp, err: tt.lookupErr},
				)),
			}

			processor.handleArgoAppUpdate(&model.PubSubMsg{
				Data: `{"application":{"metadata":{"name":"demo"}}}`,
			})

			if tt.wantError {
				if !strings.Contains(logs.String(), `"error":"`+databaseErr.Error()+`"`) {
					t.Errorf("expected the lookup service to log the database failure: %v", logs.String())
				}
				if !strings.Contains(logs.String(), `"err":"`+databaseErr.Error()+`"`) {
					t.Errorf("expected the event processor to log the database failure: %v", logs.String())
				}
			} else if logs.Len() != 0 {
				t.Errorf("unexpected error logs: %v", logs.String())
			}
			if want := tt.installedApp != nil; appService.updated != want || appService.updatedChart != want {
				t.Errorf("updated = %t, updated chart = %t, want both %t", appService.updated, appService.updatedChart, want)
			}
		})
	}
}

type noArgoPipelineRepository struct {
	pipelineConfig.PipelineRepository
}

func (*noArgoPipelineRepository) GetArgoPipelineByArgoAppName(string) ([]pipelineConfig.Pipeline, error) {
	return nil, nil
}

type argoInstalledAppLookup struct {
	repository.InstalledAppRepository
	installedApp *repository.InstalledApps
	err          error
}

func (r *argoInstalledAppLookup) GetInstalledAppByGitOpsAppName(string) (*repository.InstalledApps, error) {
	return r.installedApp, r.err
}

type argoAppStatusRecorder struct {
	app.AppService
	updated      bool
	updatedChart bool
}

func (s *argoAppStatusRecorder) UpdateDeploymentStatusForGitOpsPipelines(_ *v1alpha1.Application, _ int, _ time.Time, isAppStore bool) (bool, bool, *chartConfig.PipelineOverride, error) {
	s.updated = true
	s.updatedChart = isAppStore
	return false, false, nil, nil
}
