package services_test

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/mfamador/pistache/internal/models"

	"github.com/mfamador/pistache/internal/services"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog/log"
)

var (
	sucessRequest = &http.Request{
		Method: http.MethodGet,
		URL: &url.URL{
			Path: "/dummy",
		},
		Header: http.Header{},
	}
	postRequest = &http.Request{
		Method: http.MethodPost,
		URL: &url.URL{
			Path: "/dummy",
		},
		Header: http.Header{},
	}
	sucessResponse = &models.Response{
		StatusCode: http.StatusOK,
		Header:     nil,
		Body:       nil,
	}
	errorResponse = &models.Response{
		StatusCode: http.StatusBadRequest,
		Header:     nil,
		Body:       nil,
	}
	cf = &services.CacheConfig{
		Redis:      nil,
		Exceptions: nil,
		Methods:    nil,
		TTL: struct {
			Success int `yaml:"success"`
			Error   int `yaml:"error"`
		}{
			Success: 2,
			Error:   1,
		},
		ForwardingHeaders: nil,
		Hash: struct {
			Prefix                string `yaml:"prefix" default:"app"`
			services.HashElements `yaml:",inline"`
			Overrides             []struct {
				OriginalPath          string `yaml:"originalPath"`
				services.HashElements `yaml:",inline"`
			} `yaml:"overrides"`
		}{
			Prefix: "test",
		},
	}
)

var _ = Describe("Cache service", func() {
	var (
		s   services.Cache
		err error
	)

	BeforeEach(func() {
		s, err = services.NewCache(cf)
		Expect(err).ToNot(HaveOccurred())
	})

	It("should return nil if not cached", func() {
		key, response, err := s.GetCachedResponse(sucessRequest)
		Expect(err).ToNot(HaveOccurred())

		Expect(response).To(BeNil())
		log.Info().
			Str("key", key).
			Interface("resp", response).
			Interface("success response", sucessResponse).
			Interface("error response", errorResponse).
			Interface("resp", response).
			Msg("Response")
	})

	It("should return a cached response", func() {
		key, _, err := s.GetCachedResponse(sucessRequest)
		Expect(err).ToNot(HaveOccurred())

		success := s.Store(key, sucessResponse)
		Expect(err).ToNot(HaveOccurred())
		Expect(success).To(BeTrue())

		Expect(err).ToNot(HaveOccurred())

		response := Eventually(func() *models.Response {
			_, response, _ := s.GetCachedResponse(sucessRequest)
			return response
		}).ShouldNot(BeNil())

		log.Info().
			Str("key", key).
			Interface("resp", response).
			Interface("success response", sucessResponse).
			Interface("error response", errorResponse).
			Interface("resp", response).
			Msg("Response")
	})

	It("should skip a POST", func() {
		skip := s.Skip(postRequest)
		Expect(skip).To(BeTrue())
	})

	It("should not skip a GET", func() {
		skip := s.Skip(sucessRequest)
		Expect(skip).To(BeTrue())
	})

	It("the generated key must have the prefix", func() {
		key, _, err := s.GetCachedResponse(sucessRequest)
		Expect(err).ToNot(HaveOccurred())
		Expect(strings.HasPrefix(key, "{test}-")).To(BeTrue())
	})

})
