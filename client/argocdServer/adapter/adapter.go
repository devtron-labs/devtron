package adapter

import (
	"github.com/argoproj/argo-cd/v2/pkg/apis/application/v1alpha1"
	"github.com/devtron-labs/devtron/client/argocdServer/bean"
)

func GetArgoCdPatchReqFromDto(dto *bean.ArgoCdAppPatchReqDto) v1alpha1.Application {
	return v1alpha1.Application{
		Spec: v1alpha1.ApplicationSpec{
			Source: &v1alpha1.ApplicationSource{
				Path:           dto.ChartLocation,
				RepoURL:        dto.GitRepoUrl,
				TargetRevision: dto.TargetRevision,
			},
		},
	}
}
