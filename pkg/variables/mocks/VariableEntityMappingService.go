// Code generated by mockery v2.33.1. DO NOT EDIT.

package mocks

import (
	repository "github.com/devtron-labs/devtron/pkg/variables/repository"
	mock "github.com/stretchr/testify/mock"
)

// VariableEntityMappingService is an autogenerated mock type for the VariableEntityMappingService type
type VariableEntityMappingService struct {
	mock.Mock
}

// DeleteMappingsForEntities provides a mock function with given fields: entities, userId
func (_m *VariableEntityMappingService) DeleteMappingsForEntities(entities []repository.Entity, userId int32) error {
	ret := _m.Called(entities, userId)

	var r0 error
	if rf, ok := ret.Get(0).(func([]repository.Entity, int32) error); ok {
		r0 = rf(entities, userId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetAllMappingsForEntities provides a mock function with given fields: entities
func (_m *VariableEntityMappingService) GetAllMappingsForEntities(entities []repository.Entity) (map[repository.Entity][]string, error) {
	ret := _m.Called(entities)

	var r0 map[repository.Entity][]string
	var r1 error
	if rf, ok := ret.Get(0).(func([]repository.Entity) (map[repository.Entity][]string, error)); ok {
		return rf(entities)
	}
	if rf, ok := ret.Get(0).(func([]repository.Entity) map[repository.Entity][]string); ok {
		r0 = rf(entities)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[repository.Entity][]string)
		}
	}

	if rf, ok := ret.Get(1).(func([]repository.Entity) error); ok {
		r1 = rf(entities)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateVariablesForEntity provides a mock function with given fields: variableNames, entity, userId
func (_m *VariableEntityMappingService) UpdateVariablesForEntity(variableNames []string, entity repository.Entity, userId int32) error {
	ret := _m.Called(variableNames, entity, userId)

	var r0 error
	if rf, ok := ret.Get(0).(func([]string, repository.Entity, int32) error); ok {
		r0 = rf(variableNames, entity, userId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewVariableEntityMappingService creates a new instance of VariableEntityMappingService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewVariableEntityMappingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *VariableEntityMappingService {
	mock := &VariableEntityMappingService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
