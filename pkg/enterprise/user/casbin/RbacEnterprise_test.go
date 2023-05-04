package casbin

import (
	"encoding/csv"
	"fmt"
	"github.com/casbin/casbin"
	"github.com/casbin/casbin/model"
	"log"
	"os"
	"strings"
	"testing"
)

type req struct {
	subject  string
	resource string
	action   string
	objects  []string
}

var multiplyingFactor = 100000

func getRequests() []req {
	reqs := make([]req, 0, 40)
	for i := 1; i <= 10; i++ {
		objs := make([]string, 0, multiplyingFactor)
		for j := 0; j < i*multiplyingFactor; j++ {
			objs = append(objs, fmt.Sprintf("proj-%[1]d-%[2]d", i, j))
		}
		reqs = append(reqs, req{
			subject:  "kartik@devtron.ai",
			resource: "team",
			action:   "get",
			objects:  objs,
		})
	}
	for i := 1; i <= 10; i++ {
		objs := make([]string, 0, multiplyingFactor)
		for j := 0; j < i*multiplyingFactor; j++ {
			objs = append(objs, fmt.Sprintf("proj-%[1]d-%[2]d/app-%[1]d-%[2]d", i, j))
		}
		reqs = append(reqs, req{
			subject:  "kartik@devtron.ai",
			resource: "applications",
			action:   "create",
			objects:  objs,
		})
	}
	for i := 1; i <= 10; i++ {
		objs := make([]string, 0, multiplyingFactor)
		for j := 0; j < i*multiplyingFactor; j++ {
			objs = append(objs, fmt.Sprintf("proj-%[1]d-%[2]d/envapp-%[1]d-%[2]d/app-%[1]d-%[2]d", i, j))
		}
		reqs = append(reqs, req{
			subject:  "kartik@devtron.ai",
			resource: "helm-app",
			action:   "update",
			objects:  objs,
		})
	}
	for i := 1; i <= 10; i++ {
		objs := make([]string, 0, multiplyingFactor)
		for j := 0; j < i*multiplyingFactor; j++ {
			objs = append(objs, fmt.Sprintf("envapp-%[1]d-%[2]d/app-%[1]d-%[2]d", i, j))
		}
		reqs = append(reqs, req{
			subject:  "kartik@devtron.ai",
			resource: "environment",
			action:   "delete",
			objects:  objs,
		})
	}
	return reqs
}

func getModel() model.Model {
	file, err := os.Open("rbac_enterprise_benchmark_policies.csv")
	if err != nil {
		log.Print(err)
	}
	defer file.Close()
	m := casbin.NewModel("../../../../auth_model.conf", "")
	csvReader := csv.NewReader(file)
	lines, err := csvReader.ReadAll()
	if err != nil {
		log.Print(err)
	}
	for _, lineValues := range lines {
		readPolicies(lineValues, m)
	}
	return m
}

func readPolicies(lineValues []string, model model.Model) {
	for i := 0; i < len(lineValues); i++ {
		lineValues[i] = strings.ReplaceAll(lineValues[i], " ", "")
	}
	//testing csv file had lines like - "p|role:view_devtron_test-devtron__devtroncd_app|team|get|devtron|allow|"
	pVals := strings.Split(lineValues[0], "|")
	key := pVals[0]
	sec := key[:1]
	model[sec][key].Policy = append(model[sec][key].Policy, pVals[1:])
}

func BenchmarkEnforceForSubjectInBatchCustom(b *testing.B) {
	entertpriseEnforcer := EnterpriseEnforcerImpl{}
	e := casbin.NewEnforcer("../../../../auth_model.conf", "rbac_enterprise_benchmark_policies.csv")
	reqs := getRequests()
	for _, v := range reqs {
		b.Run(fmt.Sprintf("input-res:%s, %d resouceItems", v.resource, len(v.objects)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				entertpriseEnforcer.EnforceForSubjectInBatchCustom(v.subject, v.resource, v.action, v.objects, e.GetModel())
			}
		})
	}
}

func BenchmarkEnforceForSubjectInBatchCasbin(b *testing.B) {
	entertpriseEnforcer := EnterpriseEnforcerImpl{}
	e := casbin.NewEnforcer("../../../../auth_model.conf", "rbac_enterprise_benchmark_policies.csv")
	reqs := getRequests()
	for _, v := range reqs {
		b.Run(fmt.Sprintf("input-res:%s, %d resouceItems", v.resource, len(v.objects)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				entertpriseEnforcer.EnforceForSubjectInBatchCasbin(v.subject, v.resource, v.action, v.objects, e.GetModel())
			}
		})
	}
}
