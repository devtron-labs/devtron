// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	repository "github.com/devtron-labs/devtron/internal/sql/repository"
	mock "github.com/stretchr/testify/mock"
)

// WebhookNotificationRepository is an autogenerated mock type for the WebhookNotificationRepository type
type WebhookNotificationRepository struct {
	mock.Mock
}

// FindAll provides a mock function with given fields:
func (_m *WebhookNotificationRepository) FindAll() ([]repository.WebhookConfig, error) {
	ret := _m.Called()

	var r0 []repository.WebhookConfig
	var r1 error
	if rf, ok := ret.Get(0).(func() ([]repository.WebhookConfig, error)); ok {
		return rf()
	}
	if rf, ok := ret.Get(0).(func() []repository.WebhookConfig); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]repository.WebhookConfig)
		}
	}

	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByIds provides a mock function with given fields: ids
func (_m *WebhookNotificationRepository) FindByIds(ids []*int) ([]*repository.WebhookConfig, error) {
	ret := _m.Called(ids)

	var r0 []*repository.WebhookConfig
	var r1 error
	if rf, ok := ret.Get(0).(func([]*int) ([]*repository.WebhookConfig, error)); ok {
		return rf(ids)
	}
	if rf, ok := ret.Get(0).(func([]*int) []*repository.WebhookConfig); ok {
		r0 = rf(ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*repository.WebhookConfig)
		}
	}

	if rf, ok := ret.Get(1).(func([]*int) error); ok {
		r1 = rf(ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindByName provides a mock function with given fields: value
func (_m *WebhookNotificationRepository) FindByName(value string) ([]repository.WebhookConfig, error) {
	ret := _m.Called(value)

	var r0 []repository.WebhookConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(string) ([]repository.WebhookConfig, error)); ok {
		return rf(value)
	}
	if rf, ok := ret.Get(0).(func(string) []repository.WebhookConfig); ok {
		r0 = rf(value)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]repository.WebhookConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(value)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindOne provides a mock function with given fields: id
func (_m *WebhookNotificationRepository) FindOne(id int) (*repository.WebhookConfig, error) {
	ret := _m.Called(id)

	var r0 *repository.WebhookConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(int) (*repository.WebhookConfig, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(int) *repository.WebhookConfig); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*repository.WebhookConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MarkWebhookConfigDeleted provides a mock function with given fields: webhookConfig
func (_m *WebhookNotificationRepository) MarkWebhookConfigDeleted(webhookConfig *repository.WebhookConfig) error {
	ret := _m.Called(webhookConfig)

	var r0 error
	if rf, ok := ret.Get(0).(func(*repository.WebhookConfig) error); ok {
		r0 = rf(webhookConfig)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveWebhookConfig provides a mock function with given fields: webhookConfig
func (_m *WebhookNotificationRepository) SaveWebhookConfig(webhookConfig *repository.WebhookConfig) (*repository.WebhookConfig, error) {
	ret := _m.Called(webhookConfig)

	var r0 *repository.WebhookConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(*repository.WebhookConfig) (*repository.WebhookConfig, error)); ok {
		return rf(webhookConfig)
	}
	if rf, ok := ret.Get(0).(func(*repository.WebhookConfig) *repository.WebhookConfig); ok {
		r0 = rf(webhookConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*repository.WebhookConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(*repository.WebhookConfig) error); ok {
		r1 = rf(webhookConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateWebhookConfig provides a mock function with given fields: webhookConfig
func (_m *WebhookNotificationRepository) UpdateWebhookConfig(webhookConfig *repository.WebhookConfig) (*repository.WebhookConfig, error) {
	ret := _m.Called(webhookConfig)

	var r0 *repository.WebhookConfig
	var r1 error
	if rf, ok := ret.Get(0).(func(*repository.WebhookConfig) (*repository.WebhookConfig, error)); ok {
		return rf(webhookConfig)
	}
	if rf, ok := ret.Get(0).(func(*repository.WebhookConfig) *repository.WebhookConfig); ok {
		r0 = rf(webhookConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*repository.WebhookConfig)
		}
	}

	if rf, ok := ret.Get(1).(func(*repository.WebhookConfig) error); ok {
		r1 = rf(webhookConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewWebhookNotificationRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewWebhookNotificationRepository creates a new instance of WebhookNotificationRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewWebhookNotificationRepository(t mockConstructorTestingTNewWebhookNotificationRepository) *WebhookNotificationRepository {
	mock := &WebhookNotificationRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
