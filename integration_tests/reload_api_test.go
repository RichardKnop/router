package integration

import (
	"crypto/sha1"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("reload API endpoint", func() {

	Describe("request handling", func() {
		It("should return 200 for POST /reload", func() {
			resp := doRequest(newRequest("POST", routerAPIURL("/reload")))
			Expect(resp.StatusCode).To(Equal(205))
		})

		It("should return 404 for POST /foo", func() {
			resp := doRequest(newRequest("POST", routerAPIURL("/foo")))
			Expect(resp.StatusCode).To(Equal(404))
		})

		It("should return 404 for POST /reload/foo", func() {
			resp := doRequest(newRequest("POST", routerAPIURL("/reload/foo")))
			Expect(resp.StatusCode).To(Equal(404))
		})

		It("should return 405 for GET /reload", func() {
			resp := doRequest(newRequest("GET", routerAPIURL("/reload")))
			Expect(resp.StatusCode).To(Equal(405))
			Expect(resp.Header.Get("Allow")).To(Equal("POST"))
		})
	})

	Describe("healthcheck", func() {
		It("should return 200 and sting 'OK' on /healthcheck", func() {
			resp := doRequest(newRequest("GET", routerAPIURL("/healthcheck")))
			Expect(resp.StatusCode).To(Equal(200))
			Expect(readBody(resp)).To(Equal("OK"))
		})

		It("should return 405 for other verbs", func() {
			resp := doRequest(newRequest("POST", routerAPIURL("/healthcheck")))
			Expect(resp.StatusCode).To(Equal(405))
			Expect(resp.Header.Get("Allow")).To(Equal("GET"))
		})
	})

	Describe("route stats", func() {

		Context("with some routes loaded", func() {
			var data map[string]map[string]interface{}

			BeforeEach(func() {
				addRedirectRoute("/foo", "/bar", "prefix")
				addRedirectRoute("/baz", "/qux", "prefix")
				addRedirectRoute("/foo", "/bar/baz")
				reloadRoutes()
				resp := doRequest(newRequest("GET", routerAPIURL("/stats")))
				Expect(resp.StatusCode).To(Equal(200))
				readJSONBody(resp, &data)
			})

			It("should return the number of routes loaded", func() {
				Expect(data["routes"]["count"]).To(BeEquivalentTo(3))
			})

			It("should return a checksum calculated from the sorted paths and route_types", func() {
				hash := sha1.New()
				hash.Write([]byte("/baz(true)"))
				hash.Write([]byte("/foo(false)"))
				hash.Write([]byte("/foo(true)"))
				Expect(data["routes"]["checksum"]).To(Equal(fmt.Sprintf("%x", hash.Sum(nil))))
			})
		})

		Context("with no routes", func() {
			var data map[string]map[string]interface{}

			BeforeEach(func() {
				reloadRoutes()

				resp := doRequest(newRequest("GET", routerAPIURL("/stats")))
				Expect(resp.StatusCode).To(Equal(200))
				readJSONBody(resp, &data)
			})

			It("should return the number of routes loaded", func() {
				Expect(data["routes"]["count"]).To(BeEquivalentTo(0))
			})

			It("should return a checksum of empty string", func() {
				hash := sha1.New()
				Expect(data["routes"]["checksum"]).To(Equal(fmt.Sprintf("%x", hash.Sum(nil))))
			})
		})

		It("should return 405 for other verbs", func() {
			resp := doRequest(newRequest("POST", routerAPIURL("/stats")))
			Expect(resp.StatusCode).To(Equal(405))
			Expect(resp.Header.Get("Allow")).To(Equal("GET"))
		})
	})
})
