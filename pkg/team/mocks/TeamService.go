// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	team "github.com/devtron-labs/devtron/pkg/team"
	mock "github.com/stretchr/testify/mock"
)

// TeamService is an autogenerated mock type for the TeamService type
type TeamService struct {
	mock.Mock
}

// Create provides a mock function with given fields: request
func (_m *TeamService) Create(request *team.TeamRequest) (*team.TeamRequest, error) {
	ret := _m.Called(request)

	var r0 *team.TeamRequest
	var r1 error
	if rf, ok := ret.Get(0).(func(*team.TeamRequest) (*team.TeamRequest, error)); ok {
		return rf(request)
	}
	if rf, ok := ret.Get(0).(func(*team.TeamRequest) *team.TeamRequest); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*team.TeamRequest)
		}
	}

	if rf, ok := ret.Get(1).(func(*team.TeamRequest) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: request
func (_m *TeamService) Delete(request *team.TeamRequest) error {
	ret := _m.Called(request)

	var r0 error
	if rf, ok := ret.Get(0).(func(*team.TeamRequest) error); ok {
		r0 = rf(request)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FetchAllActive provides a mock function with given fields:
func (_m *TeamService) FetchAllActive() ([]team.TeamRequest, error) {
	ret := _m.Called()

	var r0 []team.TeamRequest
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]team.TeamRequest, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []team.TeamRequest); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]team.TeamRequest)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FetchForAutocomplete provides a mock function with given fields:
func (_m *TeamService) FetchForAutocomplete() ([]team.TeamRequest, error) {
	ret := _m.Called()

	var r0 []team.TeamRequest
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]team.TeamRequest, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []team.TeamRequest); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]team.TeamRequest)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FetchOne provides a mock function with given fields: id
func (_m *TeamService) FetchOne(id int) (*team.TeamRequest, error) {
	ret := _m.Called(id)

	var r0 *team.TeamRequest
	var r1 error
	if rf, ok := ret.Get(0).(func(int) (*team.TeamRequest, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(int) *team.TeamRequest); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*team.TeamRequest)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByIds provides a mock function with given fields: ids
func (_m *TeamService) FindByIds(ids []*int) ([]*team.TeamBean, error) {
	ret := _m.Called(ids)

	var r0 []*team.TeamBean
	var r1 error
	if rf, ok := ret.Get(0).(func([]*int) ([]*team.TeamBean, error)); ok {
		return rf(ids)
	}
	if rf, ok := ret.Get(0).(func([]*int) []*team.TeamBean); ok {
		r0 = rf(ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*team.TeamBean)
		}
	}

	if rf, ok := ret.Get(1).(func([]*int) error); ok {
		r1 = rf(ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByTeamName provides a mock function with given fields: teamName
func (_m *TeamService) FindByTeamName(teamName string) (*team.TeamRequest, error) {
	ret := _m.Called(teamName)

	var r0 *team.TeamRequest
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*team.TeamRequest, error)); ok {
		return rf(teamName)
	}
	if rf, ok := ret.Get(0).(func(string) *team.TeamRequest); ok {
		r0 = rf(teamName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*team.TeamRequest)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(teamName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: request
func (_m *TeamService) Update(request *team.TeamRequest) (*team.TeamRequest, error) {
	ret := _m.Called(request)

	var r0 *team.TeamRequest
	var r1 error
	if rf, ok := ret.Get(0).(func(*team.TeamRequest) (*team.TeamRequest, error)); ok {
		return rf(request)
	}
	if rf, ok := ret.Get(0).(func(*team.TeamRequest) *team.TeamRequest); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*team.TeamRequest)
		}
	}

	if rf, ok := ret.Get(1).(func(*team.TeamRequest) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewTeamService interface {
	mock.TestingT
	Cleanup(func())
}

// NewTeamService creates a new instance of TeamService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTeamService(t mockConstructorTestingTNewTeamService) *TeamService {
	mock := &TeamService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
