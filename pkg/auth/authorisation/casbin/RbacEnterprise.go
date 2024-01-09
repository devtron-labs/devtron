package casbin

import (
	"errors"
	"fmt"
	"github.com/devtron-labs/authenticator/jwt"
	auth "github.com/devtron-labs/devtron/pkg/auth/authorisation/globalConfig"
	util3 "github.com/devtron-labs/devtron/pkg/auth/user/util"
	"strings"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/caarlos0/env/v6"
	"github.com/casbin/casbin"
	"github.com/casbin/casbin/effect"
	"github.com/casbin/casbin/model"
	"github.com/casbin/casbin/rbac"
	"github.com/casbin/casbin/util"
	casbinv2 "github.com/casbin/casbin/v2"
	modelv2 "github.com/casbin/casbin/v2/model"
	rbac2 "github.com/casbin/casbin/v2/rbac"
	"github.com/devtron-labs/authenticator/middleware"
	"go.uber.org/zap"
)

type EnterpriseEnforcerImpl struct {
	Config *EnterpriseEnforcerConfig
	*EnforcerImpl
}

type DefinitionType int

const (
	PTypeDefinition DefinitionType = iota
	GTypeDefinition
)

type EnterpriseEnforcerConfig struct {
	EnterpriseEnforcerEnabled bool `env:"ENTERPRISE_ENFORCER_ENABLED" envDefault:"true"`
	UseCustomEnforcer         bool `env:"USE_CUSTOM_ENFORCER" envDefault:"true"`
	UseCasbinV2               bool `env:"USE_CASBIN_V2" envDefault:"false"`
	CustomRoleCacheAllowed    bool `env:"CUSTOM_ROLE_CACHE_ALLOWED" envDefault:"false"`
}

func NewEnterpriseEnforcerImpl(enforcer *casbin.SyncedEnforcer, enforcerV2 *casbinv2.SyncedEnforcer,
	sessionManager *middleware.SessionManager,
	logger *zap.SugaredLogger, casbinService CasbinService, globalAuthConfigService auth.GlobalAuthorisationConfigService) (*EnterpriseEnforcerImpl, error) {
	enforcerImpl := NewEnforcerImpl(enforcer, enforcerV2, sessionManager, logger, casbinService, globalAuthConfigService)
	enforcerConfig := &EnterpriseEnforcerConfig{}
	err := env.Parse(enforcerConfig)
	if err != nil {
		logger.Fatal("error occurred while parsing enforcer config", err)
	}
	logger.Infow("enforcer initialized", "Config", enforcerConfig)
	return &EnterpriseEnforcerImpl{EnforcerImpl: enforcerImpl, Config: enforcerConfig}, nil
}

func (e *EnforcerImpl) VerifyAndGetSubjectsForEnforcement(tokenString string) ([]string, bool) {
	claims, err := e.SessionManager.VerifyToken(tokenString)
	if err != nil {
		return nil, true
	}
	mapClaims, err := jwt.MapClaims(claims)
	if err != nil {
		return nil, true
	}
	subjects := make([]string, 0)
	email, groups := e.globalAuthConfigService.GetEmailAndGroupsFromClaims(mapClaims)
	if e.globalAuthConfigService.IsDevtronSystemManagedConfigActive() || util3.CheckIfAdminOrApiToken(email) {
		subjects = append(subjects, email)
	}
	if e.globalAuthConfigService.IsGroupClaimsConfigActive() {
		groupsCasbinNames := util3.GetGroupCasbinName(groups)
		subjects = append(subjects, groupsCasbinNames...)
	}
	for i := range subjects {
		subjects[i] = strings.ToLower(subjects[i])
	}
	return subjects, false
}

func (e *EnterpriseEnforcerImpl) Enforce(token string, resource string, action string, resourceItem string) bool {
	subjects, invalid := e.VerifyAndGetSubjectsForEnforcement(token)
	if invalid {
		return false
	}
	return e.EnforceBySubjects(subjects, resource, action, resourceItem)
}

