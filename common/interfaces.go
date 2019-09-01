package common

import "net/http"

type RequestHandler interface {
	Handle(writer http.ResponseWriter, request *http.Request)
}

type RequestChainedHandler interface {
	Handle(writer http.ResponseWriter, request *http.Request)
	SetNext(handler RequestHandler)
}
