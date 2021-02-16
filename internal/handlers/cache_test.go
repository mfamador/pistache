package handlers_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/labstack/echo/v4"
	"github.com/mfamador/pistache/internal/handlers"
	"github.com/mfamador/pistache/internal/models"
	"github.com/mfamador/pistache/internal/server"
	"github.com/mfamador/pistache/internal/services"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
)

type mockCacheService struct{}

// GetCachedResponse handles caching a response
func (cs *mockCacheService) GetCachedResponse(req *http.Request) (string, *models.Response, error) {
	log.Info().
		Interface("req", req).
		Msg("cache response")

	return "", nil, nil
}

// Store stores a response
func (cs *mockCacheService) Skip(req *http.Request) bool {
	log.Info().
		Interface("resp", req).
		Msg("Skip cache")

	return true
}

// Store stores a response
func (cs *mockCacheService) Store(s string, resp *models.Response) bool {
	log.Info().
		Str("key", s).
		Interface("resp", resp).
		Msg("Store response")

	return true
}

type mockProxyService struct{}

func (ps *mockProxyService) Request(c echo.Context) (*models.Response, error) {
	return nil, nil
}

var _ = Describe("Cache handler", func() {
	var (
		w   *httptest.ResponseRecorder
		h   handlers.Cache
		s   services.Cache
		e   *echo.Echo
		err error
	)

	BeforeEach(func() {
		w = httptest.NewRecorder()
		e, err = server.Setup()
		s = &mockCacheService{}
		ps := &mockProxyService{}
		h = handlers.NewCache(s, ps)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should cache a response", func() {
		r, err := http.NewRequest("GET", "/", nil)
		Expect(err).ToNot(HaveOccurred())

		r.Header.Add("Content-Type", "application/json")
		c := e.NewContext(r, w)
		c.SetPath("/cacheable-not-valid-request")

		Expect(h.Handle(c)).To(Succeed())

		body, err := ioutil.ReadAll(w.Body)
		Expect(err).ToNot(HaveOccurred())

		Expect(w.Code).To(Equal(http.StatusOK))
		log.Info().
			Bytes("body", body).
			Msg("Response")
	})
})
