package casbin

import (
	"context"
	"fmt"
	"github.com/devtron-labs/devtron/pkg/user/casbin/client"
	"go.uber.org/zap"
	"log"
)

type CasbinService interface {
	AddPolicyTest(iterations int)
	AddPolicy(policies []Policy) ([]Policy, error)
	LoadPolicy()
	RemovePolicy(policies []Policy) ([]Policy, error)
	GetAllSubjects() ([]string, error)
	DeleteRoleForUser(user, role string) (bool, error)
	GetRolesForUser(user string) ([]string, error)
	GetUserByRole(role string) ([]string, error)
	RemovePoliciesByRole(role string) (bool, error)
	RemovePoliciesByRoles(roles []string) (bool, error)
}

type CasbinServiceImpl struct {
	logger       *zap.SugaredLogger
	casbinClient client.CasbinClient
}

func NewCasbinServiceImpl(logger *zap.SugaredLogger,
	casbinClient client.CasbinClient) *CasbinServiceImpl {
	return &CasbinServiceImpl{
		logger:       logger,
		casbinClient: casbinClient,
	}
}

func (impl *CasbinServiceImpl) AddPolicy(policies []Policy) ([]Policy, error) {
	convertedPolicies := make([]*client.Policy, len(policies))
	for _, policy := range policies {
		convertedPolicy := &client.Policy{
			Type: string(policy.Type),
			Sub:  string(policy.Sub),
			Res:  string(policy.Res),
			Act:  string(policy.Act),
			Obj:  string(policy.Obj),
		}
		convertedPolicies = append(convertedPolicies, convertedPolicy)
	}
	in := &client.MultiPolicyObj{
		Policies: convertedPolicies,
	}
	_, err := impl.casbinClient.AddPolicy(context.Background(), in)
	if err != nil {
		return nil, err
	}
	return policies, nil
}

func (impl *CasbinServiceImpl) AddPolicyTest(iterations int) {
	convertedPolicies := make([]*client.Policy, 0, iterations)
	for i := 0; i < iterations; i++ {
		policy := Policy{Type: PolicyType(fmt.Sprintf("test-%v", i)), Sub: Subject(fmt.Sprintf("efgh-%v", i)), Res: Resource(fmt.Sprintf("abcd-%v", i)), Act: "view", Obj: Object(fmt.Sprintf("xyz-%v", i))}
		convertedPolicy := &client.Policy{
			Type: string(policy.Type),
			Sub:  string(policy.Sub),
			Res:  string(policy.Res),
			Act:  string(policy.Act),
			Obj:  string(policy.Obj),
		}
		convertedPolicies = append(convertedPolicies, convertedPolicy)
	}
	in := &client.MultiPolicyObj{
		Policies: convertedPolicies,
	}
	_, err := impl.casbinClient.AddPolicy(context.Background(), in)
	if err != nil {
		log.Println("error in addPolicyTest method", err)
	}
	return
}

func (impl *CasbinServiceImpl) LoadPolicy() {
	in := &client.EmptyObj{}
	_, _ = impl.casbinClient.LoadPolicy(context.Background(), in)
	return
}

func (impl *CasbinServiceImpl) RemovePolicy(policies []Policy) ([]Policy, error) {
	convertedPolicies := make([]*client.Policy, len(policies))
	for _, policy := range policies {
		convertedPolicy := &client.Policy{
			Type: string(policy.Type),
			Sub:  string(policy.Sub),
			Res:  string(policy.Res),
			Act:  string(policy.Act),
			Obj:  string(policy.Obj),
		}
		convertedPolicies = append(convertedPolicies, convertedPolicy)
	}
	in := &client.MultiPolicyObj{
		Policies: convertedPolicies,
	}
	_, err := impl.casbinClient.RemovePolicy(context.Background(), in)
	if err != nil {
		return nil, err
	}
	return policies, nil
}
func (impl *CasbinServiceImpl) GetAllSubjects() ([]string, error) {
	in := &client.EmptyObj{}
	resp, err := impl.casbinClient.GetAllSubjects(context.Background(), in)
	if err != nil {
		return nil, err
	}
	return resp.Subjects, nil
}
func (impl *CasbinServiceImpl) DeleteRoleForUser(user, role string) (bool, error) {
	in := &client.DeleteRoleForUserRequest{
		User: user,
		Role: role,
	}
	responseBool := false
	resp, err := impl.casbinClient.DeleteRoleForUser(context.Background(), in)
	if err != nil {
		return responseBool, err
	}
	responseBool = resp.Resp
	return responseBool, nil
}
func (impl *CasbinServiceImpl) GetRolesForUser(user string) ([]string, error) {
	in := &client.GetRolesForUserRequest{
		User: user,
	}
	resp, err := impl.casbinClient.GetRolesForUser(context.Background(), in)
	if err != nil {
		return nil, err
	}
	return resp.Roles, nil
}
func (impl *CasbinServiceImpl) GetUserByRole(role string) ([]string, error) {
	in := &client.GetUserByRoleRequest{
		Role: role,
	}
	resp, err := impl.casbinClient.GetUserByRole(context.Background(), in)
	if err != nil {
		return nil, err
	}
	return resp.Users, nil
}
func (impl *CasbinServiceImpl) RemovePoliciesByRole(role string) (bool, error) {
	in := &client.RemovePoliciesByRoleRequest{
		Role: role,
	}
	responseBool := false
	resp, err := impl.casbinClient.RemovePoliciesByRole(context.Background(), in)
	if err != nil {
		return responseBool, err
	}
	responseBool = resp.Resp
	return responseBool, nil
}
func (impl *CasbinServiceImpl) RemovePoliciesByRoles(roles []string) (bool, error) {
	in := &client.RemovePoliciesByRolesRequest{
		Roles: roles,
	}
	responseBool := false
	resp, err := impl.casbinClient.RemovePoliciesByRoles(context.Background(), in)
	if err != nil {
		return responseBool, err
	}
	responseBool = resp.Resp
	return responseBool, nil
}
