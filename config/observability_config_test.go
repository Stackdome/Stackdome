package config

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("metrics listener configuration", func() {
	It("loads the internal metrics bind address", func() {
		DeferCleanup(setEnvForTest(EnvMetricsBindAddress.Name, "127.0.0.1:19090"))
		cfg := NewApplicationConfig()

		cfg.LoadEnvVariables()

		Expect(cfg.Server.MetricsBindAddress).To(Equal("127.0.0.1:19090"))
	})
})

func setEnvForTest(key, value string) func() {
	previous, existed := os.LookupEnv(key)
	Expect(os.Setenv(key, value)).To(Succeed())
	return func() {
		if existed {
			Expect(os.Setenv(key, previous)).To(Succeed())
			return
		}
		Expect(os.Unsetenv(key)).To(Succeed())
	}
}
