package app

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	"github.com/devtron-labs/devtron/internal/util"
	mock_user "github.com/devtron-labs/devtron/pkg/auth/user/mocks"
	"github.com/devtron-labs/devtron/pkg/auth/user/mocks/casbin"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/pipeline/mock_pipeline"
	mocks_rbac "github.com/devtron-labs/devtron/util/mocks/rbac"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/go-playground/validator.v9"
)

func TestPipelineConfigRestHandlerImpl_PatchCiMaterialSource(t *testing.T) {
	logger, err := util.NewSugardLogger()
	if err != nil {
		panic(err)
	}
	type fields struct {
		userAuthService *mock_user.MockUserService
		pipelineBuilder *mock_pipeline.MockPipelineBuilder
		validator       *validator.Validate
		enforcer        *mock_casbin.MockEnforcer
		enforcerUtil    *mocks_rbac.MockEnforcerUtil
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		body               string
		setup              func(fields2 *fields)
		expectedStatusCode int
	}{
		{
			name: "when user is not found, it should return unauthorized",
			fields: fields{
				validator: validator.New(),
			},
			body: "{\"appId\":2, \"environmentId\": 1 ,\"source\":{\"type\":\"SOURCE_TYPE_BRANCH_FIXED\",\"value\":\"main10\",\"regex\":\"\"}}",
			setup: func(fields2 *fields) {
				ctrl := gomock.NewController(t)
				fields2.pipelineBuilder = mock_pipeline.NewMockPipelineBuilder(ctrl)
				fields2.enforcer = mock_casbin.NewMockEnforcer(ctrl)
				fields2.enforcerUtil = mocks_rbac.NewMockEnforcerUtil(ctrl)
				fields2.userAuthService = mock_user.NewMockUserService(ctrl)
				fields2.userAuthService.EXPECT().GetLoggedInUser(gomock.Any()).Return(int32(0), fmt.Errorf("user not found")).Times(1)
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "when request is malformed, it should return bad request",
			fields: fields{
				validator: validator.New(),
			},
			body: "{\"appId\":4, \"id\": 5 ,-\"ciMaterial\":[{\"gitMaterialId\":4,\"id\":5,\"source\":{\"type\":\"SOURCE_TYPE_BRANCH_FIXED\",\"value\":\"main3\",\"regex\":\"\"}}]}",
			setup: func(fields2 *fields) {
				ctrl := gomock.NewController(t)
				fields2.pipelineBuilder = mock_pipeline.NewMockPipelineBuilder(ctrl)
				fields2.enforcer = mock_casbin.NewMockEnforcer(ctrl)
				fields2.enforcerUtil = mocks_rbac.NewMockEnforcerUtil(ctrl)
				fields2.userAuthService = mock_user.NewMockUserService(ctrl)
				fields2.userAuthService.EXPECT().GetLoggedInUser(gomock.Any()).Return(int32(1), nil).Times(1)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "when app is not found for the given appId, it should return bad request",
			fields: fields{
				validator: validator.New(),
			},
			body: "{\"appId\":4, \"id\": 5 ,\"source\":{\"type\":\"SOURCE_TYPE_BRANCH_FIXED\",\"value\":\"main3\",\"regex\":\"\"}}",
			setup: func(fields2 *fields) {
				ctrl := gomock.NewController(t)
				fields2.pipelineBuilder = mock_pipeline.NewMockPipelineBuilder(ctrl)
				fields2.pipelineBuilder.EXPECT().GetApp(4).Return(nil, fmt.Errorf("app not found")).Times(1)
				fields2.enforcer = mock_casbin.NewMockEnforcer(ctrl)
				fields2.enforcerUtil = mocks_rbac.NewMockEnforcerUtil(ctrl)
				fields2.userAuthService = mock_user.NewMockUserService(ctrl)
				fields2.userAuthService.EXPECT().GetLoggedInUser(gomock.Any()).Return(int32(1), nil).Times(1)
				//fields2.userAuthService.EXPECT().IsSuperAdmin(1).Return(true, nil).Times(1)
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "when validator fails, it should return Bad request",
			fields: fields{
				validator: validator.New(),
			},
			body: "{\"appId\":4, \"id\": 5 ,\"ciMaterial\":[{\"gitMaterialId\":4,\"id\":5,\"source\":{\"type\":\"SOURCE_TYPE_BRANCH_FIXED\",\"value\":\"main3\",\"regex\":\"\"}}]}",
			setup: func(fields2 *fields) {
				ctrl := gomock.NewController(t)
				_ = fields2.validator.RegisterValidation("name-component", func(fl validator.FieldLevel) bool {
					return false
				})
				fields2.pipelineBuilder = mock_pipeline.NewMockPipelineBuilder(ctrl)
				fields2.pipelineBuilder.EXPECT().GetApp(4).Return(&bean.CreateAppDTO{AppName: "Super App", Id: 4, AppType: helper.Job}, nil).Times(1)
				fields2.enforcer = mock_casbin.NewMockEnforcer(ctrl)
				fields2.enforcerUtil = mocks_rbac.NewMockEnforcerUtil(ctrl)
				//fields2.enforcerUtil.EXPECT().GetAppRBACName("Super App")
				fields2.userAuthService = mock_user.NewMockUserService(ctrl)
				fields2.userAuthService.EXPECT().GetLoggedInUser(gomock.Any()).Return(int32(1), nil).Times(1)
				//fields2.userAuthService.EXPECT().IsSuperAdmin(1).Return(true, nil).Times(1)

			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "when app is not jobtype and enforce fails, it should return forbidden",
			fields: fields{
				validator: validator.New(),
			},
			body: "{\"appId\":4, \"id\": 5 ,\"ciMaterial\":[{\"gitMaterialId\":4,\"id\":5,\"source\":{\"type\":\"SOURCE_TYPE_BRANCH_FIXED\",\"value\":\"main3\",\"regex\":\"\"}}]}",
			setup: func(fields2 *fields) {
				ctrl := gomock.NewController(t)
				fields2.pipelineBuilder = mock_pipeline.NewMockPipelineBuilder(ctrl)
				fields2.pipelineBuilder.EXPECT().GetApp(4).Return(&bean.CreateAppDTO{AppName: "Super App", Id: 4, AppType: helper.CustomApp}, nil).Times(1)
				fields2.enforcer = mock_casbin.NewMockEnforcer(ctrl)
				fields2.enforcer.EXPECT().Enforce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(false).Times(1)

				fields2.enforcerUtil = mocks_rbac.NewMockEnforcerUtil(ctrl)
				fields2.enforcerUtil.EXPECT().GetAppRBACName("Super App")
				fields2.userAuthService = mock_user.NewMockUserService(ctrl)
				fields2.userAuthService.EXPECT().GetLoggedInUser(gomock.Any()).Return(int32(1), nil).Times(1)
				//fields2.userAuthService.EXPECT().IsSuperAdmin(1).Return(false, nil).Times(1)

			},
			expectedStatusCode: http.StatusForbidden,
		},
		{
			name: "when PatchCiMaterialSource call fails, it should return internal server error",
			fields: fields{
				validator: validator.New(),
			},
			body: "{\"appId\":4, \"environmentId\": 1 ,\"source\":{\"type\":\"SOURCE_TYPE_BRANCH_FIXED\",\"value\":\"main10\",\"regex\":\"\"}}",
			setup: func(fields2 *fields) {
				_ = fields2.validator.RegisterValidation("name-component", func(fl validator.FieldLevel) bool {
					return true
				})
				ctrl := gomock.NewController(t)
				fields2.pipelineBuilder = mock_pipeline.NewMockPipelineBuilder(ctrl)
				fields2.pipelineBuilder.EXPECT().GetApp(4).Return(&bean.CreateAppDTO{AppName: "Super App", Id: 4, AppType: helper.CustomApp}, nil).Times(1)
				fields2.pipelineBuilder.EXPECT().PatchCiMaterialSource(gomock.Any(), int32(1)).Return(nil, fmt.Errorf("failed to patch"))
				fields2.enforcer = mock_casbin.NewMockEnforcer(ctrl)
				fields2.enforcer.EXPECT().Enforce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).Times(1)
				fields2.enforcerUtil = mocks_rbac.NewMockEnforcerUtil(ctrl)
				fields2.enforcerUtil.EXPECT().GetAppRBACName("Super App")
				fields2.userAuthService = mock_user.NewMockUserService(ctrl)
				fields2.userAuthService.EXPECT().GetLoggedInUser(gomock.Any()).Return(int32(1), nil).Times(1)
				//fields2.userAuthService.EXPECT().IsSuperAdmin(1).Return(false, nil).Times(1)

			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "when PatchCiMaterialSource call passes, it should return statusok",
			fields: fields{
				validator: validator.New(),
			},
			body: "{\"appId\":4, \"environmentId\": 1 ,\"source\":{\"type\":\"SOURCE_TYPE_BRANCH_FIXED\",\"value\":\"main10\",\"regex\":\"\"}}",
			setup: func(fields2 *fields) {
				_ = fields2.validator.RegisterValidation("name-component", func(fl validator.FieldLevel) bool {
					return true
				})
				ctrl := gomock.NewController(t)
				fields2.pipelineBuilder = mock_pipeline.NewMockPipelineBuilder(ctrl)
				fields2.pipelineBuilder.EXPECT().GetApp(4).Return(&bean.CreateAppDTO{AppName: "Super App", Id: 4, AppType: helper.CustomApp}, nil).Times(1)
				fields2.pipelineBuilder.EXPECT().PatchCiMaterialSource(gomock.Any(), int32(1)).Return(&bean.CiMaterialPatchRequest{AppId: 2}, nil)
				fields2.enforcer = mock_casbin.NewMockEnforcer(ctrl)
				fields2.enforcer.EXPECT().Enforce(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(true).Times(1)
				fields2.enforcerUtil = mocks_rbac.NewMockEnforcerUtil(ctrl)
				fields2.enforcerUtil.EXPECT().GetAppRBACName("Super App")
				fields2.userAuthService = mock_user.NewMockUserService(ctrl)
				fields2.userAuthService.EXPECT().GetLoggedInUser(gomock.Any()).Return(int32(1), nil).Times(1)
				//fields2.userAuthService.EXPECT().IsSuperAdmin(1).Return(false, nil).Times(1)

			},
			expectedStatusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(&tt.fields)
			handler := PipelineConfigRestHandlerImpl{
				userAuthService: tt.fields.userAuthService,
				pipelineBuilder: tt.fields.pipelineBuilder,
				validator:       tt.fields.validator,
				enforcer:        tt.fields.enforcer,
				enforcerUtil:    tt.fields.enforcerUtil,
				Logger:          logger,
			}

			req, err := http.NewRequest("PATCH", "/orchestrator/app/ci-pipeline/patch-branch", bytes.NewBuffer([]byte(tt.body)))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h := http.HandlerFunc(handler.PatchCiMaterialSourceWithAppIdAndEnvironmentId)
			h.ServeHTTP(rr, req)
			assert.Equal(t, rr.Code, tt.expectedStatusCode)
		})
	}
}
