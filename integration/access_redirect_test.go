package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserAuthenticationFilter", func() {

	It("redirect when access denied", func() {
		resp, message := get("http://localhost" + server.Addr + "/pages/work-page")
		Expect(resp.StatusCode).To(Equal(200))

		messageMap := unmarshalToMap(message)
		Expect(messageMap).To(HaveKeyWithValue("status", "OK"))
		Expect(messageMap).To(HaveKeyWithValue("service", "pages"))
		Expect(messageMap).To(HaveKeyWithValue("version", "login-page"))
	})

	It("get page when access accepted", func() {
		client := buildClient()
		resp, _ := getByClient(
			client,
			"http://localhost"+server.Addr+"/authentication/google?code=google-auth-code",
		)
		Expect(resp.StatusCode).To(Equal(200))

		resp, message := getByClient(client, "http://localhost"+server.Addr+"/pages/work-page")
		Expect(resp.StatusCode).To(Equal(200))

		messageMap := unmarshalToMap(message)
		Expect(messageMap).To(HaveKeyWithValue("status", "OK"))
		Expect(messageMap).To(HaveKeyWithValue("service", "pages"))
		Expect(messageMap).To(HaveKeyWithValue("version", "work-page"))
	})
})
