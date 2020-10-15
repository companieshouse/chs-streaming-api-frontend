package logger

import (
	"github.com/companieshouse/chs.go/log"
	"net/http"
)

type Logger interface {
	Error(err error, data ...log.Data)
	Info(msg string, data ...log.Data)
	InfoR(req *http.Request, message string, data ...log.Data)
}

var logger LoggerImpl

type LoggerImpl struct {
}

func NewLogger() *LoggerImpl {
	return &logger
}

func (l *LoggerImpl) Error(err error, data ...log.Data) {
	log.Error(err, data...)
}

func (l *LoggerImpl) Info(msg string, data ...log.Data) {
	log.Info(msg, data...)
}

func (l *LoggerImpl) InfoR(req *http.Request, message string, data ...log.Data) {
	log.InfoR(req, message, data...)
}

