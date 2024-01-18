package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type UserListingRepositoryQueryBuilder struct {
	logger *zap.SugaredLogger
}

func NewUserListingRepositoryQueryBuilder(logger *zap.SugaredLogger) UserListingRepositoryQueryBuilder {
	userListingRepositoryQueryBuilder := UserListingRepositoryQueryBuilder{
		logger: logger,
	}
	return userListingRepositoryQueryBuilder
}

type SortBy string
type SortOrder string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

const (
	Email     SortBy = "email_id"
	LastLogin SortBy = "last_login"
	GroupName SortBy = "name"
)

type FetchListingRequest struct {
	Status      bean.Status `json:"status"`
	SearchKey   string      `json:"searchKey"`
	SortOrder   SortOrder   `json:"sortOrder"`
	SortBy      SortBy      `json:"sortBy"`
	Offset      int         `json:"offset"`
	Size        int         `json:"size"`
	ShowAll     bool        `json:"showAll"`
	CurrentTime time.Time   `json:"-"` // for Internal Use
}

func (impl UserListingRepositoryQueryBuilder) GetStatusFromTTL(ttl, recordedTime time.Time) bean.Status {
	if ttl.IsZero() || ttl.After(recordedTime) {
		return bean.Active
	}
	return bean.Inactive
}

func (impl UserListingRepositoryQueryBuilder) GetQueryForUserListingWithFilters(req *FetchListingRequest) string {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := ""
	if req.Status == bean.Active {
		whereCondition += "AND (time_to_live is null OR time_to_live > %s ::timestamp AT TIME ZONE 'UTC') "
	} else if req.Status == bean.Inactive {
		whereCondition += fmt.Sprintf("AND time_to_live < %s ::timestamp AT TIME ZONE 'UTC'", req.CurrentTime)
	} else if req.Status == bean.TemporaryAccess {
		whereCondition += fmt.Sprintf(" AND time_to_live > %s ::timestamp AT TIME ZONE 'UTC' ", req.CurrentTime)
	}
	if len(req.SearchKey) > 0 {
		emailIdLike := "%" + req.SearchKey + "%"
		whereCondition += fmt.Sprintf("AND email_id like '%s' ", emailIdLike)
	}

	if req.SortBy == Email {
		orderCondition += fmt.Sprintf("order by %s ", req.SortBy)
		if req.SortOrder == Desc {
			orderCondition += string(req.SortOrder)
		}
	}

	if req.Size > 0 {
		orderCondition += " limit " + strconv.Itoa(req.Size) + " offset " + strconv.Itoa(req.Offset) + ""
	}

	query := fmt.Sprintf("SELECT * FROM USERS %s %s;", whereCondition, orderCondition)
	return query
}
