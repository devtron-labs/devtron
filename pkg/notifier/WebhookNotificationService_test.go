package notifier

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/sql/repository"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	util2 "github.com/devtron-labs/devtron/internal/util"
	"github.com/devtron-labs/devtron/pkg/team/mocks"
	"github.com/stretchr/testify/mock"

	//"github.com/devtron-labs/devtron/pkg/user/repository"
	mocks3 "github.com/devtron-labs/devtron/pkg/user/repository/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_buildWebhookNewConfigs(t *testing.T) {
	type args struct {
		webhookReq []WebhookConfigDto
		userId     int32
	}
	tests := []struct {
		name string
		args args
		want []*repository.WebhookConfig
	}{
		{
			name: "test1",
			args: args{
				webhookReq: []WebhookConfigDto{
					{
						WebhookUrl: "dfcd nmc dc",
						ConfigName: "aditya",
						Payload:    map[string]interface{}{"text": "final"},
						Header:     map[string]interface{}{"Content-type": "application/json"},
					},
				},
				userId: 1,
			},
			want: []*repository.WebhookConfig{
				{
					WebHookUrl: "dfcd nmc dc",
					ConfigName: "aditya",
					Payload:    map[string]interface{}{"text": "final"},
					Header:     map[string]interface{}{"Content-type": "application/json"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildWebhookNewConfigs(tt.args.webhookReq, tt.args.userId)

			assert.Equal(t, len(tt.want), len(got), "Number of webhook configs mismatch")

			for i, want := range tt.want {
				assert.Equal(t, want.WebHookUrl, got[i].WebHookUrl, "WebHookUrl mismatch")
				assert.Equal(t, want.ConfigName, got[i].ConfigName, "ConfigName mismatch")
				assert.Equal(t, want.Payload, got[i].Payload, "Payload mismatch")
				assert.Equal(t, want.Header, got[i].Header, "Header mismatch")

			}
		})
	}
}

func TestWebhookNotificationServiceImpl_SaveOrEditNotificationConfig(t *testing.T) {
	sugaredLogger, err := util2.NewSugardLogger()
	assert.Nil(t, err)
	mockedTeamService := mocks.NewTeamService(t)
	mockedWebhookNotfRep := mocks2.NewWebhookNotificationRepository(t)
	mockedUserRepo := mocks3.NewUserRepository(t)
	mockedNotfSetRepo := mocks2.NewNotificationSettingsRepository(t)

	type args struct {
		channelReq []WebhookConfigDto
		userId     int32
	}

	tests := []struct {
		name    string
		args    args
		want    []int
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "SaveOrUpdate_ExistingConfig",
			args: args{
				channelReq: []WebhookConfigDto{
					{
						WebhookUrl: "djfndgfbd,gds",
						ConfigName: "aditya",
						Payload:    map[string]interface{}{"text": "final"},
						Header:     map[string]interface{}{"Content-type": "application/json"},
					},
				},
				userId: 2,
			},
			want:    []int{0},
			wantErr: assert.NoError,
		},
		{
			name: "SaveOrUpdate_NewConfig",
			args: args{
				channelReq: []WebhookConfigDto{
					{
						WebhookUrl: "d,fm sdfd",
						ConfigName: "aditya",
						Payload:    map[string]interface{}{"text": "final"},
						Header:     map[string]interface{}{"Content-type": "application/json"},
					},
				},
				userId: 2,
			},
			want:    []int{0},
			wantErr: assert.NoError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl := &WebhookNotificationServiceImpl{
				logger:                         sugaredLogger,
				webhookRepository:              mockedWebhookNotfRep,
				teamService:                    mockedTeamService,
				userRepository:                 mockedUserRepo,
				notificationSettingsRepository: mockedNotfSetRepo,
			}

			mockConfig := &repository.WebhookConfig{Id: 1}
			mockError := error(nil)
			mockedWebhookNotfRep.On("SaveWebhookConfig", mock.Anything).Return(mockConfig, mockError)

			got, err := impl.SaveOrEditNotificationConfig(tt.args.channelReq, tt.args.userId)

			if !tt.wantErr(t, err, fmt.Sprintf("SaveOrEditNotificationConfig(%v, %v)", tt.args.channelReq, tt.args.userId)) {
				return
			}
			assert.Equalf(t, tt.want, got, "SaveOrEditNotificationConfig(%v, %v)", tt.args.channelReq, tt.args.userId)
		})
	}
}
