package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
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
	Email     helper.SortBy = "email_id"
	LastLogin helper.SortBy = "last_login"
	GroupName helper.SortBy = "name"
)

type Status string

const (
	Active   Status = "active"
	Inactive Status = "inactive"
)

type FetchListingRequest struct {
	Status    Status    `json:"status"`
	TTL       time.Time `json:"timeToLive,omitempty"`
	SearchKey string    `json:"searchKey"`
	SortOrder SortOrder `json:"sortOrder"`
	SortBy    SortBy    `json:"sortBy"`
	Offset    int       `json:"offset"`
	Size      int       `json:"size"`
	ShowAll   bool      `json:"showAll"`
}

func (impl UserListingRepositoryQueryBuilder) GetQueryForUserListingWithFilters(req *FetchListingRequest) string {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := ""
	if len(req.Status) > 0 {
		whereCondition += fmt.Sprintf("AND status = '%s' ", req.Status)
		if !req.TTL.IsZero() {
			whereCondition += fmt.Sprintf("AND ttl < %s ", req.TTL)
		}
	}
	if len(req.SearchKey) > 0 {
		emailIdLike := "%" + req.SearchKey + "%"
		whereCondition += fmt.Sprintf("AND email_id like '%s' ", emailIdLike)
	}

	if helper.SortBy(req.SortBy) == Email {
		orderCondition += fmt.Sprintf("order by %s ", req.SortBy)
		if req.SortOrder == Desc {
			orderCondition += string(req.SortOrder)
		}
	}

	if req.Size > 0 {
		orderCondition += " limit " + strconv.Itoa(req.Size) + " offset " + strconv.Itoa(req.Offset) + ""
	}

	//query := fmt.Sprintf("SELECT u.*, ua.updated_on as last_login FROM USERS u LEFT OUTER JOIN user_audit ua on u.id = ua.user_id %s %s;", whereCondition, orderCondition)
	query := fmt.Sprintf("SELECT * FROM USERS %s %s;", whereCondition, orderCondition)
	return query
}
