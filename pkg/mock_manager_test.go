// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/asecurityteam/benthos-ext/pkg (interfaces: Manager)

// Package benthosx is a generated GoMock package.
package benthosx

import (
	types "github.com/Jeffail/benthos/lib/types"
	gomock "github.com/golang/mock/gomock"
	http "net/http"
	reflect "reflect"
)

// MockManager is a mock of Manager interface
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockManager) EXPECT() *MockManagerMockRecorder {
	return m.recorder
}

// AddMessageID mocks base method
func (m *MockManager) AddMessageID(arg0 types.Message) types.Message {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddMessageID", arg0)
	ret0, _ := ret[0].(types.Message)
	return ret0
}

// AddMessageID indicates an expected call of AddMessageID
func (mr *MockManagerMockRecorder) AddMessageID(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddMessageID", reflect.TypeOf((*MockManager)(nil).AddMessageID), arg0)
}

// GetCache mocks base method
func (m *MockManager) GetCache(arg0 string) (types.Cache, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCache", arg0)
	ret0, _ := ret[0].(types.Cache)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCache indicates an expected call of GetCache
func (mr *MockManagerMockRecorder) GetCache(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCache", reflect.TypeOf((*MockManager)(nil).GetCache), arg0)
}

// GetCondition mocks base method
func (m *MockManager) GetCondition(arg0 string) (types.Condition, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCondition", arg0)
	ret0, _ := ret[0].(types.Condition)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetCondition indicates an expected call of GetCondition
func (mr *MockManagerMockRecorder) GetCondition(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCondition", reflect.TypeOf((*MockManager)(nil).GetCondition), arg0)
}

// GetMessageResponse mocks base method
func (m *MockManager) GetMessageResponse(arg0 types.Message) (interface{}, bool) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMessageResponse", arg0)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(bool)
	return ret0, ret1
}

// GetMessageResponse indicates an expected call of GetMessageResponse
func (mr *MockManagerMockRecorder) GetMessageResponse(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMessageResponse", reflect.TypeOf((*MockManager)(nil).GetMessageResponse), arg0)
}

// GetPipe mocks base method
func (m *MockManager) GetPipe(arg0 string) (<-chan types.Transaction, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPipe", arg0)
	ret0, _ := ret[0].(<-chan types.Transaction)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPipe indicates an expected call of GetPipe
func (mr *MockManagerMockRecorder) GetPipe(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPipe", reflect.TypeOf((*MockManager)(nil).GetPipe), arg0)
}

// GetRateLimit mocks base method
func (m *MockManager) GetRateLimit(arg0 string) (types.RateLimit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRateLimit", arg0)
	ret0, _ := ret[0].(types.RateLimit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRateLimit indicates an expected call of GetRateLimit
func (mr *MockManagerMockRecorder) GetRateLimit(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRateLimit", reflect.TypeOf((*MockManager)(nil).GetRateLimit), arg0)
}

// RegisterEndpoint mocks base method
func (m *MockManager) RegisterEndpoint(arg0, arg1 string, arg2 http.HandlerFunc) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RegisterEndpoint", arg0, arg1, arg2)
}

// RegisterEndpoint indicates an expected call of RegisterEndpoint
func (mr *MockManagerMockRecorder) RegisterEndpoint(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RegisterEndpoint", reflect.TypeOf((*MockManager)(nil).RegisterEndpoint), arg0, arg1, arg2)
}

// SetMessageResponse mocks base method
func (m *MockManager) SetMessageResponse(arg0 types.Message, arg1 interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetMessageResponse", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetMessageResponse indicates an expected call of SetMessageResponse
func (mr *MockManagerMockRecorder) SetMessageResponse(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMessageResponse", reflect.TypeOf((*MockManager)(nil).SetMessageResponse), arg0, arg1)
}

// SetPipe mocks base method
func (m *MockManager) SetPipe(arg0 string, arg1 <-chan types.Transaction) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPipe", arg0, arg1)
}

// SetPipe indicates an expected call of SetPipe
func (mr *MockManagerMockRecorder) SetPipe(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPipe", reflect.TypeOf((*MockManager)(nil).SetPipe), arg0, arg1)
}

// UnsetPipe mocks base method
func (m *MockManager) UnsetPipe(arg0 string, arg1 <-chan types.Transaction) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UnsetPipe", arg0, arg1)
}

// UnsetPipe indicates an expected call of UnsetPipe
func (mr *MockManagerMockRecorder) UnsetPipe(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnsetPipe", reflect.TypeOf((*MockManager)(nil).UnsetPipe), arg0, arg1)
}
