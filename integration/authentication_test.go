package integration_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	_ "github.com/onsi/gomega"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Request", func() {

	var (
		mockServer *httptest.Server
	)

	BeforeEach(func() {
		handler := ResourceServerMock{}
		mockServer = httptest.NewServer(&handler)
	})

	AfterEach(func() {
		mockServer.Close()
	})

	It("test", func() {
		resp, err := http.Get("http://localhost:8080/api/v1/some")
		if err != nil {
			Fail(err.Error())
		}
		greeting, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			Fail(err.Error())
		}
		gomega.Expect(string(greeting)).To(gomega.Equal("404 page not found\n"))
	})
})

type ResourceServerMock struct {
}

func (r *ResourceServerMock) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprint(writer, "Hello world")
}
