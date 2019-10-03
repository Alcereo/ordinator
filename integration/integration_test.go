package integration_test

import (
	. "github.com/Alcereo/ordinator/integration/utils"
	. "github.com/Alcereo/ordinator/pkg/context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var server *http.Server
var googleApiStub *httptest.Server
var resourceStub *httptest.Server

var _ = BeforeSuite(func() {

	googleApiStub = createGoogleApiStub()
	resourceStub = createResourceServiceStub()
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
			Type:                   GoogleOauth2Authorization,
			Pattern:                "/authentication/google",
			CacheAdapterIdentifier: cacheAdapterIdentifier,
			SuccessLoginUrl:        "/api/v2/",
			AccessTokenRequestUrl:  googleApiStub.URL + "/oauth2/v4/token",
			UserInfoRequestUrl:     googleApiStub.URL + "/oauth2/v3/userinfo",
			Filters: []Filter{
				{
					Type:                   SessionFilter,
					Name:                   "Auth session filter",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					CookieDomain:           "localhost",
					CookiePath:             "/",
					CookieName:             "session",
					CookieTTLHours:         24,
					CookieRenewBeforeHours: 2,
				},
			},
		},
		{
			Type:      ReverseProxy,
			Pattern:   "/api/v2/",
			TargetUrl: resourceStub.URL,
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
		{
			Type:      ReverseProxy,
			Pattern:   "/api/v1/",
			TargetUrl: resourceStub.URL,
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
		defer GinkgoRecover()
		_ = server.ListenAndServe()
	}()
})

func createResourceServiceStub() *httptest.Server {
	return CreateServiceStub([]RequestMock{
		{
			Request: Request{
				Method: "GET",
				Url:    "/api/v2/resource",
			},
			Response: Response{
				Status:  200,
				Headers: nil,
				Body: JsonMap{
					"status":  "OK",
					"service": "resource",
					"version": "v2",
				},
			},
		},
		{
			Request: Request{
				Method: "GET",
				Url:    "/api/v1/resource",
			},
			Response: Response{
				Status:  200,
				Headers: nil,
				Body: JsonMap{
					"status":  "OK",
					"service": "resource",
					"version": "v1",
				},
			},
		},
	})
}

func createGoogleApiStub() *httptest.Server {
	return CreateServiceStub([]RequestMock{
		{
			Request: Request{
				Method: "POST",
				Url:    "/oauth2/v4/token",
				Headers: []Header{
					{
						Name:   "Content-Type",
						Regexp: "application/x-www-form-urlencoded",
					},
				},
				Body: []BodyCheck{
					URLPropsBody{
						Props: map[string]string{
							"code":          "google-auth-code",
							"client_id":     "google-client-id-1",
							"client_secret": "google-client-secret",
							"redirect_uri":  "http://localhost:8080/authentication/google",
							"grant_type":    "authorization_code",
						},
					},
				},
			},
			Response: Response{
				Status: 200,
				Headers: map[string]string{
					"Content-Type": "application/json;charset=UTF-8",
				},
				Body: JsonMap{
					"access_token":  "access-token-1",
					"refresh_token": "refresh-token-1",
					"expires_in":    9999,
					"token_type":    "authorization_code",
				},
			},
		},
	})
}

var _ = AfterSuite(func() {
	err := server.Close()
	if err != nil {
		Fail(err.Error())
	}
	resourceStub.Close()
	googleApiStub.Close()
})
