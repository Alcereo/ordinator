package integration_test

import (
	. "github.com/Alcereo/ordinator/integration/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("CSRF", func() {
	const headerName = "X-CSRF-TOKEN"

	It("CsrfFilter allow unsafe methods with CSRF token", func() {
		client := buildClient()
		resp, _ := getByClient(
			client,
			"http://localhost"+server.Addr+"/authentication/google?code=google-auth-code",
		)
		Expect(resp.StatusCode).To(Equal(200))
		csrfToken := resp.Header.Get(headerName)
		Expect(csrfToken).NotTo(BeEmpty(), "CSRF token not found in header: %v", headerName)

		resp, bytes := postJsonByClient(
			client,
			"http://localhost"+server.Addr+"/api/v3/mutable-resource",
			JsonMap{"method": "mutate"},
			func(req *http.Request) *http.Request {
				req.Header.Add(headerName, csrfToken)
				return req
			},
		)

		Expect(resp.StatusCode).To(Equal(201))

		messageMap := unmarshalToMap(bytes)
		Expect(messageMap).To(HaveKeyWithValue("status", "OK"))
		Expect(messageMap).To(HaveKeyWithValue("service", "resource"))
		Expect(messageMap).To(HaveKeyWithValue("version", "v3"))
	})

	It("CsrfFilter allow safe methods without CSRF token", func() {
		client := buildClient()
		resp, _ := getByClient(
			client,
			"http://localhost"+server.Addr+"/authentication/google?code=google-auth-code",
		)
		Expect(resp.StatusCode).To(Equal(200))

		resp, bytes := getByClient(
			client,
			"http://localhost"+server.Addr+"/api/v2/resource",
		)

		Expect(resp.StatusCode).To(Equal(200))

		messageMap := unmarshalToMap(bytes)
		Expect(messageMap).To(HaveKeyWithValue("status", "OK"))
		Expect(messageMap).To(HaveKeyWithValue("service", "resource"))
		Expect(messageMap).To(HaveKeyWithValue("version", "v2"))
	})

	It("CsrfFilter denied methods without CSRF token", func() {
		client := buildClient()
		resp, _ := getByClient(
			client,
			"http://localhost"+server.Addr+"/authentication/google?code=google-auth-code",
		)
		Expect(resp.StatusCode).To(Equal(200))
		csrfToken := resp.Header.Get(headerName)
		Expect(csrfToken).NotTo(BeEmpty(), "CSRF token not found in header: %v", headerName)

		resp, bytes := postJsonByClient(
			client,
			"http://localhost"+server.Addr+"/api/v3/mutable-resource",
			JsonMap{"method": "mutate"},
			nil,
		)

		Expect(resp.StatusCode).To(Equal(403))

		messageString := string(bytes)
		Expect(messageString).To(Equal("resolving CSRF header error. CSRF header: X-CSRF-TOKEN is empty"))
	})
})
