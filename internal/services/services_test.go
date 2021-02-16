package services_test

import (
	"fmt"
	"testing"

	"github.com/mfamador/pistache/internal/config"
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
