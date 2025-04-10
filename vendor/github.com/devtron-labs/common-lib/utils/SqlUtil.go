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
	"github.com/devtron-labs/common-lib/utils/bean"
	"github.com/go-pg/pg"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"io"
	"log"
	"net"
	"os"
	"time"
)

const (
	PgNetworkErrorLogPrefix string = "PG_NETWORK_ERROR"
	PgQueryFailLogPrefix    string = "PG_QUERY_FAIL"
	PgQuerySlowLogPrefix    string = "PG_QUERY_SLOW"
)

const (
	FAIL    string = "FAIL"
	SUCCESS string = "SUCCESS"
)

type ErrorType string

func (e ErrorType) String() string {
	return string(e)
}

const (
	NetworkErrorType ErrorType = "NETWORK_ERROR"
	SyntaxErrorType  ErrorType = "SYNTAX_ERROR"
	TimeoutErrorType ErrorType = "TIMEOUT_ERROR"
	NoErrorType      ErrorType = "NA"
)

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
			FuncName:  event.Func,
		})
	}
}

func ExecutePGQueryProcessor(cfg bean.PgQueryMonitoringConfig, event bean.PgQueryEvent) {
	queryDuration := time.Since(event.StartTime)
	var queryError bool
	pgError := event.Error
	if pgError != nil && !errors.Is(pgError, pg.ErrNoRows) && !isIntegrityViolationError(pgError) {
		queryError = true
	}
	// Expose prom metrics
	if cfg.ExportPromMetrics {
		var status string
		if queryError {
			status = FAIL
		} else {
			status = SUCCESS
		}
		PgQueryDuration.WithLabelValues(status, cfg.ServiceName, event.FuncName, getErrorType(pgError).String()).Observe(queryDuration.Seconds())
	}

	// Log pg query if enabled
	logThresholdQueries := cfg.LogSlowQuery && queryDuration.Milliseconds() > cfg.QueryDurationThreshold
	logNetworkFailure := queryError && cfg.LogAllFailureQueries && isNetworkError(pgError)
	if logNetworkFailure {
		log.Println(fmt.Sprintf("%s - query time", PgNetworkErrorLogPrefix), "duration", queryDuration.Seconds(), "query", event.Query, "pgError", pgError)
	}
	logFailureQuery := queryError && cfg.LogAllFailureQueries && !isNetworkError(pgError)
	if logFailureQuery {
		log.Println(fmt.Sprintf("%s - query time", PgQueryFailLogPrefix), "duration", queryDuration.Seconds(), "query", event.Query, "pgError", pgError)
	}
	if logThresholdQueries {
		log.Println(fmt.Sprintf("%s - query time", PgQuerySlowLogPrefix), "duration", queryDuration.Seconds(), "query", event.Query)
	}
	if cfg.LogAllQuery {
		log.Println("query time", "duration", queryDuration.Seconds(), "query", event.Query)
	}
}

var PgQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name: "pg_query_duration_seconds",
	Help: "Duration of PG queries",
}, []string{"status", "serviceName", "functionName", "errorType"})

func getErrorType(err error) ErrorType {
	if err == nil {
		return NoErrorType
	} else if errors.Is(err, os.ErrDeadlineExceeded) {
		return TimeoutErrorType
	} else if isNetworkError(err) {
		return NetworkErrorType
	}
	return SyntaxErrorType
}

func isNetworkError(err error) bool {
	if err == io.EOF {
		return true
	}
	_, ok := err.(net.Error)
	return ok
}

func isIntegrityViolationError(err error) bool {
	pgErr, ok := err.(pg.Error)
	if !ok {
		return false
	}
	return pgErr.IntegrityViolation()
}
