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
	"strconv"
)

func GetQueryForUserListingWithFilters(req *bean.ListingRequest) string {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := ""

	if len(req.SearchKey) > 0 {
		emailIdLike := "%" + req.SearchKey + "%"
		whereCondition += fmt.Sprintf("AND email_id ilike '%s' ", emailIdLike)
	}

	if len(req.SortBy) > 0 && !req.CountCheck {
		orderCondition += fmt.Sprintf("order by %s ", req.SortBy)
		// Handling it for last login as it is time and show order differs on UI.
		if req.SortBy == bean2.LastLogin && req.SortOrder == bean2.Asc {
			orderCondition += string(bean2.Desc)
		}
		if req.SortBy == bean2.Email && req.SortOrder == bean2.Desc {
			orderCondition += string(req.SortOrder)
		}
	}

	if req.Size > 0 && !req.CountCheck && !req.ShowAll {
		orderCondition += " limit " + strconv.Itoa(req.Size) + " offset " + strconv.Itoa(req.Offset) + ""
	}
	var query string
	if req.CountCheck {
		query = fmt.Sprintf(`select count(*) from users AS user_model left join lateral (select ua.* from user_audit ua where ua.user_id=user_model.id order by ua.updated_on desc limit 1) AS au ON true left join timeout_window_configuration AS timeout_window_configuration ON timeout_window_configuration.id = user_model.timeout_window_configuration_id %s %s;`, whereCondition, orderCondition)
	} else {
		// have not collected client ip here. always will be empty
		query = fmt.Sprintf(`SELECT user_model.*, timeout_window_configuration.id AS timeout_window_configuration__id, timeout_window_configuration.timeout_window_expression AS timeout_window_configuration__timeout_window_expression,timeout_window_configuration.timeout_window_expression_format AS timeout_window_configuration__timeout_window_expression_format, au.id AS user_audit__id, au.updated_on AS user_audit__updated_on, au.user_id AS user_audit__user_id, au.created_on AS user_audit__created_on,au.updated_on AS last_login FROM users AS user_model left join lateral (select ua.* from user_audit ua where ua.user_id=user_model.id order by ua.updated_on desc limit 1) au ON true LEFT JOIN timeout_window_configuration AS timeout_window_configuration ON timeout_window_configuration.id = user_model.timeout_window_configuration_id %s %s;`, whereCondition, orderCondition)
	}

	return query
}

func GetQueryForAllUserWithAudit() string {
	whereCondition := fmt.Sprintf("where active = %t AND (user_type is NULL or user_type != '%s') ", true, bean.USER_TYPE_API_TOKEN)
	orderCondition := fmt.Sprintf("order by user_model.updated_on %s", bean2.Desc)
	query := fmt.Sprintf(`SELECT "user_model".*, "user_audit"."id" AS "user_audit__id", "user_audit"."updated_on" AS "user_audit__updated_on","user_audit"."user_id" AS "user_audit__user_id" ,"user_audit"."created_on" AS "user_audit__created_on" from users As "user_model" LEFT JOIN user_audit As "user_audit" on "user_audit"."user_id" = "user_model"."id" %s %s;`, whereCondition, orderCondition)
	return query
}

func GetQueryForGroupListingWithFilters(req *bean.ListingRequest) string {
	whereCondition := fmt.Sprintf("where active = %t ", true)
	orderCondition := ""
	if len(req.SearchKey) > 0 {
		nameIdLike := "%" + req.SearchKey + "%"
		whereCondition += fmt.Sprintf("AND name ilike '%s' ", nameIdLike)
	}

	if len(req.SortBy) > 0 && !req.CountCheck {
		orderCondition += fmt.Sprintf("order by %s ", req.SortBy)
		if req.SortOrder == bean2.Desc {
			orderCondition += string(req.SortOrder)
		}
	}

	if req.Size > 0 && !req.CountCheck && !req.ShowAll {
		orderCondition += " limit " + strconv.Itoa(req.Size) + " offset " + strconv.Itoa(req.Offset) + ""
	}
	var query string
	if req.CountCheck {
		query = fmt.Sprintf("SELECT count(*) from role_group %s %s;", whereCondition, orderCondition)
	} else {
		query = fmt.Sprintf("SELECT * from role_group %s %s;", whereCondition, orderCondition)
	}
	return query

}

func GetEmailSearchQuery(usersTableAlias string, emailId string) string {
	expression := fmt.Sprintf(
		"( (%s.user_type is NULL and %s.email_id ILIKE '%s' ) or (%s.user_type='apiToken' and %s.email_id='%s') )",
		usersTableAlias, usersTableAlias, emailId, usersTableAlias, usersTableAlias, emailId)
	return expression
}
