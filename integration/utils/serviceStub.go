package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
)

func CreateServiceStub(mocks []RequestMock) *httptest.Server {
	mux := http.NewServeMux()

	for _, mReg := range mocks {
		pattern := mReg.Request.Url
		expectedMethod := mReg.Request.Method
		expectedHeaders := mReg.Request.Headers
		expectedBodyChecks := mReg.Request.Body
		responseStatus := mReg.Response.Status
		responseHeaders := mReg.Response.Headers
		responseBody := mReg.Response.Body

		mux.HandleFunc(pattern, func(writer http.ResponseWriter, request *http.Request) {
			if request.Method != expectedMethod {
				writer.WriteHeader(404)
				_, _ = fmt.Fprint(writer, "Request Method 'POST' expected. Actual: '"+request.Method+"'.")
				return
			}

			for _, check := range expectedHeaders {
				header := request.Header.Get(check.Name)
				matched, err := regexp.MatchString(check.Regexp, header)
				if err != nil {
					writer.WriteHeader(500)
					_, _ = fmt.Fprint(writer, "Parsing header regexp error: "+check.Regexp+". Detail: "+err.Error())
					return
				}
				if !matched {
					writer.WriteHeader(400)
					_, _ = fmt.Fprint(writer, "Header not matched regexp. Header: "+header+". Regexp: "+check.Regexp)
					return
				}
			}

			bytes, err := ioutil.ReadAll(request.Body)
			if err != nil {
				writer.WriteHeader(500)
				_, _ = fmt.Fprint(writer, "Reading body error: "+err.Error())
				return
			}

			for _, check := range expectedBodyChecks {
				err := check.checkBody(bytes, request)
				if err != nil {
					writer.WriteHeader(400)
					_, _ = fmt.Fprint(writer, "Body not match: "+err.Error())
					return
				}
			}

			writer.WriteHeader(responseStatus)
			for header, value := range responseHeaders {
				writer.Header().Add(header, value)
			}
			bodyBytes, err := responseBody.getString()
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

type RequestMock struct {
	Request  Request
	Response Response
}

type Header struct {
	Name   string
	Regexp string
}

type URLPropsBody struct {
	Props map[string]string
}

func (check URLPropsBody) checkBody(body []byte, req *http.Request) error {

	values, err := url.ParseQuery(string(body))
	if err != nil {
		return fmt.Errorf("parsing queries from body errror. %v", err.Error())
	}

	for key, value := range check.Props {
		if values.Get(key) != value {
			return fmt.Errorf("property %v=%v not match with expected: %v", key, values.Get(key), value)
		}
	}

	return nil
}

type Request struct {
	Method  string
	Url     string
	Headers []Header
	Body    []BodyCheck
}

type BodyCheck interface {
	checkBody([]byte, *http.Request) error
}

type StringedBody interface {
	getString() ([]byte, error)
}

type Response struct {
	Status  int
	Headers map[string]string
	Body    StringedBody
}

type JsonMap map[string]interface{}

func (s JsonMap) getString() ([]byte, error) {
	return json.Marshal(s)
}
