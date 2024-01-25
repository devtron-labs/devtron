package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	bean3 "github.com/devtron-labs/devtron/pkg/timeoutWindow/repository/bean"
	"go.uber.org/zap"
	"strconv"
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

const (
	QueryTimeFormat      string = "2006-01-02 15:04:05-07:00"
	TimeStampFormat      string = "YYYY-MM-DD HH24:MI:SS"
	TimeFormatForParsing string = "2006-01-02 15:04:05 -0700 MST"
)

//func (impl UserRepositoryQueryBuilder) GetStatusFromTTL(ttl, recordedTime time.Time) bean.Status {
//	if ttl.IsZero() || ttl.After(recordedTime) {
//		return bean.Active
//	}
//	return bean.Inactive
//}
//
//func (impl UserRepositoryQueryBuilder) GetStatusFromTimeString(timeString string, recordedTime time.Time) (bean.Status, error) {
//	formattedTime, err := time.Parse(time.RFC3339, timeString)
//	if err != nil {
//		return bean.Unknown, err
//	}
//	return impl.GetStatusFromTTL(formattedTime, recordedTime), nil
//}

func (impl UserRepositoryQueryBuilder) GetQueryForUserListingWithFilters(req *bean.FetchListingRequest) string {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := ""
	//formatted for query comparison
	formattedTimeForQuery := req.CurrentTime.Format(QueryTimeFormat)
	// Have handled both formats 1 and 2 in the query for user inactive status
	if req.Status == bean.Active {
		whereCondition += fmt.Sprintf("AND (user_model.timeout_window_configuration_id is null OR ( timeout_window_configuration.timeout_window_expression_format = %v AND TO_TIMESTAMP(timeout_window_configuration.timeout_window_expression,'%s') > '%s' ) ) ", bean3.TimeStamp, TimeStampFormat, formattedTimeForQuery) //  TODO: replace 1 with expressionformat const
	} else if req.Status == bean.Inactive {
		whereCondition += fmt.Sprintf("AND ( TO_TIMESTAMP(timeout_window_configuration.timeout_window_expression,'%s') < '%s' ) ", TimeStampFormat, formattedTimeForQuery)
	} else if req.Status == bean.TemporaryAccess {
		whereCondition += fmt.Sprintf(" AND (timeout_window_configuration.timeout_window_expression_format = %v AND TO_TIMESTAMP(timeout_window_configuration.timeout_window_expression,'%s') > '%s' ) ", bean3.TimeStamp, TimeStampFormat, formattedTimeForQuery)
	}
	if len(req.SearchKey) > 0 {
		emailIdLike := "%" + req.SearchKey + "%"
		whereCondition += fmt.Sprintf("AND email_id like '%s' ", emailIdLike)
	}

	if len(req.SortBy) > 0 && req.Size > 0 {
		orderCondition += fmt.Sprintf("order by %s ", req.SortBy)
		if req.SortOrder == bean2.Desc {
			orderCondition += string(req.SortOrder)
		}
	}

	if req.Size > 0 {
		orderCondition += " limit " + strconv.Itoa(req.Size) + " offset " + strconv.Itoa(req.Offset) + ""
	}
	var query string
	if req.Size == 0 {
		query = fmt.Sprintf("select count(*) from users AS user_model left join user_audit AS au on au.user_id=user_model.id left join timeout_window_configuration AS timeout_window_configuration on timeout_window_configuration.id=user_model.timeout_window_configuration_id %s %s;", whereCondition, orderCondition)
	} else {
		// have not collected client ip here. always will be empty
		query = fmt.Sprintf("SELECT \"user_model\".*, \"timeout_window_configuration\".\"id\" AS \"timeout_window_configuration__id\", \"timeout_window_configuration\".\"timeout_window_expression\" AS \"timeout_window_configuration__timeout_window_expression\", \"timeout_window_configuration\".\"timeout_window_expression_format\" AS \"timeout_window_configuration__timeout_window_expression_format\", \"user_audit\".\"id\" AS \"user_audit__id\", \"user_audit\".\"updated_on\" AS \"user_audit__updated_on\",\"user_audit\".\"user_id\" AS \"user_audit__user_id\" ,\"user_audit\".\"created_on\" AS \"user_audit__created_on\" ,\"user_audit\".\"updated_on\" AS \"last_login\" from users As \"user_model\" LEFT JOIN user_audit As \"user_audit\" on \"user_audit\".\"user_id\" = \"user_model\".\"id\" LEFT JOIN timeout_window_configuration AS \"timeout_window_configuration\" ON \"timeout_window_configuration\".\"id\" = \"user_model\".\"timeout_window_configuration_id\" %s %s;", whereCondition, orderCondition)
	}

	return query
}

func (impl UserRepositoryQueryBuilder) GetQueryForAllUserWithAudit() string {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := fmt.Sprintf("order by user_model.updated_on %s", bean2.Desc)
	query := fmt.Sprintf("SELECT \"user_model\".*, \"timeout_window_configuration\".\"id\" AS \"timeout_window_configuration__id\", \"timeout_window_configuration\".\"timeout_window_expression\" AS \"timeout_window_configuration__timeout_window_expression\", \"timeout_window_configuration\".\"timeout_window_expression_format\" AS \"timeout_window_configuration__timeout_window_expression_format\", \"user_audit\".\"id\" AS \"user_audit__id\", \"user_audit\".\"updated_on\" AS \"user_audit__updated_on\",\"user_audit\".\"user_id\" AS \"user_audit__user_id\" ,\"user_audit\".\"created_on\" AS \"user_audit__created_on\" from users As \"user_model\" LEFT JOIN user_audit As \"user_audit\" on \"user_audit\".\"user_id\" = \"user_model\".\"id\" LEFT JOIN timeout_window_configuration AS \"timeout_window_configuration\" ON \"timeout_window_configuration\".\"id\" = \"user_model\".\"timeout_window_configuration_id\" %s %s;", whereCondition, orderCondition)
	return query
}

//func (impl UserRepositoryQueryBuilder) GetQueryForBulkUpdate(req *bean.BulkStatusUpdateRequest) string {
//	ttlTime := req.CurrentTime.Format(QueryTimeFormat)
//	if req.Status == bean.Active {
//		if !req.TimeToLive.IsZero() {
//			ttlTime = fmt.Sprintf("'%s'", req.TimeToLive.Format(QueryTimeFormat))
//		} else {
//			ttlTime = "null"
//		}
//	} else if req.Status == bean.Inactive {
//		ttlTime = fmt.Sprintf("CURRENT_TIMESTAMP - INTERVAL '1 day' ")
//	}
//
//	//whereCondition := " where id in (" + helper.GetCommaSepratedString(req.UserIds) + ") "
//	whereCondition := fmt.Sprintf("AND (user_type is NULL or user_type != '%s');", bean.USER_TYPE_API_TOKEN)
//	query := fmt.Sprintf("UPDATE USERS set time_to_live= %s %s", ttlTime, whereCondition)
//	return query
//}
