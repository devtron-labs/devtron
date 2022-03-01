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
func (impl NotificationRouterImpl) InitNotificationRegRouter(configRouter *mux.Router) {
	configRouter.Path("").
		HandlerFunc(impl.notificationRestHandler.SaveNotificationSettings).
		Methods("POST")
	configRouter.Path("").
		HandlerFunc(impl.notificationRestHandler.UpdateNotificationSettings).
		Methods("PUT")
	configRouter.Path("").
		Queries("size", "{size}").
		Queries("offset", "{offset}").
		HandlerFunc(impl.notificationRestHandler.GetAllNotificationSettings).
		Methods("GET")
	configRouter.Path("").
		HandlerFunc(impl.notificationRestHandler.DeleteNotificationSettings).
		Methods("DELETE")

	configRouter.Path("/channel").
		HandlerFunc(impl.notificationRestHandler.SaveNotificationChannelConfig).
		Methods("POST")
	configRouter.Path("/channel").
		HandlerFunc(impl.notificationRestHandler.FindAllNotificationConfig).
		Methods("GET")
	configRouter.Path("/channel/ses/{id}").
		HandlerFunc(impl.notificationRestHandler.FindSESConfig).
		Methods("GET")
	configRouter.Path("/channel/slack/{id}").
		HandlerFunc(impl.notificationRestHandler.FindSlackConfig).
		Methods("GET")
	configRouter.Path("/channel").
		HandlerFunc(impl.notificationRestHandler.DeleteNotificationChannelConfig).
		Methods("DELETE")

	configRouter.Path("/recipient").
		Queries("value", "{value}").
		HandlerFunc(impl.notificationRestHandler.RecipientListingSuggestion).
		Methods("GET")
	configRouter.Path("/channel/autocomplete/{type}").
		HandlerFunc(impl.notificationRestHandler.FindAllNotificationConfigAutocomplete).
		Methods("GET")
	configRouter.Path("/search").
		HandlerFunc(impl.notificationRestHandler.GetOptionsForNotificationSettings).
		Methods("POST")

}
