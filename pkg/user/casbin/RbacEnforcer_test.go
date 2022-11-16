package casbin

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"strconv"
	"testing"
	"time"
)

func TestRbacEnforcer(t *testing.T) {
	t.SkipNow()
	t.Run("check for RBAC ", func(t *testing.T) {
		syncedEnforcer := Create()
		logger, err := getLogger()
		assert.Nil(t, err)
		enforcerImpl := NewEnforcerImpl(syncedEnforcer, nil, logger)
		//enforced := enforcerImpl.EnforceByEmail("pawan@devtron.ai", "application", "get", "*")
		var rvalsArr [][]interface{}
		startTime := time.Now()
		for i := 0; i < 2000; i++ {
			//st1 := time.Now()
			//response := syncedEnforcer.Enforce("test@devtron.ai", "applications123213", "get", "*")
			//fmt.Println("response", response, "elapsed", time.Since(st1))
			appName := "core/test-app1" + strconv.Itoa(i)
			rval := getRval("test@devtron.ai", "applications", "get", appName)
			rvalsArr = append(rvalsArr, rval)
		}
		responseArr := enforcerImpl.BatchEnforceForSubject("test@devtron.ai", rvalsArr)
		elapsedTime := time.Since(startTime)
		for _, response := range responseArr {
			fmt.Println("enforced", response, "timegapInMillis", elapsedTime.Milliseconds())
		}
	})
}

func getLogger() (*zap.SugaredLogger, error) {
	config := zap.NewProductionConfig()
	l, err := config.Build()
	if err != nil {
		fmt.Println("failed to create the default logger: " + err.Error())
		return nil, err
	}
	return l.Sugar(), nil
}

func getRval(rval ...interface{}) []interface{} {
	return rval
}
