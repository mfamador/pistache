package handlers_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/mfamador/pistache/internal/config"
	"github.com/mfamador/pistache/internal/handlers"
	_ "github.com/mfamador/pistache/internal/logger"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	TestConfigDir = fmt.Sprintf("%s/test", config.Dir)
)

func TestCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cache Suite")
}

var _ = Describe("Handlers", func() {
	Context("GET /healthz", func() {
		var w *httptest.ResponseRecorder
		var ct handlers.Controller
		var e *echo.Echo

		BeforeEach(func() {
			w = httptest.NewRecorder()
			e = echo.New()
		})

		It("should return 204", func() {
			r, _ := http.NewRequest("GET", "/healthz", nil)

			c := e.NewContext(r, w)

			err := ct.Healthz(c)

			Expect(err).NotTo(HaveOccurred())
			Expect(w.Code).To(Equal(204))
		})
	})
})
