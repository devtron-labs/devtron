// Code generated by MockGen. DO NOT EDIT.
// Source: internal/sql/repository/pipelineConfig/CiPipelineMaterial.go

// Package mock_pipelineConfig is a generated GoMock package.
package mocks

import (
	"github.com/devtron-labs/devtron/internal/sql/repository/pipelineConfig"
	reflect "reflect"

	pg "github.com/go-pg/pg"
	gomock "github.com/golang/mock/gomock"
)

// MockCiPipelineMaterialRepository is a mock of CiPipelineMaterialRepository interface.
type MockCiPipelineMaterialRepository struct {
	ctrl     *gomock.Controller
	recorder *MockCiPipelineMaterialRepositoryMockRecorder
}

// MockCiPipelineMaterialRepositoryMockRecorder is the mock recorder for MockCiPipelineMaterialRepository.
type MockCiPipelineMaterialRepositoryMockRecorder struct {
	mock *MockCiPipelineMaterialRepository
}

// NewMockCiPipelineMaterialRepository creates a new mock instance.
func NewMockCiPipelineMaterialRepository(ctrl *gomock.Controller) *MockCiPipelineMaterialRepository {
	mock := &MockCiPipelineMaterialRepository{ctrl: ctrl}
	mock.recorder = &MockCiPipelineMaterialRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCiPipelineMaterialRepository) EXPECT() *MockCiPipelineMaterialRepositoryMockRecorder {
	return m.recorder
}

// CheckRegexExistsForMaterial mocks base method.
func (m *MockCiPipelineMaterialRepository) CheckRegexExistsForMaterial(id int) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CheckRegexExistsForMaterial", id)
	ret0, _ := ret[0].(bool)
	return ret0
}

// CheckRegexExistsForMaterial indicates an expected call of CheckRegexExistsForMaterial.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) CheckRegexExistsForMaterial(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckRegexExistsForMaterial", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).CheckRegexExistsForMaterial), id)
}

// FindByCiPipelineIdsIn mocks base method.
func (m *MockCiPipelineMaterialRepository) FindByCiPipelineIdsIn(ids []int) ([]*pipelineConfig.CiPipelineMaterial, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByCiPipelineIdsIn", ids)
	ret0, _ := ret[0].([]*pipelineConfig.CiPipelineMaterial)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByCiPipelineIdsIn indicates an expected call of FindByCiPipelineIdsIn.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) FindByCiPipelineIdsIn(ids interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByCiPipelineIdsIn", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).FindByCiPipelineIdsIn), ids)
}

// GetById mocks base method.
func (m *MockCiPipelineMaterialRepository) GetById(id int) (*pipelineConfig.CiPipelineMaterial, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetById", id)
	ret0, _ := ret[0].(*pipelineConfig.CiPipelineMaterial)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetById indicates an expected call of GetById.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) GetById(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetById", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).GetById), id)
}

// GetByPipelineId mocks base method.
func (m *MockCiPipelineMaterialRepository) GetByPipelineId(id int) ([]*pipelineConfig.CiPipelineMaterial, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByPipelineId", id)
	ret0, _ := ret[0].([]*pipelineConfig.CiPipelineMaterial)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByPipelineId indicates an expected call of GetByPipelineId.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) GetByPipelineId(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByPipelineId", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).GetByPipelineId), id)
}

// GetByPipelineIdAndGitMaterialId mocks base method.
func (m *MockCiPipelineMaterialRepository) GetByPipelineIdAndGitMaterialId(id, gitMaterialId int) ([]*pipelineConfig.CiPipelineMaterial, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByPipelineIdAndGitMaterialId", id, gitMaterialId)
	ret0, _ := ret[0].([]*pipelineConfig.CiPipelineMaterial)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByPipelineIdAndGitMaterialId indicates an expected call of GetByPipelineIdAndGitMaterialId.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) GetByPipelineIdAndGitMaterialId(id, gitMaterialId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByPipelineIdAndGitMaterialId", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).GetByPipelineIdAndGitMaterialId), id, gitMaterialId)
}

// GetByPipelineIdForRegexAndFixed mocks base method.
func (m *MockCiPipelineMaterialRepository) GetByPipelineIdForRegexAndFixed(id int) ([]*pipelineConfig.CiPipelineMaterial, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByPipelineIdForRegexAndFixed", id)
	ret0, _ := ret[0].([]*pipelineConfig.CiPipelineMaterial)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByPipelineIdForRegexAndFixed indicates an expected call of GetByPipelineIdForRegexAndFixed.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) GetByPipelineIdForRegexAndFixed(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByPipelineIdForRegexAndFixed", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).GetByPipelineIdForRegexAndFixed), id)
}

// GetCheckoutPath mocks base method.
func (m *MockCiPipelineMaterialRepository) GetCheckoutPath(gitMaterialId int) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCheckoutPath", gitMaterialId)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCheckoutPath indicates an expected call of GetCheckoutPath.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) GetCheckoutPath(gitMaterialId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCheckoutPath", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).GetCheckoutPath), gitMaterialId)
}

// GetRegexByPipelineId mocks base method.
func (m *MockCiPipelineMaterialRepository) GetRegexByPipelineId(id int) ([]*pipelineConfig.CiPipelineMaterial, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRegexByPipelineId", id)
	ret0, _ := ret[0].([]*pipelineConfig.CiPipelineMaterial)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRegexByPipelineId indicates an expected call of GetRegexByPipelineId.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) GetRegexByPipelineId(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRegexByPipelineId", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).GetRegexByPipelineId), id)
}

// Save mocks base method.
func (m *MockCiPipelineMaterialRepository) Save(tx *pg.Tx, pipeline ...*pipelineConfig.CiPipelineMaterial) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{tx}
	for _, a := range pipeline {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Save", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) Save(tx interface{}, pipeline ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{tx}, pipeline...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).Save), varargs...)
}

// Update mocks base method.
func (m *MockCiPipelineMaterialRepository) Update(tx *pg.Tx, material ...*pipelineConfig.CiPipelineMaterial) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{tx}
	for _, a := range material {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Update", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) Update(tx interface{}, material ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{tx}, material...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).Update), varargs...)
}

// UpdateNotNull mocks base method.
func (m *MockCiPipelineMaterialRepository) UpdateNotNull(tx *pg.Tx, material ...*pipelineConfig.CiPipelineMaterial) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{tx}
	for _, a := range material {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateNotNull", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateNotNull indicates an expected call of UpdateNotNull.
func (mr *MockCiPipelineMaterialRepositoryMockRecorder) UpdateNotNull(tx interface{}, material ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{tx}, material...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateNotNull", reflect.TypeOf((*MockCiPipelineMaterialRepository)(nil).UpdateNotNull), varargs...)
}
