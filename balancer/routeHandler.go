package balancer

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ReverseProxyHandler struct {
	TargetAddress url.URL
}

func (router *ReverseProxyHandler) Handle(writer http.ResponseWriter, request *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(&router.TargetAddress)
	proxy.ServeHTTP(writer, request)
}
