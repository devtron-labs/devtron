package casbin

import (
	"fmt"
	"github.com/devtron-labs/devtron/internal/util"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestRbacEnforcer(t *testing.T) {
	t.Run("check for RBAC ", func(t *testing.T) {
		syncedEnforcer := Create()
		logger, err := util.InitLogger()
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

func getRval(rval ...interface{}) []interface{} {
	return rval
}
