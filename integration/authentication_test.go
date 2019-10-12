package integration_test

import (
	"bytes"
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
)

var _ = Describe("In ordinator gateway", func() {

	It("ReverseProxy can proxy to resources", func() {
		resp, message := get("http://localhost" + server.Addr + "/api/v1/resource")
		Expect(resp.StatusCode).To(Equal(200))

		messageMap := unmarshalToMap(message)
		Expect(messageMap).To(HaveKeyWithValue("status", "OK"))
		Expect(messageMap).To(HaveKeyWithValue("service", "resource"))
		Expect(messageMap).To(HaveKeyWithValue("version", "v1"))
	})

	It("UserAuthenticationFilter can block access to resource", func() {
		resp, _ := get("http://localhost" + server.Addr + "/api/v2/resource")
		Expect(resp.StatusCode).To(Equal(401))
	})

	It("GoogleOauth2Authorization can authenticate in google", func() {
		resp, message := get("http://localhost" + server.Addr + "/authentication/google?code=google-auth-code")
		Expect(resp.StatusCode).To(Equal(200))

		messageMap := unmarshalToMap(message)
		Expect(messageMap).To(HaveKeyWithValue("status", "OK"))
		Expect(messageMap).To(HaveKeyWithValue("service", "resource"))
		Expect(messageMap).To(HaveKeyWithValue("version", "v2"))
	})
})

func unmarshalToMap(message []byte) map[string]string {
	messageMap := make(map[string]string)
	if err := json.Unmarshal(message, &messageMap); err != nil {
		Fail(err.Error())
	}
	return messageMap
}

func get(url string) (*http.Response, []byte) {
	return getByClient(buildClient(), url)
}

func getByClient(client *http.Client, url string) (*http.Response, []byte) {
	resp, err := client.Get(url)
	if err != nil {
		Fail(err.Error())
	}
	message, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		Fail(err.Error())
	}
	return resp, message
}

type requestMutator func(r *http.Request) *http.Request

func postJsonByClient(client *http.Client, url string, body interface{}, mutator requestMutator) (*http.Response, []byte) {
	bytesValue, err := json.Marshal(body)
	if err != nil {
		Fail(err.Error())
	}
	request, err := http.NewRequest(
		"POST",
		url,
		bytes.NewReader(bytesValue),
	)
	request.Header.Set("Content-Type", "application/json")
	if mutator != nil {
		request = mutator(request)
	}
	resp, err := client.Do(request)
	if err != nil {
		Fail(err.Error())
	}
	message, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		Fail(err.Error())
	}
	return resp, message
}

func buildClient() *http.Client {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		log.Fatal(err)
	}
	return &http.Client{Jar: jar}
}
