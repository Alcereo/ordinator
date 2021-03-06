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

	//logrus.SetLevel(logrus.DebugLevel)
	//logrus.SetFormatter(&logrus.TextFormatter{
	//	ForceColors: true,
	//})
	//logrus.SetFormatter(&logrus.JSONFormatter{
	//	TimestampFormat:  "",
	//	DisableTimestamp: false,
	//	DataKey:          "",
	//	FieldMap:         nil,
	//	CallerPrettyfier: nil,
	//	PrettyPrint:      true,
	//})

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
			SuccessLoginUrl:        "/api/v2/resource",
			AccessTokenRequestUrl:  googleApiStub.URL + "/oauth2/v4/token",
			UserInfoRequestUrl:     googleApiStub.URL + "/oauth2/v3/userinfo",
			Filters: []Filter{
				{
					Type:                   SessionFilter,
					Name:                   "session filter for google auth",
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
					Name:     "log filter for: /api/v1/",
					Template: "METHOD:{{.Request.Method}} PATH:{{.Request.URL}} SESSION_ID:{{(.Request.Context.Value \"SessionContextKey\").Id}}",
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
					Name:                   "session filter for: /api/v2/",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					CookieDomain:           "localhost",
					CookiePath:             "/",
					CookieName:             "session",
					CookieTTLHours:         24,
					CookieRenewBeforeHours: 2,
				},
				{
					Type:                   UserAuthenticationFilter,
					Name:                   "auth filter for: /api/v2/",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					UserDataRequired:       true,
				},
				{
					Type:       CsrfFilter,
					Name:       "Csrf filter for v2",
					CsrfHeader: "X-CSRF-TOKEN",
					CsrfSafeMethods: []string{
						"GET",
					},
					CsrfEncryptorPrivateKey: "some-private-key",
				},
				{
					Type:     LogFilter,
					Name:     "log filter for: /api/v2/",
					Template: "METHOD:{{.Request.Method}} PATH:{{.Request.URL}} SESSION_ID:{{(.Request.Context.Value \"SessionContextKey\").Id}}",
				},
			},
		},
		{
			Type:      ReverseProxy,
			Pattern:   "/api/v3/",
			TargetUrl: resourceStub.URL,
			Filters: []Filter{
				{
					Type:                   SessionFilter,
					Name:                   "session filter for: /api/v3/",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					CookieDomain:           "localhost",
					CookiePath:             "/",
					CookieName:             "session",
					CookieTTLHours:         24,
					CookieRenewBeforeHours: 2,
				},
				{
					Type:                   UserAuthenticationFilter,
					Name:                   "auth filter for: /api/v3/",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					UserDataRequired:       true,
				},
				{
					Type:       CsrfFilter,
					Name:       "Csrf filter for v3",
					CsrfHeader: "X-CSRF-TOKEN",
					CsrfSafeMethods: []string{
						"GET",
					},
					CsrfEncryptorPrivateKey: "some-private-key",
				},
				{
					Type:     LogFilter,
					Name:     "log filter for: /api/v3/",
					Template: "METHOD:{{.Request.Method}} PATH:{{.Request.URL}} SESSION_ID:{{(.Request.Context.Value \"SessionContextKey\").Id}}",
				},
			},
		},
		{
			Type:      ReverseProxy,
			Pattern:   "/pages/work-page",
			TargetUrl: resourceStub.URL,
			Filters: []Filter{
				{
					Type:                   SessionFilter,
					Name:                   "session filter for main page",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					CookieDomain:           "localhost",
					CookiePath:             "/",
					CookieName:             "session",
					CookieTTLHours:         24,
					CookieRenewBeforeHours: 2,
				},
				{
					Type:                   UserAuthenticationFilter,
					Name:                   "auth filter for main page",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					UserDataRequired:       true,
					RedirectPage:           "/pages/login-page",
				},
			},
		},
		{
			Type:      ReverseProxy,
			Pattern:   "/pages/login-page",
			TargetUrl: resourceStub.URL,
			Filters: []Filter{
				{
					Type:                   SessionFilter,
					Name:                   "session filter for login page",
					CacheAdapterIdentifier: cacheAdapterIdentifier,
					CookieDomain:           "localhost",
					CookiePath:             "/",
					CookieName:             "session",
					CookieTTLHours:         24,
					CookieRenewBeforeHours: 2,
				},
			},
		},
	}, GoogleSecret{
		ClientId:     "google-client-id-1",
		ClientSecret: "google-client-secret",
	})
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
				Method: "POST",
				Url:    "/api/v3/mutable-resource",
			},
			Response: Response{
				Status:  201,
				Headers: nil,
				Body: JsonMap{
					"status":  "OK",
					"service": "resource",
					"version": "v3",
				},
			},
		},
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
		{
			Request: Request{
				Method: "GET",
				Url:    "/pages/work-page",
			},
			Response: Response{
				Status:  200,
				Headers: nil,
				Body: JsonMap{
					"status":  "OK",
					"service": "pages",
					"version": "work-page",
				},
			},
		},
		{
			Request: Request{
				Method: "GET",
				Url:    "/pages/login-page",
			},
			Response: Response{
				Status:  200,
				Headers: nil,
				Body: JsonMap{
					"status":  "OK",
					"service": "pages",
					"version": "login-page",
				},
			},
		},
	})
}

func createGoogleApiStub() *httptest.Server {
	return CreateServiceStub([]RequestMock{
		{ // Token retrieving request
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
		{ // User data request
			Request: Request{
				Method: "GET",
				Url:    "/oauth2/v3/userinfo",
				Headers: []Header{
					{
						Name:   "Authorization",
						Regexp: "^Bearer access-token-1$",
					},
				},
			},
			Response: Response{
				Status: 200,
				Headers: map[string]string{
					"Content-Type": "application/json;charset=UTF-8",
				},
				Body: JsonMap{
					"sub":    "user-identifier-1",
					"name":   "Aleksandr kucheev",
					"email":  "some@eamil.ru",
					"locale": "ru",
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