func (e *EnterpriseEnforcerImpl) EnforceBySubjects(subjects []string, resource string, action string, resourceItem string) bool {
	if e.Config.EnterpriseEnforcerEnabled {
		enforceResponse := e.EnforceForSubjectsInBatch(subjects, resource, action, []string{resourceItem})
		if len(enforceResponse) > 0 {
			return enforceResponse[0]
		}
		return false
	}
	// to use oss impl, running loop
	result := false
	for i := range subjects {
		resultItr := e.EnforcerImpl.EnforceByEmail(subjects[i], resource, action, resourceItem)
		if resultItr {
			result = resultItr
			break
		}
	}
	return result
}

func (e *EnterpriseEnforcerImpl) EnforceInBatch(token string, resource string, action string, vals []string) map[string]bool {
	subjects, invalid := e.VerifyAndGetSubjectsForEnforcement(token)
	if invalid {
		return make(map[string]bool)
	}
	return e.EnforceBySubjectsInBatch(subjects, resource, action, vals)
}

func (e *EnterpriseEnforcerImpl) EnforceBySubjectsInBatch(subjects []string, resource string, action string, resourceItems []string) map[string]bool {
	if e.Config.EnterpriseEnforcerEnabled {
		timestamp := time.Now()
		enforcerResponse := e.EnforceForSubjectsInBatch(subjects, resource, action, resourceItems)
		responseMap := make(map[string]bool)
		for index, resourceItem := range resourceItems {
			response := false
			if len(enforcerResponse) > index {
				response = enforcerResponse[index]
			}
			responseMap[resourceItem] = response
		}
		timegap := time.Since(timestamp)
		e.EnforcerImpl.Logger.Infow("enforce in batch ", "subjects", subjects, "resource", resource, "action",
			action, "actualSize", len(resourceItems), "returnedSize", len(responseMap), "timetakenInMillis", timegap.Milliseconds())
		return responseMap
	}
	// to use oss impl, running loop
	result := make(map[string]bool)
	for i := range subjects {
		resultItr := e.EnforcerImpl.EnforceByEmailInBatch(subjects[i], resource, action, resourceItems)
		for key, value := range resultItr {
			oldValue, ok := result[key]
			if !ok || !oldValue { //old value is not present or false, only need to update in these cases because if true then already authenticated
				result[key] = value
			}
		}
	}
	return result
}
func (e *EnterpriseEnforcerImpl) EnforceForSubjectsInBatch(subjects []string, resource string, action string, resourceItems []string) (resultArr []bool) {
	var enforcedModelV1 model.Model
	if e.Config.UseCustomEnforcer {
		enforcedModel := CustomEnforcedModel{}
		if e.Config.UseCasbinV2 {
			enforcedModelV2 := e.EnforcerV2.Enforcer.GetModel()
			enforcedModel.modelV2 = enforcedModelV2
			enforcedModel.version = CasbinV2
		} else {
			enforcedModelV1 = e.Enforcer.Enforcer.GetModel()
			enforcedModel.modelV1 = enforcedModelV1
			enforcedModel.version = CasbinV1
		}
		return e.EnforceForSubjectInBatchCustom(subjects, resource, action, resourceItems, enforcedModel)
	} else {
		if !e.Config.UseCasbinV2 {
			enforcedModelV1 = e.Enforcer.Enforcer.GetModel()
		}
		return e.EnforceForSubjectInBatchCasbin(subjects, resource, action, resourceItems, enforcedModelV1)
	}
}

