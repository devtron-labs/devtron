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

package argocdServer

import (
	"fmt"
	"github.com/argoproj/argo-cd/util/settings"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"log"
)

func init() {

	grpc_prometheus.EnableClientHandlingTimeHistogram()
}

func GetConnection(token string, settings *settings.ArgoCDSettings) *grpc.ClientConn {
	conf, err := GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	var option []grpc.DialOption
	option = append(option, grpc.WithTransportCredentials(GetTLS(settings.Certificate)))
	if len(token) > 0 {
		option = append(option, grpc.WithPerRPCCredentials(TokenAuth{token: token}))
	}
	option = append(option, grpc.WithUnaryInterceptor(grpc_prometheus.UnaryClientInterceptor))
	option = append(option, grpc.WithStreamInterceptor(grpc_prometheus.StreamClientInterceptor))

	//if conf.Environment=="DEV"{
	//	option=append(option,grpc.WithInsecure())
	//}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", conf.Host, conf.Port), option...)
	if err != nil {
		log.Fatal(err)
	}
	return conn
}
