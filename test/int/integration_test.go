package int

import (
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ashishmax31/stackdome-api-server/test/int/bootstrap"
)

var env = &bootstrap.Environment{}

// specFailed tracks whether any spec has failed, used with KEEP_RESOURCES_ON_FAILURE
// to skip data clearing and abort remaining specs so cluster state is preserved.
var specFailed bool

var _ = ReportAfterEach(func(report SpecReport) {
	if report.Failed() && os.Getenv("KEEP_RESOURCES_ON_FAILURE") == "true" {
		specFailed = true
	}
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Test Suite")
}

var _ = BeforeEach(func() {
	if specFailed {
		Skip("Previous spec failed with KEEP_RESOURCES_ON_FAILURE=true — skipping remaining specs to preserve cluster state")
	}
	By("Clearing test data")
	Expect(env.Database.ClearData(context.Background())).To(Succeed())
})

var _ = BeforeSuite(func() {
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
