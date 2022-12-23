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
	"github.com/devtron-labs/authenticator/middleware"
	casbin2 "github.com/devtron-labs/devtron/pkg/user/casbin"
	"go.uber.org/zap"
	"time"
)

type EnterpriseEnforcerImpl struct {
	Config *EnterpriseEnforcerConfig
	*casbin2.EnforcerImpl
}

type EnterpriseEnforcerConfig struct {
	EnterpriseEnforcerEnabled bool `env:"ENTERPRISE_ENFORCER_ENABLED" envDefault:"true"`
}

func NewEnterpriseEnforcerImpl(enforcer *casbin.SyncedEnforcer,
	sessionManager *middleware.SessionManager,
	logger *zap.SugaredLogger) (*EnterpriseEnforcerImpl, error) {
	enforcerImpl := casbin2.NewEnforcerImpl(enforcer, sessionManager, logger)
	enforcerConfig := &EnterpriseEnforcerConfig{}
	err := env.Parse(enforcerConfig)
	if err != nil {
		logger.Fatal("error occurred while parsing enforcer config", err)
	}
	logger.Infow("enforcer initialized", "Config", enforcerConfig)
	return &EnterpriseEnforcerImpl{EnforcerImpl: enforcerImpl, Config: enforcerConfig}, nil
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
	defer casbin2.HandlePanic()
	functions := make(map[string]govaluate.ExpressionFunction)
	enforcedModel := e.SyncedEnforcer.Enforcer.GetModel()
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
	filteredPolicies := e.GetFilteredPolicies(subject, enforcedModel["p"]["p"].Policy, enforcedModel["g"]["g"].RM)
	eft := effect.NewDefaultEffector()
	for _, resourceItem := range resourceItems {
		rvals := e.getRval(subject, resource, action, resourceItem)
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
				// log.LogPrint("Policy Rule: ", pvals)
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
				// log.LogPrint("Result: ", result)

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

func (e *EnterpriseEnforcerImpl) getRval(rval ...interface{}) []interface{} {
	return rval
}

func (e *EnterpriseEnforcerImpl) GetFilteredPolicies(subject string, policies [][]string, rm rbac.RoleManager) [][]string {
	var filteredPolicies [][]string
	for _, policy := range policies {
		role := policy[0]
		hasLink, _ := rm.HasLink(subject, role)
		if hasLink {
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
