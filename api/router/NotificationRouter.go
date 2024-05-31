/*
 * Copyright (c) 2020-2024. Devtron Inc.
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
	configRouter.Path("/channel/config").
		HandlerFunc(impl.notificationRestHandler.IsSesOrSmtpConfigured).
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
	configRouter.Path("/channel/smtp/{id}").
		HandlerFunc(impl.notificationRestHandler.FindSMTPConfig).
		Methods("GET")
	configRouter.Path("/channel/webhook/{id}").
		HandlerFunc(impl.notificationRestHandler.FindWebhookConfig).
		Methods("GET")
	configRouter.Path("/variables").
		HandlerFunc(impl.notificationRestHandler.GetWebhookVariables).
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
	configRouter.Path("/channel/config/approve").
		HandlerFunc(impl.notificationRestHandler.ApproveConfigDraftForNotification).
		Methods("POST")
	configRouter.Path("/channel/deployment/approve").
		HandlerFunc(impl.notificationRestHandler.ApproveDeploymentConfigForNotification).
		Methods("POST")

	configRouter.Path("/channel/image-promotion/approve").
		HandlerFunc(impl.notificationRestHandler.ApproveArtifactPromotion).
		Methods("POST")

}
