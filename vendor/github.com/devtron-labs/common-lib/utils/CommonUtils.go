/*
 * Copyright (c) 2024. Devtron Inc.
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

package utils

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/common-lib/git-manager/util"
	"github.com/devtron-labs/common-lib/utils/bean"
	"github.com/go-pg/pg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"math/rand"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

var chars = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

const (
	DOCKER_REGISTRY_TYPE_DOCKERHUB        = "docker-hub"
	DEVTRON_SELF_POD_UID                  = "DEVTRON_SELF_POD_UID"
	DEVTRON_SELF_POD_NAME                 = "DEVTRON_SELF_POD_NAME"
	DEVTRON_SELF_DOWNWARD_API_VOLUME      = "devtron-pod-info"
	DEVTRON_SELF_DOWNWARD_API_VOLUME_PATH = "/etc/devtron-pod-info"
	POD_LABELS                            = "labels"
	POD_ANNOTATIONS                       = "annotations"
)

// Generates random string
func Generate(size int) string {
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	for i := 0; i < size; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String()
	return str
}

func hasScheme(url string) bool {
	return len(url) >= 7 && (url[:7] == "http://" || url[:8] == "https://")
}

func GetUrlWithScheme(url string) (urlWithScheme string) {
	urlWithScheme = url
	if !hasScheme(url) {
		urlWithScheme = fmt.Sprintf("https://%s", url)
	}
	return urlWithScheme
}

func IsValidDockerTagName(tagName string) bool {
	regString := "^[a-zA-Z0-9_][a-zA-Z0-9_.-]{0,127}$"
	regexpCompile := regexp.MustCompile(regString)
	if regexpCompile.MatchString(tagName) {
		return true
	} else {
		return false
	}
}

func BuildDockerImagePath(dockerInfo bean.DockerRegistryInfo) (string, error) {
	dest := ""
	if DOCKER_REGISTRY_TYPE_DOCKERHUB == dockerInfo.DockerRegistryType {
		dest = dockerInfo.DockerRepository + ":" + dockerInfo.DockerImageTag
	} else {
		registryUrl := dockerInfo.DockerRegistryURL
		u, err := util.ParseUrl(registryUrl)
		if err != nil {
			log.Println("not a valid docker repository url")
			return "", err
		}
		u.Path = path.Join(u.Path, "/", dockerInfo.DockerRepository)
		dockerRegistryURL := u.Host + u.Path
		dest = dockerRegistryURL + ":" + dockerInfo.DockerImageTag
	}
	return dest, nil
}

func GetPGPostQueryProcessor(cfg bean.PgQueryMonitoringConfig) func(event *pg.QueryProcessedEvent) {
	return func(event *pg.QueryProcessedEvent) {
		query, err := event.FormattedQuery()
		if err != nil {
			log.Println("Error formatting query", "err", err)
			return
		}
		ExecutePGQueryProcessor(cfg, bean.PgQueryEvent{
			StartTime: event.StartTime,
			Error:     event.Error,
			Query:     query,
		})
	}
}

func ExecutePGQueryProcessor(cfg bean.PgQueryMonitoringConfig, event bean.PgQueryEvent) {
	queryDuration := time.Since(event.StartTime)
	var queryError bool
	pgError := event.Error
	if pgError != nil && !errors.Is(pgError, pg.ErrNoRows) {
		queryError = true
	}
	// Expose prom metrics
	if cfg.ExportPromMetrics {
		var status string
		if queryError {
			status = "FAIL"
		} else {
			status = "SUCCESS"
		}
		PgQueryDuration.WithLabelValues(status, cfg.ServiceName).Observe(queryDuration.Seconds())
	}

	// Log pg query if enabled
	logThresholdQueries := cfg.LogSlowQuery && queryDuration.Milliseconds() > cfg.QueryDurationThreshold
	logFailureQuery := queryError && cfg.LogAllFailureQueries
	if logFailureQuery {
		log.Println("PG_QUERY_FAIL - query time", "duration", queryDuration.Seconds(), "query", event.Query, "pgError", pgError)
	}
	if logThresholdQueries {
		log.Println("PG_QUERY_SLOW - query time", "duration", queryDuration.Seconds(), "query", event.Query)
	}
	if cfg.LogAllQuery {
		log.Println("query time", "duration", queryDuration.Seconds(), "query", event.Query)
	}
}

func GetSelfK8sUID() string {
	return os.Getenv(DEVTRON_SELF_POD_UID)
}

func GetSelfK8sPodName() string {
	return os.Getenv(DEVTRON_SELF_POD_NAME)
}

var PgQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "pg_query_duration_seconds",
	Help: "Duration of PG queries",
}, []string{"status", "serviceName"})

func ConvertTargetPlatformStringToObject(targetPlatformString string) []*bean.TargetPlatform {
	targetPlatforms := ConvertTargetPlatformStringToList(targetPlatformString)
	targetPlatformObject := []*bean.TargetPlatform{}
	for _, targetPlatform := range targetPlatforms {
		if len(targetPlatform) > 0 {
			targetPlatformObject = append(targetPlatformObject, &bean.TargetPlatform{Name: targetPlatform})
		}
	}
	return targetPlatformObject
}

func ConvertTargetPlatformStringToList(targetPlatform string) []string {
	return strings.Split(targetPlatform, ",")
}

func ConvertTargetPlatformListToString(targetPlatforms []string) string {
	return strings.Join(targetPlatforms, ",")
}