func (e *EnterpriseEnforcerImpl) EnforceForSubjectInBatchCasbin(subjects []string, resource string, action string, resourceItems []string, enforcedModel model.Model) []bool {
	defer HandlePanic()
	functions := make(map[string]govaluate.ExpressionFunction)
	fm := model.LoadFunctionMap()
	for key, function := range fm {
		functions[key] = function
	}
	functions["matchKeyByPart"] = MatchKeyByPartFunc

	if _, ok := enforcedModel["g"]; ok {
		for key, ast := range enforcedModel["g"] {
			rm := ast.RM
			functions[key] = util.GenerateGFunction(rm)
		}
	}

	expString := enforcedModel["m"]["m"].Value
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(expString, functions)
	if err != nil {
		panic(err)
	}

	rTokens := make(map[string]int, len(enforcedModel["r"]["r"].Tokens))
	for i, token := range enforcedModel["r"]["r"].Tokens {
		rTokens[token] = i
	}
	pTokens := make(map[string]int, len(enforcedModel["p"]["p"].Tokens))
	for i, token := range enforcedModel["p"]["p"].Tokens {
		pTokens[token] = i
	}
	filteredPolicies := e.GetFilteredPolicies(subjects, resource, action, enforcedModel["p"]["p"].Policy, CustomRoleManager{
		rmV1:    enforcedModel["g"]["g"].RM,
		version: CasbinV1,
	})
	eft := effect.NewDefaultEffector()

	resultArr := make([]bool, len(resourceItems))
	for _, subject := range subjects {
		for m, resourceItem := range resourceItems {
			if resultArr[m] {
				continue
			} else { // if not true then only check again with another subject
				rvals := e.getRval(subject, resource, action, strings.ToLower(resourceItem))
				parameters := enforceParameters{
					rTokens: rTokens,
					rVals:   rvals,

					pTokens: pTokens,
				}

				var policyEffects []effect.Effect
				var matcherResults []float64
				if policyLen := len(filteredPolicies); policyLen != 0 {
					policyEffects = make([]effect.Effect, policyLen)
					matcherResults = make([]float64, policyLen)
					if len(enforcedModel["r"]["r"].Tokens) != len(rvals) {
						panic(
							fmt.Sprintf(
								"Invalid Request Definition size: expected %d got %d rvals: %v",
								len(enforcedModel["r"]["r"].Tokens),
								len(rvals),
								rvals))
					}
					for i, pvals := range filteredPolicies {
						if len(enforcedModel["p"]["p"].Tokens) != len(pvals) {
							panic(
								fmt.Sprintf(
									"Invalid Policy Rule size: expected %d got %d pvals: %v",
									len(enforcedModel["p"]["p"].Tokens),
									len(pvals),
									pvals))
						}

						parameters.pVals = pvals

						result, err := expression.Eval(parameters)

						if err != nil {
							policyEffects[i] = effect.Indeterminate
							panic(err)
						}

						switch result := result.(type) {
						case bool:
							if !result {
								policyEffects[i] = effect.Indeterminate
								continue
							}
						case float64:
							if result == 0 {
								policyEffects[i] = effect.Indeterminate
								continue
							} else {
								matcherResults[i] = result
							}
						default:
							panic(errors.New("matcher result should be bool, int or float"))
						}

						if j, ok := parameters.pTokens["p_eft"]; ok {
							eft := parameters.pVals[j]
							if eft == "allow" {
								policyEffects[i] = effect.Allow
							} else if eft == "deny" {
								policyEffects[i] = effect.Deny
							} else {
								policyEffects[i] = effect.Indeterminate
							}
						} else {
							policyEffects[i] = effect.Allow
						}

						if enforcedModel["e"]["e"].Value == "priority(p_eft) || deny" {
							break
						}

					}
				} else {
					policyEffects = make([]effect.Effect, 1)
					matcherResults = make([]float64, 1)

					parameters.pVals = make([]string, len(parameters.pTokens))

					result, err := expression.Eval(parameters)
					// log.LogPrint("Result: ", result)

					if err != nil {
						policyEffects[0] = effect.Indeterminate
						panic(err)
					}

					if result.(bool) {
						policyEffects[0] = effect.Allow
					} else {
						policyEffects[0] = effect.Indeterminate
					}
				}

				// log.LogPrint("Rule Results: ", policyEffects)
				result, err := eft.MergeEffects(enforcedModel["e"]["e"].Value, policyEffects, matcherResults)
				if err != nil {
					panic(err)
				}
				resultArr[m] = result
			}
		}
	}
	return resultArr
}

