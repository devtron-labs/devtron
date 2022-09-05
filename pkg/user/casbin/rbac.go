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

package casbin

import (
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env"
	"github.com/casbin/casbin"
	"github.com/devtron-labs/authenticator/jwt"
	"github.com/devtron-labs/authenticator/middleware"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Enforcer interface {
	Enforce(emailId string, resource string, action string, resourceItem string) bool
	EnforceErr(emailId string, resource string, action string, resourceItem string) error
	EnforceByEmail(emailId string, resource string, action string, resourceItem string) bool
	EnforceByEmailInBatch(emailId string, resource string, action string, vals []string) map[string]bool
	InvalidateCache(emailId string) bool
	InvalidateCompleteCache()
	ReloadPolicy() error
	GetCacheDump() string
}

func NewEnforcerImpl(
	enforcer *casbin.SyncedEnforcer,
	sessionManager *middleware.SessionManager,
	logger *zap.SugaredLogger) *EnforcerImpl {
	lock := make(map[string]*CacheData)
	batchRequestLock := make(map[string]*sync.Mutex)
	enforcerConfig := getConfig()
	enf := &EnforcerImpl{lockCacheData: lock, enforcerRWLock: &sync.RWMutex{}, batchRequestLock: batchRequestLock, enforcerConfig: enforcerConfig,
		Cache: getEnforcerCache(logger, enforcerConfig), SyncedEnforcer: enforcer, logger: logger, SessionManager: sessionManager}
	setEnforcerImpl(enf)
	return enf
}

type CacheData struct {
	lock              *sync.RWMutex
	cacheCleaningFlag bool
	enforceReqCounter int64
}

type EnforcerConfig struct {
	CacheEnabled          bool `env:"ENFORCER_CACHE" envDefault:"false"`
	CacheExpirationInSecs int  `env:"ENFORCER_CACHE_EXPIRATION_IN_SEC" envDefault:"86400"`
	EnforcerBatchSize     int  `env:"ENFORCER_MAX_BATCH_SIZE" envDefault:"1"`
}

func getConfig() *EnforcerConfig {
	cacheConfig := &EnforcerConfig{}
	err := env.Parse(cacheConfig)
	if err != nil {
		log.Fatal(err)
	}
	return cacheConfig
}

func getEnforcerCache(logger *zap.SugaredLogger, enforcerConfig *EnforcerConfig) *cache.Cache {
	enableEnforcerCacheVal := enforcerConfig.CacheEnabled
	if enableEnforcerCacheVal {
		enforcerCacheExpirationDuration := time.Second * time.Duration(enforcerConfig.CacheExpirationInSecs)
		logger.Infow("enforce cache enabled", "expiry", enforcerCacheExpirationDuration)
		return cache.New(enforcerCacheExpirationDuration, 5*time.Minute)
	}
	return nil
}

// Enforcer is a wrapper around an Casbin enforcer that:
// * is backed by a kubernetes config map
// * has a predefined RBAC model
// * supports a built-in policy
// * supports a user-defined bolicy
// * supports a custom JWT claims enforce function
type EnforcerImpl struct {
	lockCacheData    map[string]*CacheData
	batchRequestLock map[string]*sync.Mutex
	*cache.Cache
	*casbin.SyncedEnforcer
	*middleware.SessionManager
	logger         *zap.SugaredLogger
	enforcerConfig *EnforcerConfig
	enforcerRWLock *sync.RWMutex
}

// Enforce is a wrapper around casbin.Enforce to additionally enforce a default role and a custom
// claims function
func (e *EnforcerImpl) Enforce(token string, resource string, action string, resourceItem string) bool {
	return e.enforce(token, resource, action, resourceItem)
}

func (e *EnforcerImpl) EnforceByEmail(emailId string, resource string, action string, resourceItem string) bool {
	allowed := e.enforceByEmail(emailId, resource, action, resourceItem)
	return allowed
}

func (e *EnforcerImpl) ReloadPolicy() error {
	//e.enforcerRWLock.Lock()
	//defer e.enforcerRWLock.Unlock()
	return e.SyncedEnforcer.LoadPolicy()
}

// EnforceErr is a convenience helper to wrap a failed enforcement with a detailed error about the request
func (e *EnforcerImpl) EnforceErr(emailId string, resource string, action string, resourceItem string) error {
	if !e.Enforce(emailId, resource, action, resourceItem) {
		errMsg := "permission denied"
		rvalsStrs := []string{resource, action, resourceItem}
		errMsg = fmt.Sprintf("%s: %s", errMsg, strings.Join(rvalsStrs, ", "))
		return status.Error(codes.PermissionDenied, errMsg)
	}
	return nil
}

func (e *EnforcerImpl) enforceByEmailInBatchSync(wg *sync.WaitGroup, mutex *sync.RWMutex, result map[string]bool, metrics map[int]int64, index int, emailId string, resource string, action string, vals []string) {
	defer wg.Done()
	start := time.Now()
	batchResult := make(map[string]bool)
	for _, resourceItem := range vals {
		data, err := e.enforcerEnforce(strings.ToLower(emailId), resource, action, resourceItem)
		if err == nil {
			batchResult[resourceItem] = data
		}
	}
	duration := time.Since(start)
	mutex.Lock()
	defer mutex.Unlock()
	for k, v := range batchResult {
		result[k] = v
	}
	metrics[index] = duration.Milliseconds()

}

func (e *EnforcerImpl) EnforceByEmailInBatch(emailId string, resource string, action string, vals []string) map[string]bool {
	var totalTimeGap int64 = 0
	var maxTimegap int64 = 0
	var minTimegap int64 = math.MaxInt64
	var avgTimegap float64
	batchSize := e.enforcerConfig.EnforcerBatchSize
	batchRequestLock := e.getBatchRequestLock(emailId)
	batchRequestLock.Lock()
	defer batchRequestLock.Unlock()

	var metrics = make(map[int]int64)
	result, notFoundItemList := e.batchEnforceFromCache(emailId, resource, action, vals)
	if len(result) > 0 {
		vals = notFoundItemList
	}

	totalSize := len(vals)
	wg := new(sync.WaitGroup)
	var batchMutex = &sync.RWMutex{}
	if batchSize > totalSize {
		batchSize = totalSize
	}
	wg.Add(batchSize)
	for i := 0; i < batchSize; i++ {
		startIndex := i * totalSize / batchSize
		endIndex := startIndex + totalSize/batchSize
		if endIndex > totalSize {
			endIndex = totalSize
		}
		go e.enforceByEmailInBatchSync(wg, batchMutex, result, metrics, i, emailId, resource, action, vals[startIndex:endIndex])
	}
	wg.Wait()
	for _, duration := range metrics {
		totalTimeGap += duration
		if duration > maxTimegap {
			maxTimegap = duration
		}
		if duration < minTimegap {
			minTimegap = duration
		}
	}

	enforcerCacheData := e.getEnforcerCacheLock(emailId)
	enforcerCacheData.lock.Lock()
	defer enforcerCacheData.lock.Unlock()
	returnVal := atomic.AddInt64(&enforcerCacheData.enforceReqCounter, -1)
	dataCached := false
	if enforcerCacheData.cacheCleaningFlag {
		if returnVal == 0 {
			enforcerCacheData.cacheCleaningFlag = false
		}
	} else {
		e.storeCacheData(emailId, resource, action, result)
		dataCached = true
	}

	if batchSize > 0 {
		avgTimegap = float64(totalTimeGap / int64(batchSize))
	}
	e.logger.Infow("enforce request for batch with data", "emailId", emailId, "resource", resource,
		"action", action, "totalElapsedTime", totalTimeGap, "maxTimegap", maxTimegap, "minTimegap",
		minTimegap, "avgTimegap", avgTimegap, "size", len(vals), "batchSize", batchSize, "cached", e.Cache != nil && dataCached)

	return result
}

func (e *EnforcerImpl) getBatchRequestLock(emailId string) *sync.Mutex {
	emailBatchRequestMutex, found := e.batchRequestLock[getLockKey(emailId)]
	if !found {
		emailBatchRequestMutex = &sync.Mutex{}
		e.batchRequestLock[getLockKey(emailId)] = emailBatchRequestMutex
	}
	return emailBatchRequestMutex
}

func (e *EnforcerImpl) getEnforcerCacheLock(emailId string) *CacheData {
	enforcerCacheMutex, found := e.lockCacheData[getLockKey(emailId)]
	if !found {
		enforcerCacheMutex =
			&CacheData{
				lock:              &sync.RWMutex{},
				enforceReqCounter: int64(0),
				cacheCleaningFlag: false,
			}
		e.lockCacheData[getLockKey(emailId)] = enforcerCacheMutex
	}
	return enforcerCacheMutex
}

func (e *EnforcerImpl) getCacheData(emailId string, resource string, action string) map[string]bool {
	result := make(map[string]bool)
	if e.Cache == nil {
		return result
	}
	emailResult, found := e.Cache.Get(emailId)
	if found {
		emailResultMap := emailResult.(map[string]map[string]bool)
		result = emailResultMap[getCacheKey(resource, action)]
		if result == nil {
			result = make(map[string]bool)
		}
	}
	return result
}

func (e *EnforcerImpl) batchEnforceFromCache(emailId string, resource string, action string, resourceItems []string) (map[string]bool, []string) {
	var result = make(map[string]bool)
	var notFoundDataList []string
	cacheLock := e.getEnforcerCacheLock(emailId)
	cacheLock.lock.RLock()
	defer freeCacheReadLock(cacheLock)
	enforceData := e.getCacheData(emailId, resource, action)
	if enforceData != nil {
		for _, resourceItem := range resourceItems {
			data, found := enforceData[resourceItem]
			if found {
				result[resourceItem] = data
			} else {
				notFoundDataList = append(notFoundDataList, resourceItem)
			}
		}
	}

	return result, notFoundDataList
}

func (e *EnforcerImpl) enforceFromCache(emailId string, resource string, action string, resourceItem string) (bool, bool) {
	cacheLock := e.getEnforcerCacheLock(emailId)
	cacheLock.lock.RLock()
	defer freeCacheReadLock(cacheLock)
	enforceData := e.getCacheData(emailId, resource, action)
	data, found := enforceData[resourceItem]
	return data, found
}

func freeCacheReadLock(cacheData *CacheData) {
	atomic.AddInt64(&cacheData.enforceReqCounter, 1)
	cacheData.lock.RUnlock()
}

func (e *EnforcerImpl) storeCacheData(emailId string, resource string, action string, result map[string]bool) {
	if e.Cache == nil {
		return
	}
	emailResult, found := e.Cache.Get(emailId)
	if !found {
		emailResult = make(map[string]map[string]bool)
	}
	emailResult.(map[string]map[string]bool)[getCacheKey(resource, action)] = result
	e.Cache.Set(emailId, emailResult, cache.DefaultExpiration)
}

func getCacheKey(resource string, action string) string {
	return resource + "$$" + action
}

func getLockKey(emailId string) string {
	return emailId
}

func (e *EnforcerImpl) InvalidateCache(emailId string) bool {
	cacheLock := e.getEnforcerCacheLock(emailId)
	cacheLock.lock.Lock()
	defer cacheLock.lock.Unlock()
	e.logger.Debugw("invalidating cache & setting flag ", "emailId", emailId)
	cacheLock.cacheCleaningFlag = true
	if e.Cache != nil {
		e.Cache.Delete(emailId)
		return true
	}
	return false
}

func (e *EnforcerImpl) InvalidateCompleteCache() {
	for emailId, _ := range e.lockCacheData {
		e.InvalidateCache(emailId)
	}
	if e.Cache != nil {
		e.Cache.Flush()
	}
}

func (e *EnforcerImpl) GetCacheDump() string {
	if e.Cache == nil {
		return "not-enabled"
	}
	items := e.Cache.Items()
	cacheData, err := json.Marshal(items)
	if err != nil {
		e.logger.Infow("error occurred while taking cache dump", "reason", err)
		return ""
	}
	return string(cacheData)
}

// enforce is a helper to additionally check a default role and invoke a custom claims enforcement function
func (e *EnforcerImpl) enforce(token string, resource string, action string, resourceItem string) bool {
	// check the default role
	email, invalid := e.verifyTokenAndGetEmail(token)
	if invalid {
		return false
	}
	return e.enforceByEmail(strings.ToLower(email), resource, action, resourceItem)
}

func (e *EnforcerImpl) enforceAndUpdateCache(email string, resource string, action string, resourceItem string) bool {
	cacheData := e.getEnforcerCacheLock(email)
	cacheData.lock.Lock()
	defer cacheData.lock.Unlock()
	enforcedStatus, err := e.enforcerEnforce(email, resource, action, resourceItem)
	returnVal := atomic.AddInt64(&cacheData.enforceReqCounter, -1)
	if cacheData.cacheCleaningFlag || err != nil {
		if returnVal == 0 {
			cacheData.cacheCleaningFlag = false
		}
		e.logger.Debugw("not updating enforcer status for cache", "email", email, "resource", resource,
			"action", action, "resourceItem", resourceItem, "enforceReqCounter", cacheData.enforceReqCounter, "err", err == nil)
		return enforcedStatus
	}
	enforceData := e.getCacheData(email, resource, action)
	enforceData[resourceItem] = enforcedStatus
	e.storeCacheData(email, resource, action, enforceData)
	return enforcedStatus
}

func (e *EnforcerImpl) enforcerEnforce(email string, resource string, action string, resourceItem string) (bool, error) {
	//e.enforcerRWLock.RLock()
	//defer e.enforcerRWLock.RUnlock()
	response, err := e.SyncedEnforcer.EnforceSafe(email, resource, action, resourceItem)
	if err != nil {
		e.logger.Errorw("error occurred while enforcing safe", "email", email,
			"resource", resource, "action", action, "resourceItem", resourceItem, "reason", err)
	}
	return response, err
}

func (e *EnforcerImpl) verifyTokenAndGetEmail(tokenString string) (string, bool) {
	claims, err := e.SessionManager.VerifyToken(tokenString)
	if err != nil {
		return "", true
	}
	mapClaims, err := jwt.MapClaims(claims)
	if err != nil {
		return "", true
	}
	email := jwt.GetField(mapClaims, "email")
	sub := jwt.GetField(mapClaims, "sub")
	if email == "" && (sub == "admin" || sub == "admin:login") {
		email = "admin"
	}
	return email, false
}

// enforce is a helper to additionally check a default role and invoke a custom claims enforcement function
func (e *EnforcerImpl) enforceByEmail(emailId string, resource string, action string, resourceItem string) bool {
	defer handlePanic()
	response, found := e.enforceFromCache(emailId, resource, action, resourceItem)
	if found {
		cacheData := e.getEnforcerCacheLock(emailId)
		cacheData.lock.Lock()
		defer cacheData.lock.Unlock()
		returnVal := atomic.AddInt64(&cacheData.enforceReqCounter, -1)
		if returnVal == 0 && cacheData.cacheCleaningFlag {
			cacheData.cacheCleaningFlag = false
		}
		return response
	}
	enforcedStatus := e.enforceAndUpdateCache(emailId, resource, action, resourceItem)
	return enforcedStatus
}

// MatchKeyByPartFunc is the wrapper of our own customised MatchKeyByPart Func
func MatchKeyByPartFunc(args ...interface{}) (interface{}, error) {
	name1 := args[0].(string)
	name2 := args[1].(string)

	return bool(MatchKeyByPart(name1, name2)), nil
}

// MatchKeyByPart checks whether values in key1 matches all values of key2(values are obtained by splitting key by "/")
// For example - key1 =  "a/b/c" matches key2 = "a/*/c" but not matches for key2 = "a/*/d"
func MatchKeyByPart(key1 string, key2 string) bool {

	if key2 == "*" {
		//policy must be for super-admin role or global-env action
		//no need to check further
		return true
	}

	key1Vals := strings.Split(key1, "/")
	key2Vals := strings.Split(key2, "/")

	if (len(key1Vals) != len(key2Vals)) || len(key1Vals) == 0 {
		//values in keys should be more than zero and must be equal
		return false
	}

	for i, key2Val := range key2Vals {
		key1Val := key1Vals[i]

		if key2Val == "" || key1Val == "" {
			//empty values are not allowed in any key
			return false
		} else {
			// getting index of "*" in key2, will check values of key1 accordingly
			//for example - key2Val = a/bc*/d & key1Val = a/bcd/d, in this case "bc" will be checked in key1Val(upto index of "*")
			j := strings.Index(key2Val, "*")
			if j == -1 {
				if key1Val != key2Val {
					return false
				}
			} else if len(key1Val) > j {
				if key1Val[:j] != key2Val[:j] {
					return false
				}
			} else {
				if key1Val != key2Val[:j] {
					return false
				}
			}
		}
	}
	return true
}
