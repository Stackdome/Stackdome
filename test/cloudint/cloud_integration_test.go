//go:build cloud_e2e

package cloudint

import (
	"context"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/test/int/bootstrap"
)

var cloudEnv = &bootstrap.Environment{}

func TestStackdomeCloudIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stackdome Cloud Integration Suite")
}

var _ = BeforeSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	Expect(bootstrap.SetupCloud(cloudEnv, ctx)).To(Succeed())
	DeferCleanup(func() {
		cancel()
		cloudEnv.Cleanup()
	})
})