func (e *EnterpriseEnforcerImpl) EnforceForSubjectInBatchCustom(subjects []string, resource string, action string, resourceItems []string, enforcedModel CustomEnforcedModel) (resultArr []bool) {
	defer HandlePanic()
	filteredPolicies := e.GetFilteredPolicies(subjects, resource, action, enforcedModel.getPolicy("p"), enforcedModel.getCustomRM("g", e.Config.CustomRoleCacheAllowed))
	subjectForRVal := "" //using empty string as subject in rVal because for custom enforcer we do not need subject value
	for _, resourceItem := range resourceItems {
		rVals := e.getRvalCustom(subjectForRVal, resource, action, strings.ToLower(resourceItem))
		result := false
		if policyLen := len(filteredPolicies); policyLen != 0 {
			if len(enforcedModel.getTokens("r")) != len(rVals) { //will break if our code assumptions for definition check and auth_model.conf mismatch
				panic(
					fmt.Sprintf(
						"Invalid Request Definition size: expected %d got %d rVals: %v",
						len(enforcedModel.getTokens("r")),
						len(rVals),
						rVals))
			}
			for _, pVals := range filteredPolicies {
				if len(enforcedModel.getTokens("p")) != len(pVals) { //will break if our code assumptions for definition check and auth_model.conf mismatch
					panic(
						fmt.Sprintf(
							"Invalid Policy Rule size: expected %d got %d pVals: %v",
							len(enforcedModel.getTokens("p")),
							len(pVals),
							pVals))
				}
				definitionEvalResult := e.EvaluateDefinitions(PTypeDefinition, pVals, rVals)
				if !definitionEvalResult { //continuing on getting deny or indeterminate (not allow)
					continue
				}
				//assumptions
				//1. every policy have effect at 4th place, order - [sub, res, act, obj, eft]
				//2. assuming policy effect is "some(where (p_eft == allow)) && !some(where (p_eft == deny))"
				eft := pVals[4]
				if eft == "allow" {
					result = true
				} else if eft == "deny" {
					result = false
					break
				}
			}
		}
		resultArr = append(resultArr, result)
	}
	return resultArr
}

type CustomEnforcedModel struct {
	modelV1 model.Model
	modelV2 modelv2.Model
	version Version
}

func (model CustomEnforcedModel) getTokens(tokenKey string) []string {
	switch model.version {
	case CasbinV2:
		return model.modelV2[tokenKey][tokenKey].Tokens
	default:
		return model.modelV1[tokenKey][tokenKey].Tokens
	}
}

func (model CustomEnforcedModel) getPolicy(tokenKey string) [][]string {
	switch model.version {
	case CasbinV2:
		return model.modelV2[tokenKey][tokenKey].Policy
	default:
		return model.modelV1[tokenKey][tokenKey].Policy
	}
}

func (model CustomEnforcedModel) getCustomRM(tokenKey string, customRoleCacheAllowed bool) CustomRoleManager {
	switch model.version {
	case CasbinV2:
		return CustomRoleManager{
			rmV2:                   model.modelV2[tokenKey][tokenKey].RM,
			version:                CasbinV2,
			roles:                  make(map[string]bool),
			customRoleCacheAllowed: customRoleCacheAllowed,
		}
	default:
		return CustomRoleManager{
			rmV1:    model.modelV1[tokenKey][tokenKey].RM,
			version: CasbinV1,
		}
	}
}

type CustomRoleManager struct {
	customRoleCacheAllowed bool
	roles                  map[string]bool
	rolesUpdated           bool
	rmV1                   rbac.RoleManager
	rmV2                   rbac2.RoleManager
	version                Version
}

