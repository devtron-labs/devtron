/*
 * Copyright (c) 2020-2024. Devtron Inc.
 */

package session

import (
	"context"
	"github.com/argoproj/argo-cd/v2/pkg/apiclient/session"
	"github.com/devtron-labs/devtron/client/argocdServer/connection"
	"time"
)

type ServiceClient interface {
	Create(ctxt context.Context, userName string, password string) (string, error)
}

type ServiceClientImpl struct {
	ssc session.SessionServiceClient
}

func NewSessionServiceClient(argoCDConnectionManager connection.ArgoCDConnectionManager) *ServiceClientImpl {
	// this function only called when gitops configured and user ask for creating acd token
	conn := argoCDConnectionManager.GetConnection("")
	ssc := session.NewSessionServiceClient(conn)
	return &ServiceClientImpl{ssc: ssc}
}

func (c *ServiceClientImpl) Create(ctxt context.Context, userName string, password string) (string, error) {
	session := session.SessionCreateRequest{
		Username: userName,
		Password: password,
	}
	ctx, cancel := context.WithTimeout(ctxt, 100*time.Second)
	defer cancel()
	resp, err := c.ssc.Create(ctx, &session)
	if err != nil {
		return "", err
	}
	//argocdServer.SetTokenAuth(resp.Token)
	//fmt.Printf("%+v\n", resp)
	return resp.Token, nil
}
