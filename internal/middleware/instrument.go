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

package middleware

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HttpLabels struct {
	UrlPaths string `env:"URL_PATHS" envDefault:"[]"`
}

var UrlLabelsMapping map[string]map[string]string
var UniqueKeys []string

const (
	path   string = "path"
	method string = "method"
	status string = "status"
)

func GetHttpLabels() (*HttpLabels, error) {
	cfg := &HttpLabels{}
	err := env.Parse(cfg)
	return cfg, err
}

type UrlPaths struct {
	Url   []string          `json:"url"`
	Label map[string]string `json:"label"`
}

func getLabels() []string {
	httpLabels, err := GetHttpLabels()
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var data []UrlPaths
	if err := json.Unmarshal([]byte(httpLabels.UrlPaths), &data); err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	urlMappings := make(map[string]map[string]string)
	keys := make(map[string]bool)

	for _, obj := range data {
		for _, urlStr := range obj.Url {
			urlStr = strings.TrimSpace(urlStr)
			if existingLabels, exists := urlMappings[urlStr]; exists {
				for key, value := range obj.Label {
					strValue := strings.TrimSpace(value)
					existingLabels[key] = strValue
				}
			} else {
				labels := make(map[string]string)
				for key, value := range obj.Label {
					strValue := strings.TrimSpace(value)
					labels[key] = strValue
				}
				urlMappings[urlStr] = labels
			}
		}
		for key := range obj.Label {
			keys[key] = true
		}
	}

	UrlLabelsMapping = urlMappings
	uniqueKeys := []string{path, method, status}
	for key := range keys {
		uniqueKeys = append(uniqueKeys, key)
	}
	UniqueKeys = uniqueKeys
	return uniqueKeys
}

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "orchestrator_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, getLabels())
)

var PgQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "pg_query_duration_seconds",
	Help: "Duration of PG queries",
}, []string{"label"})

var CdDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "cd_duration_seconds",
	Help: "Duration of CD process",
}, []string{"appName", "status", "envName", "deploymentType"})

var GitOpsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "git_ops_duration_seconds",
	Help: "Duration of GitOps",
}, []string{"operationName", "methodName", "status"})

var CiDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "ci_duration_seconds",
	Help:    "Duration of CI process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var CacheDownloadDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "cache_download_duration_seconds",
	Help:    "Duration of Cache Download process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var PreCiDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "pre_ci_duration_seconds",
	Help:    "Duration of Pre CI process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var BuildDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "build_duration_seconds",
	Help:    "Duration of Build process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var PostCiDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "post_ci_duration_seconds",
	Help:    "Duration of Post CI process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var CacheUploadDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "cache_upload_duration_seconds",
	Help:    "Duration of Cache Upload process",
	Buckets: prometheus.LinearBuckets(20, 20, 5),
}, []string{"pipelineName", "appName"})

var AppListingDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "app_listing_duration_seconds",
	Help: "Duration of App Listing process",
}, []string{"MethodName", "AppType"})

var requestCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "orchestrator_http_requests_total",
		Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
	},
	[]string{path, method, status})

var currentRequestGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "orchestrator_http_requests_current",
	Help: "no of request being served currently",
}, []string{path, method})

var AcdGetResourceCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "acd_get_resource_counter",
}, []string{"appId", "envId", "acdAppName"})

var CdTriggerCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "cd_trigger_counter",
}, []string{"appName", "envName"})

var CiTriggerCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ci_trigger_counter",
}, []string{"appName", "pipelineName"})

var DeploymentStatusCronDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "deployment_status_cron_process_time",
}, []string{"cronName"})

var TerminalSessionRequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "initiate_terminal_session_request_counter",
	Help: "count of requests for initiated, established and closed terminal sessions",
}, []string{"sessionAction", "isError"})

var TerminalSessionDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "terminal_session_duration",
	Help: "duration of each terminal session",
}, []string{"podName", "namespace", "clusterId"})

func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		urlPath, _ := route.GetPathTemplate()
		urlMethod := r.Method

		g := currentRequestGauge.WithLabelValues(urlPath, urlMethod)
		g.Inc()
		defer g.Dec()

		d := NewDelegator(w, nil)
		next.ServeHTTP(d, r)

		// Construct label array for Prometheus
		valuesArray := []string{urlPath, urlMethod, strconv.Itoa(d.Status())}
		if key, ok := UrlLabelsMapping[urlPath]; !ok {
			for _, labelKey := range UniqueKeys {
				if areDefaultKeys(labelKey) {
					continue
				}
				valuesArray = append(valuesArray, "")
			}
		} else {
			for _, labelKey := range UniqueKeys {
				if labelValue, ok1 := key[labelKey]; !ok1 {
					if areDefaultKeys(labelKey) {
						continue
					}
					valuesArray = append(valuesArray, "")
				} else {
					valuesArray = append(valuesArray, labelValue)
				}
			}
		}

		httpDuration.WithLabelValues(valuesArray...).Observe(time.Since(start).Seconds())
		requestCounter.WithLabelValues(urlPath, urlMethod, strconv.Itoa(d.Status())).Inc()
	})
}

func areDefaultKeys(labelKey string) bool {
	return labelKey == path || labelKey == status || labelKey == method
}

func IncTerminalSessionRequestCounter(sessionAction string, isError string) {
	TerminalSessionRequestCounter.WithLabelValues(sessionAction, isError).Inc()
}

func RecordTerminalSessionDurationMetrics(podName, namespace, clusterId string, sessionDuration float64) {
	TerminalSessionDuration.WithLabelValues(podName, namespace, clusterId).Observe(sessionDuration)
}
