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

package helper

import (
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	bean2 "github.com/devtron-labs/devtron/pkg/auth/user/bean"
	"github.com/devtron-labs/devtron/util"
)

func GetQueryForUserListingWithFilters(req *bean.ListingRequest) (string, []interface{}) {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := ""
	var queryParams []interface{}
	if len(req.SearchKey) > 0 {
		whereCondition += " AND email_id ilike ? "
		queryParams = append(queryParams, util.GetLIKEClauseQueryParam(req.SearchKey))
	}

	if len(req.SortBy) > 0 && !req.CountCheck {
		orderCondition += " order by "
		// Handling it for last login as it is time and show order differs on UI.
		if req.SortBy == bean2.LastLogin && req.SortOrder == bean2.Asc {
			orderCondition += fmt.Sprintf(" %s %s ", bean2.LastLogin, bean2.Desc)
		} else {
			orderCondition += fmt.Sprintf(" %s ", bean2.LastLogin)
		}
		if req.SortBy == bean2.Email && req.SortOrder == bean2.Desc {
			orderCondition += fmt.Sprintf(" %s %s ", bean2.Email, bean2.Desc)
		} else {
			orderCondition += fmt.Sprintf(" %s ", bean2.Email)
		}
	}

	if req.Size > 0 && !req.CountCheck && !req.ShowAll {
		orderCondition += " limit ? offset ? "
		queryParams = append(queryParams, req.Size, req.Offset)
	}
	var query string
	if req.CountCheck {
		query = fmt.Sprintf("select count(*) from users AS user_model left join user_audit AS au on au.user_id=user_model.id %s %s;", whereCondition, orderCondition)
	} else {
		// have not collected client ip here. always will be empty
		query = fmt.Sprintf(`SELECT "user_model".*, "user_audit"."id" AS "user_audit__id", "user_audit"."updated_on" AS "user_audit__updated_on","user_audit"."user_id" AS "user_audit__user_id" ,"user_audit"."created_on" AS "user_audit__created_on" ,"user_audit"."updated_on" AS "last_login" from users As "user_model" LEFT JOIN user_audit As "user_audit" on "user_audit"."user_id" = "user_model"."id" %s %s;`, whereCondition, orderCondition)
	}

	return query, queryParams
}

func GetQueryForAllUserWithAudit() string {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := fmt.Sprintf("order by user_model.updated_on %s", bean2.Desc)
	query := fmt.Sprintf(`SELECT "user_model".*, "user_audit"."id" AS "user_audit__id", "user_audit"."updated_on" AS "user_audit__updated_on","user_audit"."user_id" AS "user_audit__user_id" ,"user_audit"."created_on" AS "user_audit__created_on" from users As "user_model" LEFT JOIN user_audit As "user_audit" on "user_audit"."user_id" = "user_model"."id" %s %s;`, whereCondition, orderCondition)
	return query
}

func GetQueryForGroupListingWithFilters(req *bean.ListingRequest) (string, []interface{}) {
	var queryParams []interface{}
	whereCondition := " where active = ? "
	queryParams = append(queryParams, true)
	if len(req.SearchKey) > 0 {
		whereCondition += " AND name ilike ? "
		queryParams = append(queryParams, util.GetLIKEClauseQueryParam(req.SearchKey))
	}

	orderCondition := ""
	if len(req.SortBy) > 0 && !req.CountCheck {
		orderCondition += fmt.Sprintf(" order by %s ", req.SortBy)
		if req.SortOrder == bean2.Desc {
			orderCondition += fmt.Sprintf(" %s ", bean2.Desc)
		}
	}
	if req.Size > 0 && !req.CountCheck && !req.ShowAll {
		orderCondition += " limit ? offset ? "
		queryParams = append(queryParams, req.Size, req.Offset)
	}
	var query string
	if req.CountCheck {
		query = fmt.Sprintf("SELECT count(*) from role_group %s %s;", whereCondition, orderCondition)
	} else {
		query = fmt.Sprintf("SELECT * from role_group %s %s;", whereCondition, orderCondition)
	}
	return query, queryParams

}

func GetEmailSearchQuery(usersTableAlias string, emailId string) string {
	expression := fmt.Sprintf(
		"( (%s.user_type is NULL and %s.email_id ILIKE '%s' ) or (%s.user_type='apiToken' and %s.email_id='%s') )",
		usersTableAlias, usersTableAlias, emailId, usersTableAlias, usersTableAlias, emailId)
	return expression
}
