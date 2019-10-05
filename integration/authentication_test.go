package integration_test

import (
	"encoding/json"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
)

var _ = Describe("Ordinator gateway", func() {

	It("can reverse proxy to unblocked resources", func() {
		resp, message := get("http://localhost" + server.Addr + "/api/v1/resource")
		Expect(resp.StatusCode).To(Equal(200))

		messageMap := unmarshalToMap(message)
		Expect(messageMap).To(HaveKeyWithValue("status", "OK"))
		Expect(messageMap).To(HaveKeyWithValue("service", "resource"))
		Expect(messageMap).To(HaveKeyWithValue("version", "v1"))
	})

	It("can block access to resource", func() {
		resp, _ := get("http://localhost" + server.Addr + "/api/v2/resource")
		Expect(resp.StatusCode).To(Equal(401))
	})

	It("can authenticate in google", func() {
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
	resp, err := buildClient().Get(url)
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
