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

package session

import (
	"context"
	"github.com/argoproj/argo-cd/pkg/apiclient/session"
	"github.com/argoproj/argo-cd/util/settings"
	"github.com/devtron-labs/devtron/client/argocdServer"
	"time"
)

type ServiceClient interface {
	Create(ctxt context.Context, userName string, password string) (string, error)
}

type ServiceClientImpl struct {
	ssc session.SessionServiceClient
}

func NewSessionServiceClient(settings *settings.ArgoCDSettings) *ServiceClientImpl {
	conn := argocdServer.GetConnection("", settings)
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
