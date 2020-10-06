// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/companieshouse/chs-streaming-api/consumer (interfaces: Interface,FactoryInterface)

package mock_consumer

import (
	sarama "github.com/Shopify/sarama"
	consumer "github.com/companieshouse/chs-streaming-api-frontend/consumer"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockInterface is a mock of Interface interface
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// ConsumePartition mocks base method
func (m *MockInterface) ConsumePartition(arg0 int32, arg1 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConsumePartition", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ConsumePartition indicates an expected call of ConsumePartition
func (mr *MockInterfaceMockRecorder) ConsumePartition(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConsumePartition", reflect.TypeOf((*MockInterface)(nil).ConsumePartition), arg0, arg1)
}

// Messages mocks base method
func (m *MockInterface) Messages() <-chan *sarama.ConsumerMessage {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Messages")
	ret0, _ := ret[0].(<-chan *sarama.ConsumerMessage)
	return ret0
}

// Messages indicates an expected call of Messages
func (mr *MockInterfaceMockRecorder) Messages() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Messages", reflect.TypeOf((*MockInterface)(nil).Messages))
}

// MockFactoryInterface is a mock of FactoryInterface interface
type MockFactoryInterface struct {
	ctrl     *gomock.Controller
	recorder *MockFactoryInterfaceMockRecorder
}

// MockFactoryInterfaceMockRecorder is the mock recorder for MockFactoryInterface
type MockFactoryInterfaceMockRecorder struct {
	mock *MockFactoryInterface
}

// NewMockFactoryInterface creates a new mock instance
func NewMockFactoryInterface(ctrl *gomock.Controller) *MockFactoryInterface {
	mock := &MockFactoryInterface{ctrl: ctrl}
	mock.recorder = &MockFactoryInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockFactoryInterface) EXPECT() *MockFactoryInterfaceMockRecorder {
	return m.recorder
}

// Initialise mocks base method
func (m *MockFactoryInterface) Initialise(arg0 string, arg1 []string) consumer.Interface {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Initialise", arg0, arg1)
	ret0, _ := ret[0].(consumer.Interface)
	return ret0
}

// Initialise indicates an expected call of Initialise
func (mr *MockFactoryInterfaceMockRecorder) Initialise(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Initialise", reflect.TypeOf((*MockFactoryInterface)(nil).Initialise), arg0, arg1)
}
