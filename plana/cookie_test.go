package plana_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/arisu-archive/go-plana/plana"
)

var _ = Describe("CookieService", func() {
	Describe("GetCookie", func() {
		It("decodes JSON response into Cookie", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Method).To(Equal(http.MethodPost))
				Expect(r.URL.Path).To(Equal("/cookie"))
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"success":true,"cookie":"abc","timestamp":"2025-12-22T00:00:00Z"}`))
			}))
			DeferCleanup(ts.Close)

			baseURL, err := url.Parse(ts.URL)
			Expect(err).NotTo(HaveOccurred())

			client := plana.NewClient(nil, nil, ts.Client())
			client.GetCookieURL = baseURL

			out, err := client.Cookie.GetCookie(context.Background(), plana.GetCookieOptions{UserID: "u", Seed: "s"})
			Expect(err).NotTo(HaveOccurred())
			Expect(out).NotTo(BeNil())
			Expect(out.Success).To(BeTrue())
			Expect(out.Cookie).To(Equal("abc"))
			Expect(out.Timestamp.Equal(time.Date(2025, 12, 22, 0, 0, 0, 0, time.UTC))).To(BeTrue())
		})

		It("adds Authorization header when token provided", func() {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				Expect(r.Header.Get("Authorization")).To(Equal("Bearer token123"))
				_, _ = w.Write([]byte(`{"success":true,"cookie":"abc","timestamp":"2025-12-22T00:00:00Z"}`))
			}))
			DeferCleanup(ts.Close)

			baseURL, err := url.Parse(ts.URL)
			Expect(err).NotTo(HaveOccurred())

			client := plana.NewClient(nil, nil, ts.Client())
			client.GetCookieURL = baseURL

			_, err = client.Cookie.GetCookie(context.Background(), plana.GetCookieOptions{UserID: "u", Seed: "s", AuthToken: strings.Repeat("token", 1) + "123"})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
