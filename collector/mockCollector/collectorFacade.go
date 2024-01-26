// Code generated by MockGen. DO NOT EDIT.
// Source: collectorFacade.go

// Package mock_collector is a generated GoMock package.
package mock_collector

import (
	reflect "reflect"

	collectorType "github.com/LL-res/CRM/collector/collectorType"
	key "github.com/LL-res/CRM/common/key"
	gomock "github.com/golang/mock/gomock"
)

// MockCollectorFacade is a mock of CollectorFacade interface.
type MockCollectorFacade struct {
	ctrl     *gomock.Controller
	recorder *MockCollectorFacadeMockRecorder
}

// MockCollectorFacadeMockRecorder is the mock recorder for MockCollectorFacade.
type MockCollectorFacadeMockRecorder struct {
	mock *MockCollectorFacade
}

// NewMockCollectorFacade creates a new mock instance.
func NewMockCollectorFacade(ctrl *gomock.Controller) *MockCollectorFacade {
	mock := &MockCollectorFacade{ctrl: ctrl}
	mock.recorder = &MockCollectorFacadeMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCollectorFacade) EXPECT() *MockCollectorFacadeMockRecorder {
	return m.recorder
}

// CreateCollector mocks base method.
func (m *MockCollectorFacade) CreateCollector(nmk key.NoModelKey) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CreateCollector", nmk)
}

// CreateCollector indicates an expected call of CreateCollector.
func (mr *MockCollectorFacadeMockRecorder) CreateCollector(nmk interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateCollector", reflect.TypeOf((*MockCollectorFacade)(nil).CreateCollector), nmk)
}

// DeleteCollector mocks base method.
func (m *MockCollectorFacade) DeleteCollector(nmk key.NoModelKey) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "DeleteCollector", nmk)
}

// DeleteCollector indicates an expected call of DeleteCollector.
func (mr *MockCollectorFacadeMockRecorder) DeleteCollector(nmk interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteCollector", reflect.TypeOf((*MockCollectorFacade)(nil).DeleteCollector), nmk)
}

// GetCapFromCollector mocks base method.
func (m *MockCollectorFacade) GetCapFromCollector(nmk key.NoModelKey) int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCapFromCollector", nmk)
	ret0, _ := ret[0].(int)
	return ret0
}

// GetCapFromCollector indicates an expected call of GetCapFromCollector.
func (mr *MockCollectorFacadeMockRecorder) GetCapFromCollector(nmk interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCapFromCollector", reflect.TypeOf((*MockCollectorFacade)(nil).GetCapFromCollector), nmk)
}

// GetCollectorKeySet mocks base method.
func (m *MockCollectorFacade) GetCollectorKeySet() map[key.NoModelKey]struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCollectorKeySet")
	ret0, _ := ret[0].(map[key.NoModelKey]struct{})
	return ret0
}

// GetCollectorKeySet indicates an expected call of GetCollectorKeySet.
func (mr *MockCollectorFacadeMockRecorder) GetCollectorKeySet() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCollectorKeySet", reflect.TypeOf((*MockCollectorFacade)(nil).GetCollectorKeySet))
}

// GetMetricFromCollector mocks base method.
func (m *MockCollectorFacade) GetMetricFromCollector(nmk key.NoModelKey, length int) []float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetricFromCollector", nmk, length)
	ret0, _ := ret[0].([]float64)
	return ret0
}

// GetMetricFromCollector indicates an expected call of GetMetricFromCollector.
func (mr *MockCollectorFacadeMockRecorder) GetMetricFromCollector(nmk, length interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetricFromCollector", reflect.TypeOf((*MockCollectorFacade)(nil).GetMetricFromCollector), nmk, length)
}

// Init mocks base method.
func (m *MockCollectorFacade) Init(config collectorType.CollectorConfig) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Init", config)
	ret0, _ := ret[0].(error)
	return ret0
}

// Init indicates an expected call of Init.
func (mr *MockCollectorFacadeMockRecorder) Init(config interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Init", reflect.TypeOf((*MockCollectorFacade)(nil).Init), config)
}

// WaitToGetMetric mocks base method.
func (m *MockCollectorFacade) WaitToGetMetric(nmk key.NoModelKey, length int) []float64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitToGetMetric", nmk, length)
	ret0, _ := ret[0].([]float64)
	return ret0
}

// WaitToGetMetric indicates an expected call of WaitToGetMetric.
func (mr *MockCollectorFacadeMockRecorder) WaitToGetMetric(nmk, length interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitToGetMetric", reflect.TypeOf((*MockCollectorFacade)(nil).WaitToGetMetric), nmk, length)
}