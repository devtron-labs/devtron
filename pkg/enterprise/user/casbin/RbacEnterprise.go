package casbin

import (
	"errors"
	"fmt"
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
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	"go.uber.org/zap"
	"strings"
	"time"
)

type EnterpriseEnforcerImpl struct {
	Config *EnterpriseEnforcerConfig
	*casbin2.EnforcerImpl
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
}

func NewEnterpriseEnforcerImpl(enforcer *casbin.SyncedEnforcer, enforcerV2 *casbinv2.SyncedEnforcer,
	sessionManager *middleware.SessionManager,
	logger *zap.SugaredLogger, casbinService casbin2.CasbinService) (*EnterpriseEnforcerImpl, error) {
	enforcerImpl := casbin2.NewEnforcerImpl(enforcer, enforcerV2, sessionManager, logger, casbinService)
	enforcerConfig := &EnterpriseEnforcerConfig{}
	err := env.Parse(enforcerConfig)
	if err != nil {
		logger.Fatal("error occurred while parsing enforcer config", err)
	}
	logger.Infow("enforcer initialized", "Config", enforcerConfig)
	return &EnterpriseEnforcerImpl{EnforcerImpl: enforcerImpl, Config: enforcerConfig}, nil
}

func (e *EnterpriseEnforcerImpl) Enforce(token string, resource string, action string, resourceItem string) bool {
	email, invalid := e.VerifyTokenAndGetEmail(token)
	if invalid {
		return false
	}
	return e.EnforceByEmail(strings.ToLower(email), resource, action, resourceItem)
}

func (e *EnterpriseEnforcerImpl) EnforceByEmail(emailId string, resource string, action string, resourceItem string) bool {
	if e.Config.EnterpriseEnforcerEnabled {
		enforceResponse := e.EnforceForSubjectInBatch(emailId, resource, action, []string{resourceItem})
		if len(enforceResponse) > 0 {
			return enforceResponse[0]
		}
		return false
	}
	return e.EnforcerImpl.EnforceByEmail(emailId, resource, action, resourceItem)
}

func (e *EnterpriseEnforcerImpl) EnforceByEmailInBatch(emailId string, resource string, action string, resourceItems []string) map[string]bool {
	emailId = strings.ToLower(emailId)
	if e.Config.EnterpriseEnforcerEnabled {
		timestamp := time.Now()
		enforcerResponse := e.EnforceForSubjectInBatch(emailId, resource, action, resourceItems)
		responseMap := make(map[string]bool)
		for index, resourceItem := range resourceItems {
			response := false
			if len(enforcerResponse) > index {
				response = enforcerResponse[index]
			}
			responseMap[resourceItem] = response
		}
		timegap := time.Since(timestamp)
		e.EnforcerImpl.Logger.Infow("enforce in batch ", "email", emailId, "resource", resource, "action",
			action, "actualSize", len(resourceItems), "returnedSize", len(responseMap), "timetakenInMillis", timegap.Milliseconds())
		return responseMap
	}
	return e.EnforcerImpl.EnforceByEmailInBatch(emailId, resource, action, resourceItems)
}
func (e *EnterpriseEnforcerImpl) EnforceForSubjectInBatch(subject string, resource string, action string, resourceItems []string) (resultArr []bool) {
	var enforcedModelV1 model.Model
	if e.Config.UseCustomEnforcer {
		enforcedModel := CustomEnforcedModel{}
		if e.Config.UseCasbinV2 {
			enforcedModelV2 := e.EnforcerV2.Enforcer.GetModel()
			enforcedModel.modelV2 = enforcedModelV2
			enforcedModel.version = casbin2.CasbinV2
		} else {
			enforcedModelV1 = e.Enforcer.Enforcer.GetModel()
			enforcedModel.modelV1 = enforcedModelV1
			enforcedModel.version = casbin2.CasbinV1
		}
		return e.EnforceForSubjectInBatchCustom(subject, resource, action, resourceItems, enforcedModel)
	} else {
		if !e.Config.UseCasbinV2 {
			enforcedModelV1 = e.Enforcer.Enforcer.GetModel()
		}
		return e.EnforceForSubjectInBatchCasbin(subject, resource, action, resourceItems, enforcedModelV1)
	}
}

