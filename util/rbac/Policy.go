package rbac

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/api/bean"
	"strings"
)

func GenerateDefaultPolicies(team string, entityName string, env string) ([]bean.PolicyRequest, error) {
	managerPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"user\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"notification\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:manager_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        }\r\n    ]\r\n}"
	adminPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"*\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:admin_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        }\r\n    ]\r\n}"
	triggerPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"trigger\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"trigger\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:trigger_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        }\r\n    ]\r\n}"
	viewPolicies := "{\r\n    \"data\": [\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"applications\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>/<APP_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"global-environment\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<ENV_OBJ>\"\r\n        },\r\n        {\r\n            \"type\": \"p\",\r\n            \"sub\": \"role:view_<TEAM>_<ENV>_<APP>\",\r\n            \"res\": \"team\",\r\n            \"act\": \"get\",\r\n            \"obj\": \"<TEAM_OBJ>\"\r\n        }\r\n    ]\r\n}"

	managerPolicies = strings.ReplaceAll(managerPolicies, "<TEAM>", team)
	managerPolicies = strings.ReplaceAll(managerPolicies, "<ENV>", env)
	managerPolicies = strings.ReplaceAll(managerPolicies, "<APP>", entityName)

	adminPolicies = strings.ReplaceAll(adminPolicies, "<TEAM>", team)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<ENV>", env)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<APP>", entityName)

	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<TEAM>", team)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<ENV>", env)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<APP>", entityName)

	viewPolicies = strings.ReplaceAll(viewPolicies, "<TEAM>", team)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<ENV>", env)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<APP>", entityName)

	//for START in Casbin Object
	teamObj := team
	envObj := env
	appObj := entityName
	if len(teamObj) == 0 {
		teamObj = "*"
	}
	if len(envObj) == 0 {
		envObj = "*"
	}
	if len(appObj) == 0 {
		appObj = "*"
	}
	managerPolicies = strings.ReplaceAll(managerPolicies, "<TEAM_OBJ>", teamObj)
	managerPolicies = strings.ReplaceAll(managerPolicies, "<ENV_OBJ>", envObj)
	managerPolicies = strings.ReplaceAll(managerPolicies, "<APP_OBJ>", appObj)

	adminPolicies = strings.ReplaceAll(adminPolicies, "<TEAM_OBJ>", teamObj)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<ENV_OBJ>", envObj)
	adminPolicies = strings.ReplaceAll(adminPolicies, "<APP_OBJ>", appObj)

	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<TEAM_OBJ>", teamObj)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<ENV_OBJ>", envObj)
	triggerPolicies = strings.ReplaceAll(triggerPolicies, "<APP_OBJ>", appObj)

	viewPolicies = strings.ReplaceAll(viewPolicies, "<TEAM_OBJ>", teamObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<ENV_OBJ>", envObj)
	viewPolicies = strings.ReplaceAll(viewPolicies, "<APP_OBJ>", appObj)
	//for START in Casbin Object Ends Here

	var policies []bean.PolicyRequest
	var policiesManager bean.PolicyRequest
	err := json.Unmarshal([]byte(managerPolicies), &policiesManager)
	if err != nil {
		fmt.Printf("decode policy err %v for %s\nr", err, managerPolicies)
		return []bean.PolicyRequest{}, err
	}
	policies = append(policies, policiesManager)
	fmt.Printf("add policy request %v\n", policiesManager)
	//casbin.AddPolicy(policiesManager.Data)

	var policiesAdmin bean.PolicyRequest
	err = json.Unmarshal([]byte(adminPolicies), &policiesAdmin)
	if err != nil {
		fmt.Printf("decode policy err %v for %s\nr", err, adminPolicies)
		return []bean.PolicyRequest{}, err
	}
	policies = append(policies, policiesAdmin)

	fmt.Printf("add policy request %v\n", policiesAdmin)
	//casbin.AddPolicy(policiesAdmin.Data)

	var policiesTrigger bean.PolicyRequest
	err = json.Unmarshal([]byte(triggerPolicies), &policiesTrigger)
	if err != nil {
		fmt.Printf("decode policy err %v for %s\nr", err, triggerPolicies)
		return []bean.PolicyRequest{}, err
	}
	policies = append(policies, policiesTrigger)
	fmt.Printf("add policy request %v\n", policiesTrigger)
	//casbin.AddPolicy(policiesTrigger.Data)

	var policiesView bean.PolicyRequest
	err = json.Unmarshal([]byte(viewPolicies), &policiesView)
	if err != nil {
		fmt.Printf("decode policy err %v for %s\nr", err, viewPolicies)
		return []bean.PolicyRequest{}, err
	}
	policies = append(policies, policiesView)
	fmt.Printf("add policy request %v\n", policiesView)
	//casbin.AddPolicy(policiesView.Data)
	return policies, nil
}

