/*
 * Copyright (c) 2020-2024. Devtron Inc.
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

package k8s

import (
	"github.com/caarlos0/env"
	"go.uber.org/zap"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/client-go/rest"
	"net"
	"net/http"
	"time"
)

type DefaultK8sHttpTransportConfig struct{}

type CustomK8sHttpTransportConfig struct {
	UseCustomTransport  bool `env:"USE_CUSTOM_HTTP_TRANSPORT" envDefault:"false"`
	TimeOut             int  `env:"K8s_TCP_TIMEOUT" envDefault:"30"`
	KeepAlive           int  `env:"K8s_TCP_KEEPALIVE" envDefault:"30"`
	TLSHandshakeTimeout int  `env:"K8s_TLS_HANDSHAKE_TIMEOUT" envDefault:"10"`
	MaxIdleConnsPerHost int  `env:"K8s_CLIENT_MAX_IDLE_CONNS_PER_HOST" envDefault:"25"`
	IdleConnTimeout     int  `env:"K8s_TCP_IDLE_CONN_TIMEOUT" envDefault:"300"`
}

func NewDefaultK8sHttpTransportConfig() *DefaultK8sHttpTransportConfig {
	return &DefaultK8sHttpTransportConfig{}
}

func NewCustomK8sHttpTransportConfig(logger *zap.SugaredLogger) *CustomK8sHttpTransportConfig {
	customK8sHttpTransportConfig := &CustomK8sHttpTransportConfig{}
	err := env.Parse(customK8sHttpTransportConfig)
	if err != nil {
		logger.Errorw("error in parsing custom k8s http configurations from env", "err", err)
	}
	return customK8sHttpTransportConfig
}

type TransportType string

const (
	TransportTypeDefault    TransportType = "default"
	TransportTypeOverridden TransportType = "overridden"
)

type HttpTransportConfig struct {
	customHttpClientConfig  HttpTransportInterface
	defaultHttpClientConfig HttpTransportInterface
}

func NewHttpTransportConfig(logger *zap.SugaredLogger) *HttpTransportConfig {
	return &HttpTransportConfig{
		customHttpClientConfig:  NewCustomK8sHttpTransportConfig(logger),
		defaultHttpClientConfig: NewDefaultK8sHttpTransportConfig(),
	}
}

type HttpTransportInterface interface {
	OverrideConfigWithCustomTransport(config *rest.Config) (*rest.Config, error)
}

// OverrideConfigWithCustomTransport
// sets returns the given rest config without any modifications even if UseCustomTransport is enabled.
// This is used when we want to use the default rest.Config provided by the client-go library.
func (impl *DefaultK8sHttpTransportConfig) OverrideConfigWithCustomTransport(config *rest.Config) (*rest.Config, error) {
	return config, nil
}

// OverrideConfigWithCustomTransport
// overrides the given rest config with custom transport if UseCustomTransport is enabled.
// if the config already has a defined transport, we don't override it.
func (impl *CustomK8sHttpTransportConfig) OverrideConfigWithCustomTransport(config *rest.Config) (*rest.Config, error) {
	if !impl.UseCustomTransport || config.Transport != nil {
		return config, nil
	}

	dial := (&net.Dialer{
		Timeout:   time.Duration(impl.TimeOut) * time.Second,
		KeepAlive: time.Duration(impl.KeepAlive) * time.Second,
	}).DialContext

	// Get the TLS options for this client config
	tlsConfig, err := rest.TLSConfigFor(config)
	if err != nil {
		return nil, err
	}

	transport := utilnet.SetTransportDefaults(&http.Transport{
		Proxy:               config.Proxy,
		TLSHandshakeTimeout: time.Duration(impl.TLSHandshakeTimeout) * time.Second,
		TLSClientConfig:     tlsConfig,
		MaxIdleConns:        impl.MaxIdleConnsPerHost,
		MaxConnsPerHost:     impl.MaxIdleConnsPerHost,
		MaxIdleConnsPerHost: impl.MaxIdleConnsPerHost,
		DialContext:         dial,
		DisableCompression:  config.DisableCompression,
		IdleConnTimeout:     time.Duration(impl.IdleConnTimeout) * time.Second,
	})

	rt, err := rest.HTTPWrappersForConfig(config, transport)
	if err != nil {
		return nil, err
	}

	config.Transport = rt
	config.Timeout = time.Duration(impl.TimeOut) * time.Second

	// set default tls config and remove auth/exec provides since we use it in a custom transport.
	// we already set tls config in the transport
	config.TLSClientConfig = rest.TLSClientConfig{}
	config.AuthProvider = nil
	config.ExecProvider = nil

	return config, nil
}
