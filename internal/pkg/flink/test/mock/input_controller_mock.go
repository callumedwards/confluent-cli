// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/confluentinc/cli/internal/pkg/flink/internal/controller (interfaces: InputControllerInterface)

// Package mock is a generated GoMock package.
package mock

import (
	reflect "reflect"

	prompt "github.com/confluentinc/go-prompt"
	gomock "github.com/golang/mock/gomock"
)

// MockInputControllerInterface is a mock of InputControllerInterface interface.
type MockInputControllerInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInputControllerInterfaceMockRecorder
}

// MockInputControllerInterfaceMockRecorder is the mock recorder for MockInputControllerInterface.
type MockInputControllerInterfaceMockRecorder struct {
	mock *MockInputControllerInterface
}

// NewMockInputControllerInterface creates a new mock instance.
func NewMockInputControllerInterface(ctrl *gomock.Controller) *MockInputControllerInterface {
	mock := &MockInputControllerInterface{ctrl: ctrl}
	mock.recorder = &MockInputControllerInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInputControllerInterface) EXPECT() *MockInputControllerInterfaceMockRecorder {
	return m.recorder
}

// GetMaxCol mocks base method.
func (m *MockInputControllerInterface) GetMaxCol() (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMaxCol")
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMaxCol indicates an expected call of GetMaxCol.
func (mr *MockInputControllerInterfaceMockRecorder) GetMaxCol() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMaxCol", reflect.TypeOf((*MockInputControllerInterface)(nil).GetMaxCol))
}

// Prompt mocks base method.
func (m *MockInputControllerInterface) Prompt() *prompt.Prompt {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Prompt")
	ret0, _ := ret[0].(*prompt.Prompt)
	return ret0
}

// Prompt indicates an expected call of Prompt.
func (mr *MockInputControllerInterfaceMockRecorder) Prompt() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prompt", reflect.TypeOf((*MockInputControllerInterface)(nil).Prompt))
}

// RunInteractiveInput mocks base method.
func (m *MockInputControllerInterface) RunInteractiveInput() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RunInteractiveInput")
}

// RunInteractiveInput indicates an expected call of RunInteractiveInput.
func (mr *MockInputControllerInterfaceMockRecorder) RunInteractiveInput() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RunInteractiveInput", reflect.TypeOf((*MockInputControllerInterface)(nil).RunInteractiveInput))
}
