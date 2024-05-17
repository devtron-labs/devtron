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

func GetHttpLabels() (*HttpLabels, error) {
	cfg := &HttpLabels{}
	err := env.Parse(cfg)
	return cfg, err
}

/*
Example:
Input data:
[

	{
	    "url": ["http://example.com", "http://example.org"],
	    "label": {"name": "Example", "type": "Website" , "key1": "value1"}
	},
	{
	    "url": ["http://example.com, http://example.net"],
	    "label": {"name": "Example2", "type": "Site"}
	}

]

Output:

	{
	    "http://example.com": {"name": "Example2", "type": "Site", "key1": "value1"},
	    "http://example.org": {"name": "Example", "type": "Website", "key1": "value1"},
	    "http://example.net": {"name": "Example2", "type": "Site"}
	}

Explanation:
1. The function initializes an empty map `urlMappings` to store URL to label mappings.
2. It iterates over each object in the input `data` array.
3. For each object, it retrieves the "url" array and "label" map. If either is not in the expected format, it skips that object.
4. It then iterates over each URL in the "url" array, handling cases where multiple URLs are separated by commas.
5. For each URL, it checks if it already exists in `urlMappings`. If it does, it updates the existing labels. If not, it creates a new entry for that URL.
6. Finally, it returns the `urlMappings` map, which contains each URL mapped to its corresponding labels.
*/
func getUrlLabelMapping(data []map[string]interface{}) map[string]map[string]string {
	// Define a map to store the URL to labels mappings
	urlMappings := make(map[string]map[string]string)

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
	}
	return urlMappings
}

/*
Example:
Suppose the `GetHttpLabels` function returns the following data:

	{
	    "UrlPaths": `[
	        {
	            "url": ["http://example.com", "http://example.org"],
	            "label": {"name": "Example", "type": "Website"}
	        },
	        {
	            "url": ["http://example.com, http://example.net"],
	            "label": {"name": "Example2", "type": "Site"}
	        }
	    ]`
	}

Output:
[

	"path",
	"method",
	"status",
	"name",
	"type"

]

Explanation:
1. The function calls `GetHttpLabels` to retrieve JSON data containing URL and label information.
2. It initializes an empty slice `data` to hold the unmarshaled JSON data.
3. It unmarshals the JSON data into the `data` slice of maps.
4. It calls `getUrlLabelMapping` to process the URL and label data, resulting in a `UrlLabelsMapping`.
5. It initializes a map `keys` to store unique label keys.
6. It iterates over each object in `data` to extract label keys, splitting keys by ":" and storing the first part.
7. It adds some predefined keys ("path", "method", "status") to the unique keys list.
8. It iterates through the `keys` map to add all unique keys to the `uniqueKeys` slice.
9. Finally, it returns the `uniqueKeys` slice.

Note: The global variables `UrlLabelsMapping` and `UniqueKeys` are updated within the function.
*/
func getLabels() []string {
	httpLabels, err := GetHttpLabels()
	if err != nil {
		fmt.Println(err)
	}

	var data []map[string]interface{}
	// Unmarshal JSON into the defined struct
	if err := json.Unmarshal([]byte(httpLabels.UrlPaths), &data); err != nil {
		fmt.Println("Error:", err)
	}

	UrlLabelsMapping = getUrlLabelMapping(data)

	// Define a map to store unique keys (labels)
	keys := make(map[string]bool)

	// Iterate through each object in the array
	for _, obj := range data {
		// Get the "label" object from each object
		labelObj, ok := obj["label"].(map[string]interface{})
		if !ok {
			continue
		}
		// Iterate through each key-value pair in the "label" object
		for key := range labelObj {
			// Split the key by ":" to get the key part
			parts := strings.Split(key, ":")
			if len(parts) > 0 {
				keys[parts[0]] = true
			}
		}
	}
	// Extract the unique keys
	var uniqueKeys []string
	uniqueKeys = append(uniqueKeys, "path")
	uniqueKeys = append(uniqueKeys, "method")
	uniqueKeys = append(uniqueKeys, "status")
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
		path, _ := route.GetPathTemplate()
		method := r.Method

		// Gauge to track the current number of requests
		g := currentRequestGauge.WithLabelValues(path, method)
		g.Inc()
		defer g.Dec()

		// Delegator to capture the status code
		d := NewDelegator(w, nil)
		next.ServeHTTP(d, r)

		// Construct label array for Prometheus
		strArr := []string{path, method, strconv.Itoa(d.Status())}
		if key, ok := UrlLabelsMapping[path]; !ok {
			for _, labelKey := range UniqueKeys {
				if labelKey == "path" || labelKey == "status" || labelKey == "method" {
					continue
				}
				strArr = append(strArr, "")
			}
		} else {
			for _, labelKey := range UniqueKeys {
				if labelValue, ok1 := key[labelKey]; !ok1 {
					if labelKey == "path" || labelKey == "status" || labelKey == "method" {
						continue
					}
					strArr = append(strArr, "")
				} else {
					strArr = append(strArr, labelValue)
				}
			}
		}

		// Record the duration and increment the request counter
		httpDuration.WithLabelValues(strArr...).Observe(time.Since(start).Seconds())
		requestCounter.WithLabelValues(path, method, strconv.Itoa(d.Status())).Inc()
	})
}

func IncTerminalSessionRequestCounter(sessionAction string, isError string) {
	TerminalSessionRequestCounter.WithLabelValues(sessionAction, isError).Inc()
}

func RecordTerminalSessionDurationMetrics(podName, namespace, clusterId string, sessionDuration float64) {
	TerminalSessionDuration.WithLabelValues(podName, namespace, clusterId).Observe(sessionDuration)
}
