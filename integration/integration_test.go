package integration_test

import (
	. "github.com/Alcereo/ordinator/pkg/context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var server *http.Server

var _ = BeforeSuite(func() {
	context := NewContext()

	cacheAdapterIdentifier := "main-adapter"
	context.SetupCache([]CacheAdapter{
		{
			Identifier:             cacheAdapterIdentifier,
			Type:                   GoCache,
			ExpirationTimeHours:    1,
			EvictScheduleTimeHours: 1,
		},
	})
	context.SetupRouters([]Router{
		{
			Type:      ReverseProxy,
			Pattern:   "/api/v1/",
			TargetUrl: "http://localhost:8081/",
			Filters: []Filter{
				{
					Type:                   SessionFilter,
					Name:                   "Test session filter",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					CookieDomain:           "localhost",
					CookiePath:             "/",
					CookieName:             "session",
					CookieTTLHours:         24,
					CookieRenewBeforeHours: 2,
				},
				{
					Type:     LogFilter,
					Name:     "Simple request log",
					Template: "METHOD:{{.Request.Method}} PATH:{{.Request.URL}} SESSION_ID:{{(.Request.Context.Value \"SessionContextKey\").Id}}",
				},
			},
		},
	}, GoogleSecret{})
	server = context.BuildServer(8080)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			Fail(err.Error())
		}
	}()
})

var _ = AfterSuite(func() {
	err := server.Close()
	if err != nil {
		Fail(err.Error())
	}
})
