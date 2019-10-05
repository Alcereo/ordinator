package common

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

type RequestHandler interface {
	Handle(log *logrus.Entry, writer http.ResponseWriter, request *http.Request)
}

type RequestChainedHandler interface {
	Handle(log *logrus.Entry, writer http.ResponseWriter, request *http.Request)
	SetNext(handler RequestHandler)
}
