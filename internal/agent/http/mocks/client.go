// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/agent/http/client.go

// Package mock_http is a generated GoMock package.
package mock_http

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockIClient is a mock of IClient interface.
type MockIClient struct {
	ctrl     *gomock.Controller
	recorder *MockIClientMockRecorder
}

// MockIClientMockRecorder is the mock recorder for MockIClient.
type MockIClientMockRecorder struct {
	mock *MockIClient
}

// NewMockIClient creates a new mock instance.
func NewMockIClient(ctrl *gomock.Controller) *MockIClient {
	mock := &MockIClient{ctrl: ctrl}
	mock.recorder = &MockIClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIClient) EXPECT() *MockIClientMockRecorder {
	return m.recorder
}

// Do mocks base method.
func (m *MockIClient) Do(req *http.Request) (*http.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Do", req)
	ret0, _ := ret[0].(*http.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Do indicates an expected call of Do.
func (mr *MockIClientMockRecorder) Do(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Do", reflect.TypeOf((*MockIClient)(nil).Do), req)
}