func (e *EnterpriseEnforcerImpl) EnforceForSubjectInBatchCasbin(subject string, resource string, action string, resourceItems []string, enforcedModel model.Model) (resultArr []bool) {
	defer casbin2.HandlePanic()
	functions := make(map[string]govaluate.ExpressionFunction)
	fm := model.LoadFunctionMap()
	for key, function := range fm {
		functions[key] = function
	}
	functions["matchKeyByPart"] = casbin2.MatchKeyByPartFunc

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
	filteredPolicies := e.GetFilteredPolicies(subject, resource, action, enforcedModel["p"]["p"].Policy, CustomRoleManager{
		rmV1:    enforcedModel["g"]["g"].RM,
		version: casbin2.CasbinV1,
	})
	eft := effect.NewDefaultEffector()
	for _, resourceItem := range resourceItems {
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

		resultArr = append(resultArr, result)
	}
	return resultArr
}

func (e *EnterpriseEnforcerImpl) EnforceForSubjectInBatchCustom(subject string, resource string, action string, resourceItems []string, enforcedModel CustomEnforcedModel) (resultArr []bool) {
	defer casbin2.HandlePanic()
	filteredPolicies := e.GetFilteredPolicies(subject, resource, action, enforcedModel.getPolicy("p"), enforcedModel.getCustomRM("g"))

	for _, resourceItem := range resourceItems {
		rVals := e.getRvalCustom(subject, resource, action, strings.ToLower(resourceItem))
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
	version casbin2.Version
}

func (model CustomEnforcedModel) getTokens(tokenKey string) []string {
	switch model.version {
	case casbin2.CasbinV2:
		return model.modelV2[tokenKey][tokenKey].Tokens
	default:
		return model.modelV1[tokenKey][tokenKey].Tokens
	}
}

func (model CustomEnforcedModel) getPolicy(tokenKey string) [][]string {
	switch model.version {
	case casbin2.CasbinV2:
		return model.modelV2[tokenKey][tokenKey].Policy
	default:
		return model.modelV1[tokenKey][tokenKey].Policy
	}
}

func (model CustomEnforcedModel) getCustomRM(tokenKey string) CustomRoleManager {
	switch model.version {
	case casbin2.CasbinV2:
		return CustomRoleManager{
			rmV2:    model.modelV2[tokenKey][tokenKey].RM,
			version: casbin2.CasbinV2,
		}
	default:
		return CustomRoleManager{
			rmV1:    model.modelV1[tokenKey][tokenKey].RM,
			version: casbin2.CasbinV1,
		}
	}
}

type CustomRoleManager struct {
	rmV1    rbac.RoleManager
	rmV2    rbac2.RoleManager
	version casbin2.Version
}

func (rm CustomRoleManager) hasLink(subject string, role string) (bool, error) {
	switch rm.version {
	case casbin2.CasbinV2:
		return rm.rmV2.HasLink(subject, role)
	default:
		return rm.rmV1.HasLink(subject, role)
	}
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
		result = casbin2.MatchKeyByPart(rVals[3], pVals[3])
	}
	return result
}

func (e *EnterpriseEnforcerImpl) getRval(rVal ...interface{}) []interface{} {
	return rVal
}

func (e *EnterpriseEnforcerImpl) getRvalCustom(rVal ...string) []string {
	return rVal
}

func (e *EnterpriseEnforcerImpl) GetFilteredPolicies(subject string, resource string, action string, policies [][]string, rm CustomRoleManager) [][]string {
	var filteredPolicies [][]string
	for _, policy := range policies {
		role := policy[0]
		policyResource := policy[1]
		policyAction := policy[2]
		hasLink, _ := rm.hasLink(subject, role)
		e.Logger.Debugw("casbin version in use", "version", rm.version)
		if hasLink {
			if !casbin2.MatchKeyByPart(action, policyAction) {
				continue
			}
			if !casbin2.MatchKeyByPart(resource, policyResource) {
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
