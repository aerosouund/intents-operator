// Code generated by MockGen. DO NOT EDIT.
// Source: ../istiopolicy/policy_manager.go

// Package intentsreconcilersmocks is a generated GoMock package.
package intentsreconcilersmocks

import (
	context "context"
	reflect "reflect"

	v2alpha1 "github.com/otterize/intents-operator/src/operator/api/v2alpha1"
	gomock "go.uber.org/mock/gomock"
)

// MockPolicyManager is a mock of PolicyManager interface.
type MockPolicyManager struct {
	ctrl     *gomock.Controller
	recorder *MockPolicyManagerMockRecorder
}

// MockPolicyManagerMockRecorder is the mock recorder for MockPolicyManager.
type MockPolicyManagerMockRecorder struct {
	mock *MockPolicyManager
}

// NewMockPolicyManager creates a new mock instance.
func NewMockPolicyManager(ctrl *gomock.Controller) *MockPolicyManager {
	mock := &MockPolicyManager{ctrl: ctrl}
	mock.recorder = &MockPolicyManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPolicyManager) EXPECT() *MockPolicyManagerMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockPolicyManager) Create(ctx context.Context, clientIntents *v2alpha1.ClientIntents, clientServiceAccount string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, clientIntents, clientServiceAccount)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockPolicyManagerMockRecorder) Create(ctx, clientIntents, clientServiceAccount interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockPolicyManager)(nil).Create), ctx, clientIntents, clientServiceAccount)
}

// DeleteAll mocks base method.
func (m *MockPolicyManager) DeleteAll(ctx context.Context, clientIntents *v2alpha1.ClientIntents) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAll", ctx, clientIntents)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAll indicates an expected call of DeleteAll.
func (mr *MockPolicyManagerMockRecorder) DeleteAll(ctx, clientIntents interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAll", reflect.TypeOf((*MockPolicyManager)(nil).DeleteAll), ctx, clientIntents)
}

// RemoveDeprecatedPoliciesForClient mocks base method.
func (m *MockPolicyManager) RemoveDeprecatedPoliciesForClient(ctx context.Context, clientIntents *v2alpha1.ClientIntents) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveDeprecatedPoliciesForClient", ctx, clientIntents)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveDeprecatedPoliciesForClient indicates an expected call of RemoveDeprecatedPoliciesForClient.
func (mr *MockPolicyManagerMockRecorder) RemoveDeprecatedPoliciesForClient(ctx, clientIntents interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveDeprecatedPoliciesForClient", reflect.TypeOf((*MockPolicyManager)(nil).RemoveDeprecatedPoliciesForClient), ctx, clientIntents)
}

// UpdateIntentsStatus mocks base method.
func (m *MockPolicyManager) UpdateIntentsStatus(ctx context.Context, clientIntents *v2alpha1.ClientIntents, clientServiceAccount string, missingSideCar bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateIntentsStatus", ctx, clientIntents, clientServiceAccount, missingSideCar)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateIntentsStatus indicates an expected call of UpdateIntentsStatus.
func (mr *MockPolicyManagerMockRecorder) UpdateIntentsStatus(ctx, clientIntents, clientServiceAccount, missingSideCar interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateIntentsStatus", reflect.TypeOf((*MockPolicyManager)(nil).UpdateIntentsStatus), ctx, clientIntents, clientServiceAccount, missingSideCar)
}

// UpdateServerSidecar mocks base method.
func (m *MockPolicyManager) UpdateServerSidecar(ctx context.Context, clientIntents *v2alpha1.ClientIntents, serverName string, missingSideCar bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateServerSidecar", ctx, clientIntents, serverName, missingSideCar)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateServerSidecar indicates an expected call of UpdateServerSidecar.
func (mr *MockPolicyManagerMockRecorder) UpdateServerSidecar(ctx, clientIntents, serverName, missingSideCar interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateServerSidecar", reflect.TypeOf((*MockPolicyManager)(nil).UpdateServerSidecar), ctx, clientIntents, serverName, missingSideCar)
}
