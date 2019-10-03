package integration_test

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/url"
)

var _ = Describe("Request", func() {

	It("Can reverse proxy to unblocked resources", func() {
		resp, err := http.Get("http://localhost" + server.Addr + "/api/v1/resource")
		if err != nil {
			Fail(err.Error())
		}
		message, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			Fail(err.Error())
		}
		Expect(resp.StatusCode).To(Equal(200))
		messageMap := make(map[string]string)
		err = json.Unmarshal(message, &messageMap)
		if err != nil {
			Fail(err.Error())
		}
		Expect(messageMap).To(HaveKeyWithValue("status", "OK"))
		Expect(messageMap).To(HaveKeyWithValue("service", "resource"))
		Expect(messageMap).To(HaveKeyWithValue("version", "v1"))
	})

	It("checkGoogleStub", func() {
		params := url.Values{}
		params.Add("code", "google-auth-code")
		params.Add("client_id", "google-client-id-1")
		params.Add("client_secret", "google-client-secret")
		params.Add("redirect_uri", "http://localhost:8080/authentication/google")
		params.Add("grant_type", "authorization_code")

		bodyReader := bytes.NewBufferString(params.Encode())
		resp, err := http.Post(googleApiStub.URL+"/oauth2/v4/token",
			"application/x-www-form-urlencoded",
			bodyReader,
		)
		if err != nil {
			Fail(err.Error())
		}
		response, err := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
		if err != nil {
			Fail(err.Error())
		}

		result := make(map[string]interface{})
		err = json.Unmarshal(response, &result)
		Expect(resp.StatusCode).To(Equal(200), "Body: %v", string(response))
		Expect(result).To(HaveKeyWithValue("access_token", "access-token-1"), "Body: %v", string(response))
	})
})
