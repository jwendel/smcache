// Code generated by MockGen. DO NOT EDIT.
// Source: secret_api_wrapper.go

// Package mocks is a generated GoMock package.
package mocks

import (
	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	gomock "github.com/golang/mock/gomock"
	secretmanager0 "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
	reflect "reflect"
)

// MocksecretClient is a mock of secretClient interface
type MocksecretClient struct {
	ctrl     *gomock.Controller
	recorder *MocksecretClientMockRecorder
}

// MocksecretClientMockRecorder is the mock recorder for MocksecretClient
type MocksecretClientMockRecorder struct {
	mock *MocksecretClient
}

// NewMocksecretClient creates a new mock instance
func NewMocksecretClient(ctrl *gomock.Controller) *MocksecretClient {
	mock := &MocksecretClient{ctrl: ctrl}
	mock.recorder = &MocksecretClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MocksecretClient) EXPECT() *MocksecretClientMockRecorder {
	return m.recorder
}

// AccessSecretVersion mocks base method
func (m *MocksecretClient) AccessSecretVersion(req *secretmanager0.AccessSecretVersionRequest) (*secretmanager0.AccessSecretVersionResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AccessSecretVersion", req)
	ret0, _ := ret[0].(*secretmanager0.AccessSecretVersionResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AccessSecretVersion indicates an expected call of AccessSecretVersion
func (mr *MocksecretClientMockRecorder) AccessSecretVersion(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AccessSecretVersion", reflect.TypeOf((*MocksecretClient)(nil).AccessSecretVersion), req)
}

// ListSecretVersions mocks base method
func (m *MocksecretClient) ListSecretVersions(req *secretmanager0.ListSecretVersionsRequest) *secretmanager.SecretVersionIterator {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSecretVersions", req)
	ret0, _ := ret[0].(*secretmanager.SecretVersionIterator)
	return ret0
}

// ListSecretVersions indicates an expected call of ListSecretVersions
func (mr *MocksecretClientMockRecorder) ListSecretVersions(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSecretVersions", reflect.TypeOf((*MocksecretClient)(nil).ListSecretVersions), req)
}

// DestroySecretVersion mocks base method
func (m *MocksecretClient) DestroySecretVersion(req *secretmanager0.DestroySecretVersionRequest) (*secretmanager0.SecretVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DestroySecretVersion", req)
	ret0, _ := ret[0].(*secretmanager0.SecretVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DestroySecretVersion indicates an expected call of DestroySecretVersion
func (mr *MocksecretClientMockRecorder) DestroySecretVersion(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DestroySecretVersion", reflect.TypeOf((*MocksecretClient)(nil).DestroySecretVersion), req)
}

// CreateSecret mocks base method
func (m *MocksecretClient) CreateSecret(req *secretmanager0.CreateSecretRequest) (*secretmanager0.Secret, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateSecret", req)
	ret0, _ := ret[0].(*secretmanager0.Secret)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateSecret indicates an expected call of CreateSecret
func (mr *MocksecretClientMockRecorder) CreateSecret(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateSecret", reflect.TypeOf((*MocksecretClient)(nil).CreateSecret), req)
}

// AddSecretVersion mocks base method
func (m *MocksecretClient) AddSecretVersion(req *secretmanager0.AddSecretVersionRequest) (*secretmanager0.SecretVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddSecretVersion", req)
	ret0, _ := ret[0].(*secretmanager0.SecretVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddSecretVersion indicates an expected call of AddSecretVersion
func (mr *MocksecretClientMockRecorder) AddSecretVersion(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddSecretVersion", reflect.TypeOf((*MocksecretClient)(nil).AddSecretVersion), req)
}

// DeleteSecret mocks base method
func (m *MocksecretClient) DeleteSecret(req *secretmanager0.DeleteSecretRequest) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSecret", req)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteSecret indicates an expected call of DeleteSecret
func (mr *MocksecretClientMockRecorder) DeleteSecret(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSecret", reflect.TypeOf((*MocksecretClient)(nil).DeleteSecret), req)
}
