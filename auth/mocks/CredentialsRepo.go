// Code generated by mockery v2.52.2. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// CredentialsRepo is an autogenerated mock type for the CredentialsRepo type
type CredentialsRepo struct {
	mock.Mock
}

// GetCredentials provides a mock function with given fields: ctx, login
func (_m *CredentialsRepo) GetCredentials(ctx context.Context, login string) (string, bool, error) {
	ret := _m.Called(ctx, login)

	if len(ret) == 0 {
		panic("no return value specified for GetCredentials")
	}

	var r0 string
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, bool, error)); ok {
		return rf(ctx, login)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, login)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) bool); ok {
		r1 = rf(ctx, login)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(context.Context, string) error); ok {
		r2 = rf(ctx, login)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// SetCredentials provides a mock function with given fields: ctx, login, password
func (_m *CredentialsRepo) SetCredentials(ctx context.Context, login string, password string) error {
	ret := _m.Called(ctx, login, password)

	if len(ret) == 0 {
		panic("no return value specified for SetCredentials")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, login, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewCredentialsRepo creates a new instance of CredentialsRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCredentialsRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *CredentialsRepo {
	mock := &CredentialsRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
