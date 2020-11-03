package mocking

import (
	"github.com/companieshouse/chs.go/log"
	"github.com/stretchr/testify/mock"
	"net/http"
)

type MockLogger struct {
	mock.Mock
}

func (l *MockLogger) Error(err error, data ...log.Data) {
	l.Called(err, data)
}

func (l *MockLogger) Info(msg string, data ...log.Data) {
	l.Called(msg, data)
}

func (l *MockLogger) InfoR(req *http.Request, message string, data ...log.Data) {
	l.Called(req, message, data)
}

func (l *MockLogger) Debug(msg string, data ...log.Data) {
	l.Called(msg, data)
}
