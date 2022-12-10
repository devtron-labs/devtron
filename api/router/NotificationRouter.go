/*
 * Copyright (c) 2020 Devtron Labs
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package router

import (
	"github.com/devtron-labs/devtron/api/restHandler"
	"github.com/gorilla/mux"
)

type NotificationRouter interface {
	InitNotificationRegRouter(gocdRouter *mux.Router)
}
type NotificationRouterImpl struct {
	notificationRestHandler restHandler.NotificationRestHandler
}

func NewNotificationRouterImpl(notificationRestHandler restHandler.NotificationRestHandler) *NotificationRouterImpl {
	return &NotificationRouterImpl{notificationRestHandler: notificationRestHandler}
}
func (router *NotificationRouterImpl) InitNotificationRegRouter(configRouter *mux.Router) {
	configRouter.Path("").
		HandlerFunc(router.notificationRestHandler.SaveNotificationSettings).
		Methods("POST")
	configRouter.Path("").
		HandlerFunc(router.notificationRestHandler.UpdateNotificationSettings).
		Methods("PUT")
	configRouter.Path("").
		Queries("size", "{size}").
		Queries("offset", "{offset}").
		HandlerFunc(router.notificationRestHandler.GetAllNotificationSettings).
		Methods("GET")
	configRouter.Path("").
		HandlerFunc(router.notificationRestHandler.DeleteNotificationSettings).
		Methods("DELETE")

	configRouter.Path("/channel").
		HandlerFunc(router.notificationRestHandler.SaveNotificationChannelConfig).
		Methods("POST")
	configRouter.Path("/channel").
		HandlerFunc(router.notificationRestHandler.FindAllNotificationConfig).
		Methods("GET")
	configRouter.Path("/channel/ses/{id}").
		HandlerFunc(router.notificationRestHandler.FindSESConfig).
		Methods("GET")
	configRouter.Path("/channel/slack/{id}").
		HandlerFunc(router.notificationRestHandler.FindSlackConfig).
		Methods("GET")
	configRouter.Path("/channel/smtp/{id}").
		HandlerFunc(router.notificationRestHandler.FindSMTPConfig).
		Methods("GET")
	configRouter.Path("/channel").
		HandlerFunc(router.notificationRestHandler.DeleteNotificationChannelConfig).
		Methods("DELETE")

	configRouter.Path("/recipient").
		Queries("value", "{value}").
		HandlerFunc(router.notificationRestHandler.RecipientListingSuggestion).
		Methods("GET")
	configRouter.Path("/channel/autocomplete/{type}").
		HandlerFunc(router.notificationRestHandler.FindAllNotificationConfigAutocomplete).
		Methods("GET")
	configRouter.Path("/search").
		HandlerFunc(router.notificationRestHandler.GetOptionsForNotificationSettings).
		Methods("POST")

}
