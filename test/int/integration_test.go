package int

import (
	"context"
	"testing"
	"time"

	"github.com/joho/godotenv"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/test/int/bootstrap"
)

var env = &bootstrap.Environment{}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var _ = BeforeSuite(func() {
	err := godotenv.Load("../../.env")
	Expect(err).NotTo(HaveOccurred(), "Failed to load .env file for integration tests")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*15)
	bootstrapErr := bootstrap.Setup(env, ctx)
	if bootstrapErr != nil {
		Expect(bootstrapErr).NotTo(HaveOccurred(), "Failed to bootstrap integration test environment")
	}
	DeferCleanup(func() {
		cancel()
		if env != nil {
			env.Cleanup()
		}
	})
})

// GetEnvironment returns the shared test environment for use in test suites
func GetEnvironment() *bootstrap.Environment {
	return env
}
