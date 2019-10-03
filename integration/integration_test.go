package integration_test

import (
	"encoding/json"
	"fmt"
	. "github.com/Alcereo/ordinator/pkg/context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
)

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var server *http.Server
var googleApiStub *httptest.Server
var resourceStub *httptest.Server

type RequestMock struct {
	request  Request
	response Response
}

type Header struct {
	name   string
	regexp string
}

type URLPropsBody struct {
	props map[string]string
}

func (check URLPropsBody) checkBody(body []byte, req *http.Request) error {

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return fmt.Errorf("parsing queries from body errror. %v", err.Error())
	}

	for key, value := range check.props {
		if values.Get(key) != value {
			return fmt.Errorf("property %v=%v not match with expected: %v", key, values.Get(key), value)
		}
	}

	return nil
}

type Request struct {
	method  string
	url     string
	headers []Header
	body    []BodyCheck
}

type BodyCheck interface {
	checkBody([]byte, *http.Request) error
}

type StringedBody interface {
	getString() ([]byte, error)
}

type Response struct {
	status  int
	headers map[string]string
	body    StringedBody
}

type jsonMap map[string]interface{}

func (s jsonMap) getString() ([]byte, error) {
	return json.Marshal(s)
}

var _ = BeforeSuite(func() {

	googleApiStub = createServerStub([]RequestMock{
		{
			request: Request{
				method: "POST",
				url:    "/oauth2/v4/token",
				headers: []Header{
					{
						name:   "Content-Type",
						regexp: "application/x-www-form-urlencoded",
					},
				},
				body: []BodyCheck{
					URLPropsBody{
						props: map[string]string{
							"code":          "google-auth-code",
							"client_id":     "google-client-id-1",
							"client_secret": "google-client-secret",
							"redirect_uri":  "http://localhost:8080/authentication/google",
							"grant_type":    "authorization_code",
						},
					},
				},
			},
			response: Response{
				status: 200,
				headers: map[string]string{
					"Content-Type": "application/json;charset=UTF-8",
				},
				body: jsonMap{
					"access_token":  "access-token-1",
					"refresh_token": "refresh-token-1",
					"expires_in":    9999,
					"token_type":    "authorization_code",
				},
			},
		},
	})

	resourceStub = createServerStub([]RequestMock{
		{
			request: Request{
				method: "GET",
				url:    "/api/v2/resource",
			},
			response: Response{
				status:  200,
				headers: nil,
				body: jsonMap{
					"status":  "OK",
					"service": "resource",
					"version": "v2",
				},
			},
		},
		{
			request: Request{
				method: "GET",
				url:    "/api/v1/resource",
			},
			response: Response{
				status:  200,
				headers: nil,
				body: jsonMap{
					"status":  "OK",
					"service": "resource",
					"version": "v1",
				},
			},
		},
	})

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

func createServerStub(mocks []RequestMock) *httptest.Server {
	mux := http.NewServeMux()

	for _, mReg := range mocks {
		mux.HandleFunc(mReg.request.url, func(writer http.ResponseWriter, request *http.Request) {
			if request.Method != mReg.request.method {
				writer.WriteHeader(404)
				_, _ = fmt.Fprint(writer, "Request Method 'POST' expected. Actual: '"+request.Method+"'.")
				return
			}

			for _, check := range mReg.request.headers {
				header := request.Header.Get(check.name)
				matched, err := regexp.MatchString(check.regexp, header)
				if err != nil {
					writer.WriteHeader(500)
					_, _ = fmt.Fprint(writer, "Parsing header regexp error: "+check.regexp+". Detail: "+err.Error())
					return
				}
				if !matched {
					writer.WriteHeader(400)
					_, _ = fmt.Fprint(writer, "Header not matched regexp. Header: "+header+". Regexp: "+check.regexp)
					return
				}
			}

			bytes, err := ioutil.ReadAll(request.Body)
			if err != nil {
				writer.WriteHeader(500)
				_, _ = fmt.Fprint(writer, "Reading body error: "+err.Error())
				return
			}

			for _, check := range mReg.request.body {
				err := check.checkBody(bytes, request)
				if err != nil {
					writer.WriteHeader(400)
					_, _ = fmt.Fprint(writer, "Body not match: "+err.Error())
					return
				}
			}

			writer.WriteHeader(mReg.response.status)
			for header, value := range mReg.response.headers {
				writer.Header().Add(header, value)
			}
			bodyBytes, err := mReg.response.body.getString()
			if err != nil {
				writer.WriteHeader(500)
				_, _ = fmt.Fprint(writer, "Writing body error: "+err.Error())
				return
			}
			_, err = fmt.Fprint(writer, string(bodyBytes))
			if err != nil {
				writer.WriteHeader(500)
				_, _ = fmt.Fprint(writer, "Writing body error: "+err.Error())
				return
			}
		})
	}

	return httptest.NewServer(mux)
}

var _ = AfterSuite(func() {
	err := server.Close()
	if err != nil {
		Fail(err.Error())
	}
	resourceStub.Close()
	googleApiStub.Close()
})
