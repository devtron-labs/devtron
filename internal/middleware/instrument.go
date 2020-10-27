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
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"strconv"
	"time"
)

var (
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "orchestrator_http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"path", "method", "status"})
)

var requestCounter = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "orchestrator_http_requests_total",
		Help: "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
	},
	[]string{"path", "method", "status"})

var currentRequestGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "orchestrator_http_requests_current",
	Help: "no of request being served currently",
}, []string{"path", "method"})

var AcdGetResourceCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "acd_get_resource_counter",
}, []string{"appId", "envId", "acdAppName"})

var CdTriggerCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "cd_trigger_counter",
}, []string{"appId", "envId", "pipelineId"})

var CiTriggerCounter = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "ci_trigger_counter",
}, []string{"appId", "pipelineId"})

// prometheusMiddleware implements mux.MiddlewareFunc.
func PrometheusMiddleware(next http.Handler) http.Handler {
	//	prometheus.MustRegister(requestCounter)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		path, _ := route.GetPathTemplate()
		method := r.Method
		g := currentRequestGauge.WithLabelValues(path, method)
		g.Inc()
		defer g.Dec()
		d := newDelegator(w, nil)
		next.ServeHTTP(d, r)
		httpDuration.WithLabelValues(path, method, strconv.Itoa(d.Status())).Observe(time.Since(start).Seconds())
		requestCounter.WithLabelValues(path, method, strconv.Itoa(d.Status())).Inc()
	})
}
