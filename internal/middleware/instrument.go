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

// urlpattern
// [{"url":["url1,url2"], "label":{k1:v1,k2:v2}},
// {"url":["url4,url5"], "label":{k5:v6,k1:v8}}
// ]

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

/*
getLabels processes URL and label information, storing mappings and unique keys.

Example:
Suppose the `GetHttpLabels` function returns the following data:

	{
	    "UrlPaths": `[
	        {
	            "url": ["http://example.com", "http://example.org"],
	            "label": {"name": "Example", "type": "Website", "key1": "value1"}
	        },
	        {
	            "url": ["http://example.com, http://example.net"],
	            "label": {"name": "Example2", "type": "Site"}
	        }
	    ]`
	}

Output:
Unique Keys:
[

	"path",
	"method",
	"status",
	"name",
	"type",
	"key1"

]
URL Labels Mapping:

	{
	    "http://example.com": {"name": "Example2", "type": "Site", "key1": "value1"},
	    "http://example.org": {"name": "Example", "type": "Website", "key1": "value1"},
	    "http://example.net": {"name": "Example2", "type": "Site"}
	}

Explanation:
1. The function calls `GetHttpLabels` to retrieve JSON data containing URL and label information.
2. It unmarshals the JSON data into a slice of maps.
3. It initializes maps to store URL to label mappings and unique label keys.
4. It iterates over the data to process URLs and labels, merging them into `urlMappings`.
5. It also collects unique keys while processing labels.
6. The global variables `UrlLabelsMapping` and `UniqueKeys` are updated with the mappings and unique keys, respectively.
*/
func getLabels() []string {
	httpLabels, err := GetHttpLabels()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var data []map[string]interface{}
	// Unmarshal JSON into the defined struct
	if err := json.Unmarshal([]byte(httpLabels.UrlPaths), &data); err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	// Define a map to store the URL to labels mappings
	urlMappings := make(map[string]map[string]string)
	// Define a map to store unique keys (labels)
	keys := make(map[string]bool)

	// Iterate through each object in the array
	for _, obj := range data {
		// Get the "url" array from each object
		urls, ok := obj["url"].([]interface{})
		if !ok {
			continue // Skip if "url" is not an array
		}

		// Get the "label" object from each object
		labelObj, ok := obj["label"].(map[string]interface{})
		if !ok {
			continue // Skip if "label" is not a map
		}

		// Iterate through each URL and add the mappings
		for _, url := range urls {
			urlStr, ok := url.(string)
			if !ok {
				continue // Skip if URL is not a string
			}

			// Handle URLs separated by commas and trim whitespace
			urlArray := strings.Split(urlStr, ",")
			for _, sUrl := range urlArray {
				sUrl = strings.TrimSpace(sUrl)

				// If the URL is already in the map, merge the labels
				if existingLabels, exists := urlMappings[sUrl]; exists {
					for key, value := range labelObj {
						strValue, ok := value.(string)
						if !ok {
							continue // Skip if label value is not a string
						}
						strValue = strings.TrimSpace(strValue)
						existingLabels[key] = strValue
					}
				} else {
					// Create a new map for each URL to store its labels
					labels := make(map[string]string)
					for key, value := range labelObj {
						strValue, ok := value.(string)
						if !ok {
							continue // Skip if label value is not a string
						}
						strValue = strings.TrimSpace(strValue)
						labels[key] = strValue
					}
					urlMappings[sUrl] = labels
				}
			}
		}

		// Add keys to the unique keys map
		for key := range labelObj {
			// Split the key by ":" to get the key part
			parts := strings.Split(key, ":")
			if len(parts) > 0 {
				keys[parts[0]] = true
			}
		}
	}

	UrlLabelsMapping = urlMappings

	// Extract the unique keys
	var uniqueKeys []string
	uniqueKeys = append(uniqueKeys, path, method, status)
	for key := range keys {
		key = strings.TrimSpace(key)
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

/*
PrometheusMiddleware
Example:
Suppose the incoming HTTP request has the following details:
- Path: "/api/data"
- Method: "GET"

And the `UrlLabelsMapping` contains:

	{
	    "/api/data": {"name": "DataEndpoint", "type": "API"}
	}

Output:
The Prometheus metrics for the request would be recorded with the following labels:
- path: "/api/data"
- method: "GET"
- status: "200"
- name: "DataEndpoint"
- type: "API"

Explanation:
1. The middleware function PrometheusMiddleware takes an `http.Handler` and returns a new handler that includes Prometheus metrics.
2. Inside the handler function:
  - It captures the start time of the request.
  - It retrieves the current route and path template using `mux.CurrentRoute`.
  - It initializes a gauge `g` to track the current number of requests.
  - It uses `NewDelegator` to wrap the response writer, allowing the status code to be captured.
  - It serves the request using the next handler in the chain.
  - It constructs a `strArr` slice to hold label values for Prometheus metrics.
  - It checks if the path exists in `UrlLabelsMapping`:
  - If it does, it appends the label values from the mapping.
  - If it doesn't, it appends empty strings for missing labels.
  - It records the duration of the request using `httpDuration` metric.
  - It increments the request counter `requestCounter` with the path, method, and status code.

Metrics:
- `httpDuration`: Observes the duration of the HTTP request.
- `requestCounter`: Counts the number of requests processed.

Note: The global variables `UrlLabelsMapping` and `UniqueKeys` are used for label extraction.
*/
func PrometheusMiddleware(next http.Handler) http.Handler {
	// prometheus.MustRegister(requestCounter)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		route := mux.CurrentRoute(r)
		urlPath, _ := route.GetPathTemplate()
		urlMethod := r.Method

		// Gauge to track the current number of requests
		g := currentRequestGauge.WithLabelValues(urlPath, urlMethod)
		g.Inc()
		defer g.Dec()

		// Delegator to capture the status code
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

		// Record the duration and increment the request counter
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