func (rm *CustomRoleManager) hasLink(subjects []string, role string) (bool, error) {
	for i := range subjects {
		hasLinkItr := false
		var err error
		switch rm.version {
		case CasbinV2:
			if rm.rolesUpdated {
				_, ok := rm.roles[role]
				return ok, nil
			}
			hasLinkItr, err = rm.rmV2.HasLink(subjects[i], role)
		default:
			hasLinkItr, err = rm.rmV1.HasLink(subjects[i], role)
		}
		if err != nil || hasLinkItr {
			return hasLinkItr, err
		}
	}
	return false, nil
}

func (rm *CustomRoleManager) checkAndUpdateSubjectRolesCache(subject string) {
	if !rm.customRoleCacheAllowed {
		return
	}
	roles, err := rm.rmV2.GetRoles(subject)
	if err != nil {
		return
	}
	result := rm.updateRolesCache(roles)
	rm.rolesUpdated = result
}

func (rm *CustomRoleManager) updateRolesCache(roles []string) bool {
	for _, role := range roles {
		rm.roles[role] = true
		roles1, err := rm.rmV2.GetRoles(role)
		if err != nil {
			return false
		}
		if result := rm.updateRolesCache(roles1); !result {
			return false
		}
	}
	return true
}

func (e *EnterpriseEnforcerImpl) EvaluateDefinitions(t DefinitionType, pVals, rVals []string) bool {
	switch t {
	case PTypeDefinition:
		return e.EvaluatePTypeDefinition(pVals, rVals)
	default:
		return false
	}
}

func (e *EnterpriseEnforcerImpl) EvaluatePTypeDefinition(pVals, rVals []string) bool {
	result := true
	if len(rVals) > len(pVals) || len(pVals) < 4 || len(rVals) < 4 { //need minimum 4 values for evaluating; values are - [sub, res, act, obj]
		result = false
	} else {
		//only checking resourceObject and not resource, action
		//because we have already got filtered policies on the basis of their matching
		result = MatchKeyByPart(rVals[3], pVals[3])
	}
	return result
}

func (e *EnterpriseEnforcerImpl) getRval(rVal ...interface{}) []interface{} {
	return rVal
}

func (e *EnterpriseEnforcerImpl) getRvalCustom(rVal ...string) []string {
	return rVal
}

func (e *EnterpriseEnforcerImpl) GetFilteredPolicies(subjects []string, resource string, action string, policies [][]string, rm CustomRoleManager) [][]string {
	var filteredPolicies [][]string
	for i := range subjects {
		rm.checkAndUpdateSubjectRolesCache(subjects[i])
	}
	for _, policy := range policies {
		role := policy[0]
		policyResource := policy[1]
		policyAction := policy[2]
		hasLink, _ := rm.hasLink(subjects, role)
		if hasLink {
			if !MatchKeyByPart(action, policyAction) {
				continue
			}
			if !MatchKeyByPart(resource, policyResource) {
				continue
			}
			filteredPolicies = append(filteredPolicies, policy)
		}
	}
	return filteredPolicies
}

type enforceParameters struct {
	rTokens map[string]int
	rVals   []interface{}

	pTokens map[string]int
	pVals   []string
}

// implements govaluate.Parameters
func (p enforceParameters) Get(name string) (interface{}, error) {
	if name == "" {
		return nil, nil
	}

	switch name[0] {
	case 'p':
		i, ok := p.pTokens[name]
		if !ok {
			return nil, errors.New("No parameter '" + name + "' found.")
		}
		return p.pVals[i], nil
	case 'r':
		i, ok := p.rTokens[name]
		if !ok {
			return nil, errors.New("No parameter '" + name + "' found.")
		}
		return p.rVals[i], nil
	default:
		return nil, errors.New("No parameter '" + name + "' found.")
	}
}
