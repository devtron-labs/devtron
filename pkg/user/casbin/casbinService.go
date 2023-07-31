package casbin

import (
	"context"
	"github.com/devtron-labs/devtron/pkg/user/casbin/client"
	"go.uber.org/zap"
)

type CasbinService interface {
	AddPolicy(policies []Policy) error
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

func (impl *CasbinServiceImpl) AddPolicy(policies []Policy) error {
	convertedPolicies := make([]*client.Policy, 0, len(policies))
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
	resp, err := impl.casbinClient.AddPolicy(context.Background(), in)
	if err != nil {
		return err
	}
	if resp != nil && len(resp.FailedPolicies) > 0 {
		impl.logger.Errorw("error in adding all policies", "err", err, "failedPolicies", resp.FailedPolicies, "error message", resp.ErrorMessage)
	}
	return nil
}

func (impl *CasbinServiceImpl) LoadPolicy() {
	in := &client.EmptyObj{}
	_, _ = impl.casbinClient.LoadPolicy(context.Background(), in)
	return
}

func (impl *CasbinServiceImpl) RemovePolicy(policies []Policy) ([]Policy, error) {
	convertedPolicies := make([]*client.Policy, 0, len(policies))
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

	resp, err := impl.casbinClient.RemovePolicy(context.Background(), in)
	if err != nil {
		impl.logger.Errorw("error in removing policies", "err", err, "policies", resp.Policies)
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
