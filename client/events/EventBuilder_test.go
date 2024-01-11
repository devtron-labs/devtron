package client

import (
	"errors"
	"reflect"
	"testing"

	repository2 "github.com/devtron-labs/devtron/internal/sql/repository"
	"github.com/devtron-labs/devtron/internal/sql/repository/chartConfig/mocks"
	repository "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging"
	mocks5 "github.com/devtron-labs/devtron/internal/sql/repository/imageTagging/mocks"
	mocks4 "github.com/devtron-labs/devtron/internal/sql/repository/mocks"
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	mocks2 "github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig/mocks"
	util2 "github.com/devtron-labs/devtron/internal/util"
	repository3 "github.com/devtron-labs/devtron/pkg/auth/user/repository"
	mocks3 "github.com/devtron-labs/devtron/pkg/auth/user/repository/mocks"
	"github.com/devtron-labs/devtron/pkg/bean"
	"github.com/devtron-labs/devtron/pkg/notifier"
)

func TestEventSimpleFactoryImpl_BuildExtraApprovalData(t *testing.T) {
	t.SkipNow()
	defaultSesConfig := &repository2.SESConfig{
		Id:         1,
		Region:     "us-east-1",
		AccessKey:  "12e2wdwewd",
		SecretKey:  "ddsewsafcds",
		FromEmail:  "abc@gmail.com",
		ConfigName: "ses-config",
		Deleted:    false,
		Default:    true,
	}

	event := Event{
		EventTypeId:   4,
		PipelineId:    1,
		AppId:         1,
		EnvId:         2,
		CorrelationId: "dfsfdsf",
	}
	approvalActionRequest := bean.UserApprovalActionRequest{
		AppId:      1,
		ActionType: 0,
		PipelineId: 1,
		ArtifactId: 23,
		ApprovalNotificationConfig: bean.ApprovalNotificationConfig{
			EmailIds: []string{"abc@gmail.com", "abcd123@gmail.com"},
		},
		ApprovalRequestId: 0,
	}
	cdPipeline := &pipelineConfig.Pipeline{
		Id:    1,
		AppId: 1,
	}
	imageComment := repository.ImageComment{
		Id:         1,
		Comment:    "no comment",
		ArtifactId: approvalActionRequest.ArtifactId,
		UserId:     1,
	}
	imageTags := []*repository.ImageTag{
		{
			Id:         1,
			TagName:    "no tag",
			AppId:      1,
			ArtifactId: approvalActionRequest.ArtifactId,
		},
	}
	ciArtifact := repository2.CiArtifact{
		Id:         1,
		PipelineId: 1,
		Image:      "ashexp:435345ds",
	}
	user := repository3.UserModel{
		Id:      1,
		EmailId: "abc@gmail.com",
	}
	eventResponse5 := Event{
		EventTypeId:   4,
		PipelineId:    1,
		CorrelationId: "dfsfdsf",
		Payload: &Payload{
			DockerImageUrl: "ashexp:435345ds",
			Providers: []*notifier.Provider{
				{
					Destination: "ses",
					ConfigId:    0,
					Recipient:   "abc@gmail.com",
				},
				{
					Destination: "ses",
					ConfigId:    0,
					Recipient:   "abcd123@gmail.com",
				},
			},
			ImageTagNames: []string{
				"no tag",
			},
			ImageComment:      "no comment",
			ImageApprovalLink: "/dashboard/app/1/trigger?approval-node&imageTag=435345ds",
		},
		AppId: 1,
		EnvId: 2,
	}
	eventResponse := Event{
		EventTypeId:   4,
		PipelineId:    1,
		CorrelationId: "dfsfdsf",
		Payload: &Payload{
			DockerImageUrl: "ashexp:435345ds",
			TriggeredBy:    "abc@gmail.com",
			Providers: []*notifier.Provider{
				{
					Destination: "ses",
					ConfigId:    0,
					Recipient:   "abc@gmail.com",
				},
				{
					Destination: "ses",
					ConfigId:    0,
					Recipient:   "abcd123@gmail.com",
				},
			},
			ImageTagNames: []string{
				"no tag",
			},
			ImageComment:      "no comment",
			ImageApprovalLink: "/dashboard/app/1/trigger?approval-node&imageTag=435345ds",
		},
		AppId: 1,
		EnvId: 2,
	}
	eventResponse1 := Event{
		EventTypeId:   4,
		PipelineId:    1,
		CorrelationId: "dfsfdsf",
		AppId:         1,
		EnvId:         2,
	}
	eventResponse2 := Event{
		EventTypeId:   4,
		PipelineId:    1,
		Payload:       &Payload{},
		CorrelationId: "dfsfdsf",
		AppId:         1,
		EnvId:         2,
	}
	eventResponse3 := Event{
		EventTypeId: 4,
		PipelineId:  1,
		Payload: &Payload{
			ImageComment: "no comment",
		},
		CorrelationId: "dfsfdsf",
		AppId:         1,
		EnvId:         2,
	}
	eventResponse4 := Event{
		EventTypeId: 4,
		PipelineId:  1,
		Payload: &Payload{
			ImageComment: "no comment",
			ImageTagNames: []string{
				"no tag",
			},
		},
		CorrelationId: "dfsfdsf",
		AppId:         1,
		EnvId:         2,
	}

	type args struct {
		event                 Event
		approvalActionRequest bean.UserApprovalActionRequest
		cdPipeline            *pipelineConfig.Pipeline
		userId                int32
	}
	tests := []struct {
		name    string
		args    args
		want    Event
		wantErr bool
	}{

		{
			name: "SES config not found",
			args: args{
				event:                 event,
				approvalActionRequest: approvalActionRequest,
				cdPipeline:            cdPipeline,
				userId:                int32(1),
			},
			want:    eventResponse1, // Expect the same event as SES config not found.
			wantErr: false,
		},
		{
			name: "Image comment not found",
			args: args{
				event:                 event,
				approvalActionRequest: approvalActionRequest,
				cdPipeline:            cdPipeline,
				userId:                int32(1),
			},
			want:    eventResponse2, // Expect the same event as image comment not found.
			wantErr: false,
		},
		{
			name: "Image tags not found",
			args: args{
				event:                 event,
				approvalActionRequest: approvalActionRequest,
				cdPipeline:            cdPipeline,
				userId:                int32(1),
			},
			want:    eventResponse3, // Expect the same event as image tags not found.
			wantErr: false,
		},
		{
			name: "CI artifact not found",
			args: args{
				event:                 event,
				approvalActionRequest: approvalActionRequest,
				cdPipeline:            cdPipeline,
				userId:                int32(1),
			},
			want:    eventResponse4, // Expect the same event as CI artifact not found.
			wantErr: false,
		},
		{
			name: "User ID is 0",
			args: args{
				event:                 event,
				approvalActionRequest: approvalActionRequest,
				cdPipeline:            cdPipeline,
				userId:                int32(0),
			},
			want:    eventResponse5, // Expect the event with empty TriggeredBy field.
			wantErr: false,
		},
		{
			name: "Valid case with all data available",
			args: args{
				event:                 event,
				approvalActionRequest: approvalActionRequest,
				cdPipeline:            cdPipeline,
				userId:                int32(1),
			},
			want:    eventResponse,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impl, ciArtifactRepository, userRepository, sesNotificationRepository, _, imageTaggingRepository := InitEventSimpleFactoryImpl(t)
			var err error
			if tt.name == "SES config not found" {
				sesNotificationRepository.On("FindDefault").Return(nil, errors.New("ses not found"))
			}
			if tt.name == "Image comment not found" {
				sesNotificationRepository.On("FindDefault").Return(defaultSesConfig, nil)
				imageTaggingRepository.On("GetImageComment", approvalActionRequest.ArtifactId).Return(repository.ImageComment{}, errors.New("image comments not found"))

			}
			if tt.name == "Image tags not found" {
				sesNotificationRepository.On("FindDefault").Return(defaultSesConfig, nil)
				imageTaggingRepository.On("GetImageComment", approvalActionRequest.ArtifactId).Return(imageComment, nil)
				imageTaggingRepository.On("GetTagsByArtifactId", approvalActionRequest.ArtifactId).Return(nil, errors.New("image tags not found"))
			}
			if tt.name == "CI artifact not found" {
				sesNotificationRepository.On("FindDefault").Return(defaultSesConfig, nil)
				imageTaggingRepository.On("GetImageComment", approvalActionRequest.ArtifactId).Return(imageComment, nil)
				imageTaggingRepository.On("GetTagsByArtifactId", approvalActionRequest.ArtifactId).Return(imageTags, nil)
				ciArtifactRepository.On("Get", approvalActionRequest.ArtifactId).Return(nil, errors.New("ci-artifact not found"))
			}
			if tt.name == "User ID is 0" {
				sesNotificationRepository.On("FindDefault").Return(defaultSesConfig, nil)
				imageTaggingRepository.On("GetImageComment", approvalActionRequest.ArtifactId).Return(imageComment, nil)
				imageTaggingRepository.On("GetTagsByArtifactId", approvalActionRequest.ArtifactId).Return(imageTags, nil)
				ciArtifactRepository.On("Get", approvalActionRequest.ArtifactId).Return(&ciArtifact, nil)

			}
			if tt.name == "Valid case with all data available" {
				//smtpNotificationRepository.On("FindDefault").Return(defaultSmtpConfig, nil)
				sesNotificationRepository.On("FindDefault").Return(defaultSesConfig, nil)
				imageTaggingRepository.On("GetImageComment", approvalActionRequest.ArtifactId).Return(imageComment, nil)
				imageTaggingRepository.On("GetTagsByArtifactId", approvalActionRequest.ArtifactId).Return(imageTags, nil)
				ciArtifactRepository.On("Get", approvalActionRequest.ArtifactId).Return(&ciArtifact, nil)
				userRepository.On("GetById", int32(1)).Return(&user, nil)
			}
			got := impl.BuildExtraApprovalData(tt.args.event, tt.args.approvalActionRequest, tt.args.cdPipeline, tt.args.userId)

			if tt.wantErr {
				// Verify error scenario
				if err == nil {
					t.Errorf("Expected error, but got nil")
				}
			} else {
				// Verify normal scenario
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("BuildExtraApprovalData() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func InitEventSimpleFactoryImpl(t *testing.T) (*EventSimpleFactoryImpl, *mocks4.CiArtifactRepository, *mocks3.UserRepository, *mocks4.SESNotificationRepository, *mocks4.SMTPNotificationRepository, *mocks5.ImageTaggingRepository) {
	logger, _ := util2.NewSugardLogger()
	cdWorkflowRepository := mocks2.NewCdWorkflowRepository(t)
	pipelineOverrideRepository := mocks.NewPipelineOverrideRepository(t)
	ciWorkflowRepository := mocks2.NewCiWorkflowRepository(t)
	ciPipelineMaterialRepository := mocks2.NewCiPipelineMaterialRepository(t)
	ciPipelineRepository := mocks2.NewCiPipelineRepository(t)
	pipelineRepository := mocks2.NewPipelineRepository(t)
	userRepository := mocks3.NewUserRepository(t)
	ciArtifactRepository := mocks4.NewCiArtifactRepository(t)
	DeploymentApprovalRepository := mocks2.NewDeploymentApprovalRepository(t)
	sesNotificationRepository := mocks4.NewSESNotificationRepository(t)
	smtpNotificationRepository := mocks4.NewSMTPNotificationRepository(t)
	imageTaggingRepository := mocks5.NewImageTaggingRepository(t)
	impl := &EventSimpleFactoryImpl{
		logger:                       logger,
		cdWorkflowRepository:         cdWorkflowRepository,
		pipelineOverrideRepository:   pipelineOverrideRepository,
		ciWorkflowRepository:         ciWorkflowRepository,
		ciPipelineMaterialRepository: ciPipelineMaterialRepository,
		ciPipelineRepository:         ciPipelineRepository,
		pipelineRepository:           pipelineRepository,
		userRepository:               userRepository,
		ciArtifactRepository:         ciArtifactRepository,
		DeploymentApprovalRepository: DeploymentApprovalRepository,
		sesNotificationRepository:    sesNotificationRepository,
		smtpNotificationRepository:   smtpNotificationRepository,
		imageTaggingRepository:       imageTaggingRepository,
	}

	return impl, ciArtifactRepository, userRepository, sesNotificationRepository, smtpNotificationRepository, imageTaggingRepository
}
