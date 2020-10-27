// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import context "context"
import grpc "google.golang.org/grpc"
import mock "github.com/stretchr/testify/mock"
import session "github.com/argoproj/argo-cd/pkg/apiclient/session"

// SessionServiceClient is an autogenerated mock type for the SessionServiceClient type
type SessionServiceClient struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in, opts
func (_m *SessionServiceClient) Create(ctx context.Context, in *session.SessionCreateRequest, opts ...grpc.CallOption) (*session.SessionResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *session.SessionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *session.SessionCreateRequest, ...grpc.CallOption) *session.SessionResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*session.SessionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *session.SessionCreateRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, in, opts
func (_m *SessionServiceClient) Delete(ctx context.Context, in *session.SessionDeleteRequest, opts ...grpc.CallOption) (*session.SessionResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *session.SessionResponse
	if rf, ok := ret.Get(0).(func(context.Context, *session.SessionDeleteRequest, ...grpc.CallOption) *session.SessionResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*session.SessionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *session.SessionDeleteRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
