package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"github.com/devtron-labs/devtron/internal/sql/repository/helper"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type UserRepositoryQueryBuilder struct {
	logger *zap.SugaredLogger
}

func NewUserRepositoryQueryBuilder(logger *zap.SugaredLogger) UserRepositoryQueryBuilder {
	userListingRepositoryQueryBuilder := UserRepositoryQueryBuilder{
		logger: logger,
	}
	return userListingRepositoryQueryBuilder
}

const QueryTimeFormat = "2006-01-02 15:04:05-07:00"

func (impl UserRepositoryQueryBuilder) GetStatusFromTTL(ttl, recordedTime time.Time) bean.Status {
	if ttl.IsZero() || ttl.After(recordedTime) {
		return bean.Active
	}
	return bean.Inactive
}

func (impl UserRepositoryQueryBuilder) GetQueryForUserListingWithFilters(req *bean.FetchListingRequest) string {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := ""
	// formatted for query comparison
	formattedTimeForQuery := req.CurrentTime.Format(QueryTimeFormat)
	if req.Status == bean.Active {
		whereCondition += fmt.Sprintf("AND (time_to_live is null OR time_to_live > '%s' ) ", formattedTimeForQuery)
	} else if req.Status == bean.Inactive {
		whereCondition += fmt.Sprintf("AND time_to_live < '%s' ", formattedTimeForQuery)
	} else if req.Status == bean.TemporaryAccess {
		whereCondition += fmt.Sprintf(" AND time_to_live > '%s' ", formattedTimeForQuery)
	}
	if len(req.SearchKey) > 0 {
		emailIdLike := "%" + req.SearchKey + "%"
		whereCondition += fmt.Sprintf("AND email_id like '%s' ", emailIdLike)
	}

	if req.SortBy == bean2.Email {
		orderCondition += fmt.Sprintf("order by %s ", req.SortBy)
		if req.SortOrder == bean2.Desc {
			orderCondition += string(req.SortOrder)
		}
	}

	if req.Size > 0 {
		orderCondition += " limit " + strconv.Itoa(req.Size) + " offset " + strconv.Itoa(req.Offset) + ""
	}

	query := fmt.Sprintf("SELECT * FROM USERS %s %s;", whereCondition, orderCondition)
	return query
}

func (impl UserRepositoryQueryBuilder) GetQueryForBulkUpdate(req *bean.BulkStatusUpdateRequest) string {
	ttlTime := req.CurrentTime.Format(QueryTimeFormat)
	if req.Status == bean.Active {
		if !req.TimeToLive.IsZero() {
			ttlTime = fmt.Sprintf("'%s'", req.TimeToLive.Format(QueryTimeFormat))
		} else {
			ttlTime = "null"
		}
	} else if req.Status == bean.Inactive {
		ttlTime = fmt.Sprintf("CURRENT_TIMESTAMP - INTERVAL '1 day' ")
	}

	whereCondition := " where id in (" + helper.GetCommaSepratedString(req.UserIds) + ") "
	whereCondition += fmt.Sprintf("AND (user_type is NULL or user_type != '%s');", bean.USER_TYPE_API_TOKEN)
	query := fmt.Sprintf("UPDATE USERS set time_to_live= %s %s", ttlTime, whereCondition)
	return query
}
