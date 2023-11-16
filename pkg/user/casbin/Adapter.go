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
	"fmt"
	"github.com/casbin/casbin"
	casbinv2 "github.com/casbin/casbin/v2"
	xormadapter "github.com/casbin/xorm-adapter"
	xormadapter2 "github.com/casbin/xorm-adapter/v2"
	"github.com/devtron-labs/devtron/pkg/sql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"os"
	"strings"
)

type Version string

const (
	CasbinV1 Version = "V1"
	CasbinV2 Version = "V2"
)

var e *casbin.SyncedEnforcer
var e2 *casbinv2.SyncedEnforcer
var enforcerImplRef *EnforcerImpl
var casbinService CasbinService
var casbinVersion Version

type Subject string
type Resource string
type Action string
type Object string
type PolicyType string

func isV2() bool {
	return casbinVersion == CasbinV2
}

func setCasbinVersion() {
	version := os.Getenv("USE_CASBIN_V2")
	if version == "true" {
		casbinVersion = CasbinV2
		return
	}
	casbinVersion = CasbinV1
}

func Create() *casbin.SyncedEnforcer {
	setCasbinVersion()
	if isV2() {
		return nil
	}

	metav1.Now()
	config, err := sql.GetConfig() //FIXME: use this from wire
	if err != nil {
		log.Fatal(err)
	}
	dataSource := fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable", config.User, config.Password, config.Addr, config.Port)
	a, err := xormadapter.NewAdapter("postgres", dataSource, false) // Your driver and data source.
	if err != nil {
		log.Fatal(err)
	}
	auth, err1 := casbin.NewSyncedEnforcerSafe("./auth_model.conf", a)

	if err1 != nil {
		log.Fatal(err1)
	}
	e = auth
	err = e.LoadPolicy()
	log.Println("casbin Policies Loaded Successfully")
	if err != nil {
		log.Fatal(err)
	}
	//adding our key matching func - MatchKeyFunc, to enforcer
	e.AddFunction("matchKeyByPart", MatchKeyByPartFunc)
	return e
}

func CreateV2() *casbinv2.SyncedEnforcer {
	setCasbinVersion()
	if !isV2() {
		return nil
	}

	metav1.Now()
	config, err := sql.GetConfig() //FIXME: use this from wire
	if err != nil {
		log.Fatal(err)
	}
	dataSource := fmt.Sprintf("user=%s password=%s host=%s port=%s sslmode=disable", config.User, config.Password, config.Addr, config.Port)
	a, err := xormadapter2.NewAdapter("postgres", dataSource, false) // Your driver and data source.
	if err != nil {
		log.Fatal(err)
	}
	//Adapter

	auth, err1 := casbinv2.NewSyncedEnforcer("./auth_model.conf", a)
	if err1 != nil {
		log.Fatal(err1)
	}
	e2 = auth
	err = e2.LoadPolicy()
	log.Println("v2 casbin Policies Loaded Successfully")
	if err != nil {
		log.Fatal(err)
	}
	//adding our key matching func - MatchKeyFunc, to enforcer
	e2.AddFunction("matchKeyByPart", MatchKeyByPartFunc)
	return e2
}

func setEnforcerImpl(ref *EnforcerImpl) {
	enforcerImplRef = ref
}
func setCasbinService(service CasbinService) {
	casbinService = service
}

func AddPolicy(policies []Policy) error {
	err := casbinService.AddPolicy(policies)
	if err != nil {
		log.Println("casbin policy addition failed", "err", err)
		return err
	}
	return nil
}

func LoadPolicy() {
	defer HandlePanic()
	isCasbinV2, err := enforcerImplRef.ReloadPolicy()
	if err != nil {
		fmt.Println("error in reloading policies", err)
	} else {
		if isCasbinV2 {
			fmt.Println("V2 policy reloaded successfully")
		} else {
			fmt.Println("policy reloaded successfully")
		}
	}
}

func RemovePolicy(policies []Policy) []Policy {
	policy, err := casbinService.RemovePolicy(policies)
	if err != nil {
		log.Println(err)
	}
	return policy
}

func GetAllSubjects() []string {
	if isV2() {
		return e2.GetAllSubjects()
	}
	return e.GetAllSubjects()
}

func DeleteRoleForUser(user string, role string) bool {
	user = strings.ToLower(user)
	var response bool
	var err error
	if isV2() {
		response, err = e2.DeleteRoleForUser(user, role)
		if err != nil {
			log.Println(err)
		}
	} else {
		response = e.DeleteRoleForUser(user, role)
	}
	enforcerImplRef.InvalidateCache(user)
	return response
}

func GetRolesForUser(user string) ([]string, error) {
	user = strings.ToLower(user)
	if isV2() {
		return e2.GetRolesForUser(user)
	}
	return e.GetRolesForUser(user)
}

func GetUserByRole(role string) ([]string, error) {
	role = strings.ToLower(role)
	if isV2() {
		return e2.GetUsersForRole(role)
	}
	return e.GetUsersForRole(role)
}

func RemovePoliciesByRole(role string) bool {
	role = strings.ToLower(role)
	policyResponse, err := casbinService.RemovePoliciesByRole(role)
	if err != nil {
		return false
	}
	enforcerImplRef.InvalidateCompleteCache()
	return policyResponse
}

func RemovePoliciesByRoles(roles []string) (bool, error) {
	policyResponse, err := casbinService.RemovePoliciesByRoles(roles)
	enforcerImplRef.InvalidateCompleteCache()
	return policyResponse, err
}

func HandlePanic() {
	if err := recover(); err != nil {
		log.Println("panic occurred:", err)
	}
}

type Policy struct {
	Type PolicyType `json:"type"`
	Sub  Subject    `json:"sub"`
	Res  Resource   `json:"res"`
	Act  Action     `json:"act"`
	Obj  Object     `json:"obj"`
}
