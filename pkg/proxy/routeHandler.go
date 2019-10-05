package proxy

import (
	"github.com/sirupsen/logrus"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type ReverseProxyHandler struct {
	TargetAddress url.URL
}

func (router *ReverseProxyHandler) Handle(log *logrus.Entry, writer http.ResponseWriter, request *http.Request) {
	proxy := httputil.NewSingleHostReverseProxy(&router.TargetAddress)
	proxy.ServeHTTP(writer, request)
}
