package bean

import (
	"encoding/json"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/sql"
	jsonpatch1 "github.com/evanphx/json-patch"
	"github.com/mattbaird/jsonpatch"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"strings"
)

type LockConfigRequest struct {
	Allowed bool     `json:"allowed, notnull"`
	Config  []string `json:"config, notnull"`
}

type LockConfigResponse struct {
	Id      int      `json:"id,pk"`
	Allowed bool     `json:"allowed, notnull"`
	Config  []string `json:"config, notnull"`
}

type LockConfiguration struct {
	tableName struct{} `sql:"lock_configuration"`
	Id        int      `sql:"id,pk"`
	Config    string   `sql:"config"`
	Active    bool     `sql:"active"`
	sql.AuditLog
}

type LockConfig struct {
	Path    string
	Allowed bool
}

func (impl *LockConfiguration) ConvertDBDtoToResponse() *LockConfigResponse {
	config, allowed := getConfigAndStatus(impl.Config)
	return &LockConfigResponse{
		Id:      impl.Id,
		Config:  config,
		Allowed: allowed,
	}
}

func (impl *LockConfigRequest) ConvertRequestToDBDto() *LockConfiguration {
	config := impl.getLockConfig()
	return &LockConfiguration{
		Config: config,
		Active: true,
	}
}

func getConfigAndStatus(config string) ([]string, bool) {
	var configs []string
	allowed := true
	var lockConfigs []LockConfig
	_ = json.Unmarshal([]byte(config), &lockConfigs)
	for _, lockConfig := range lockConfigs {
		configs = append(configs, lockConfig.Path)
		allowed = allowed && lockConfig.Allowed
	}
	return configs, allowed

}

func (impl *LockConfigRequest) getLockConfig() string {
	var lockConfigs []LockConfig
	for _, config := range impl.Config {
		lockConfig := LockConfig{
			Path:    config,
			Allowed: impl.Allowed,
		}
		lockConfigs = append(lockConfigs, lockConfig)
	}
	byteConfig, _ := json.Marshal(lockConfigs)
	return string(byteConfig)
}

func CreateConfigEmptyJsonPatch(config string) ([]jsonpatch.JsonPatchOperation, error) {
	emptyJson := `{
    }`
	patch1, err := jsonpatch.CreatePatch([]byte(config), []byte(emptyJson))
	if err != nil {
		fmt.Printf("Error creating JSON patch: %v", err)
		return nil, err
	}
	return patch1, nil
}

func GetJsonParentPathMap(patch []jsonpatch.JsonPatchOperation) map[string]bool {
	paths := make(map[string]bool)
	for _, path := range patch {
		// As path start with '/' we are splitting the string and getting the second element eg:- /a/b/c result :- [ , a, b, c]
		res := strings.Split(path.Path, "/")
		paths["/"+res[1]] = true
	}
	return paths
}

func ModifyEmptyPatchBasedOnChanges(patch []jsonpatch.JsonPatchOperation, paths map[string]bool) []jsonpatch.JsonPatchOperation {
	for index, path := range patch {
		if paths[path.Path] {
			patch = append(patch[:index], patch[index+1:]...)
		}
	}
	return patch
}

func ApplyJsonPatch(patch []jsonpatch.JsonPatchOperation, config string) (string, error) {
	marsh, err := json.Marshal(patch)
	if err != nil {
		return "", err
	}
	decodedPatch, err := jsonpatch1.DecodePatch(marsh)
	if err != nil {
		return "", err
	}
	modified, err := decodedPatch.Apply([]byte(config))
	if err != nil {
		return "", err
	}
	return string(modified), nil
}

func CheckForLockedKeyInModifiedJson(lockConfig *LockConfigResponse, configJson string) bool {
	isLockConfigError := true
	obj, err := oj.ParseString(configJson)
	if err != nil {
		return false
	}
	for _, config := range lockConfig.Config {
		x, err := jp.ParseString(config)
		if err != nil {
			return false
		}
		ys := x.Get(obj)
		if len(ys) != 0 {
			isLockConfigError = true
		}
	}
	return isLockConfigError
}
