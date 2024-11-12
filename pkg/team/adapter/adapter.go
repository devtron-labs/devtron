package adapter

import (
	"github.com/devtron-labs/devtron/pkg/sql"
	"github.com/devtron-labs/devtron/pkg/team/bean"
	"github.com/devtron-labs/devtron/pkg/team/repository"
)

func ConvertTeamDBObjToDTO(teamDBObj *repository.Team) *bean.TeamRequest {
	return &bean.TeamRequest{
		Id:        teamDBObj.Id,
		Name:      teamDBObj.Name,
		Active:    teamDBObj.Active,
		CreatedOn: teamDBObj.CreatedOn,
		UserId:    teamDBObj.CreatedBy,
	}
}

func ConvertTeamDTOToDbObj(teamDTO *bean.TeamRequest) *repository.Team {
	return &repository.Team{
		Id:     teamDTO.Id,
		Name:   teamDTO.Name,
		Active: teamDTO.Active,
		AuditLog: sql.AuditLog{
			CreatedOn: teamDTO.CreatedOn,
			CreatedBy: teamDTO.UserId,
			UpdatedOn: teamDTO.CreatedOn,
			UpdatedBy: teamDTO.UserId,
		},
	}
}
