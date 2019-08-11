package balancer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type RequestHandler interface {
	Handle(writer http.ResponseWriter, request *http.Request)
}

type RequestChainedHandler interface {
	Handle(writer http.ResponseWriter, request *http.Request)
	SetNext(handler RequestHandler)
}

type ReversiveProxyHandler struct {
	TargetAddress url.URL
}

func (router *ReversiveProxyHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	proxy:= httputil.NewSingleHostReverseProxy(&router.TargetAddress)
	proxy.ServeHTTP(writer, request)
}