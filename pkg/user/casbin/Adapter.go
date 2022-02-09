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
	"github.com/casbin/xorm-adapter"
	"github.com/devtron-labs/devtron/pkg/sql"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log"
	"strings"
)

var e *casbin.Enforcer

type Subject string
type Resource string
type Action string
type Object string
type PolicyType string

type Policy struct {
	Type PolicyType `json:"type"`
	Sub  Subject    `json:"sub"`
	Res  Resource   `json:"res"`
	Act  Action     `json:"act"`
	Obj  Object     `json:"obj"`
}

func Create() *casbin.Enforcer {
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
	auth, err1 := casbin.NewEnforcerSafe("./auth_model.conf", a)
	if err1 != nil {
		log.Fatal(err1)
	}
	e = auth
	err = e.LoadPolicy()
	if err != nil {
		log.Fatal(err)
	}
	//adding our key matching func - MatchKeyFunc, to enforcer
	e.AddFunction("matchKeyByPart", MatchKeyByPartFunc)
	return e
}

func AddPolicy(policies []Policy) []Policy {
	LoadPolicy()
	var failed = []Policy{}
	for _, p := range policies {
		success := false
		if strings.ToLower(string(p.Type)) == "p" && p.Sub != "" && p.Res != "" && p.Act != "" && p.Obj != "" {
			sub := strings.ToLower(string(p.Sub))
			res := strings.ToLower(string(p.Res))
			act := strings.ToLower(string(p.Act))
			obj := strings.ToLower(string(p.Obj))
			success = e.AddPolicy([]string{sub, res, act, obj, "allow"})
		} else if strings.ToLower(string(p.Type)) == "g" && p.Sub != "" && p.Obj != "" {
			sub := strings.ToLower(string(p.Sub))
			obj := strings.ToLower(string(p.Obj))
			success = e.AddGroupingPolicy([]string{sub, obj})
		}
		if !success {
			failed = append(failed, p)
		}
	}
	if len(policies) != len(failed) {
		err := e.LoadPolicy()
		if err != nil {
			fmt.Println("error in reloading policies", err)
		} else {
			fmt.Println("policy reloaded successfully")
		}
	}
	return failed
}

func LoadPolicy() {
	err := e.LoadPolicy()
	if err != nil {
		fmt.Println("error in reloading policies", err)
	} else {
		fmt.Println("policy reloaded successfully")
	}
}

func RemovePolicy(policies []Policy) []Policy {
	var failed = []Policy{}
	for _, p := range policies {
		success := false
		if strings.ToLower(string(p.Type)) == "p" && p.Sub != "" && p.Res != "" && p.Act != "" && p.Obj != "" {
			success = e.RemovePolicy([]string{strings.ToLower(string(p.Sub)), strings.ToLower(string(p.Res)), strings.ToLower(string(p.Act)), strings.ToLower(string(p.Obj))})
		} else if strings.ToLower(string(p.Type)) == "g" && p.Sub != "" && p.Obj != "" {
			success = e.RemoveGroupingPolicy([]string{strings.ToLower(string(p.Sub)), strings.ToLower(string(p.Obj))})
		}
		if !success {
			failed = append(failed, p)
		}
	}
	if len(policies) != len(failed) {
		_ = e.LoadPolicy()
	}
	return failed
}

func GetAllSubjects() []string {
	return e.GetAllSubjects()
}

func DeleteRoleForUser(user string, role string) bool {
	user = strings.ToLower(user)
	return e.DeleteRoleForUser(user, role)
}

func GetRolesForUser(user string) ([]string, error) {
	user = strings.ToLower(user)
	return e.GetRolesForUser(user)
}

func GetUserByRole(role string) ([]string, error) {
	role = strings.ToLower(role)
	return e.GetUsersForRole(role)
}
