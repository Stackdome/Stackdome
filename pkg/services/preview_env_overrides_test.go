package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/stackfile"
)

var _ = Describe("applyEnvOverrides", func() {
	It("layers stackfile < .env.preview < config, applied to every resource", func() {
		sf := &stackfile.Stackfile{
			Name: "demo",
			Resources: map[string]stackfile.Resource{
				"web": {Image: "nginx:1", Env: map[string]string{"FOO": "stack"}},
				"api": {Image: "api:1"}, // nil Env
			},
		}

		applyEnvOverrides(sf,
			map[string]string{"FOO": "repo", "BAR": "repo"},
			models.EnvVars{{Name: "FOO", Value: "cfg"}},
		)

		Expect(sf.Resources["web"].Env).To(Equal(map[string]string{"FOO": "cfg", "BAR": "repo"}))
		Expect(sf.Resources["api"].Env).To(Equal(map[string]string{"FOO": "cfg", "BAR": "repo"}))
	})

	It("is a no-op when there are no overrides", func() {
		sf := &stackfile.Stackfile{
			Name:      "demo",
			Resources: map[string]stackfile.Resource{"web": {Image: "nginx:1"}},
		}
		applyEnvOverrides(sf, nil, nil)
		Expect(sf.Resources["web"].Env).To(BeNil())
	})

	It("flows overrides into the built stack's resource env", func() {
		sf := &stackfile.Stackfile{
			Name:      "demo",
			Resources: map[string]stackfile.Resource{"web": {Image: "nginx:1", Env: map[string]string{"FOO": "stack"}}},
		}
		applyEnvOverrides(sf, map[string]string{"BAR": "repo"}, models.EnvVars{{Name: "FOO", Value: "cfg"}})

		stack, err := sf.ToStack()
		Expect(err).ToNot(HaveOccurred())

		var web *string
		got := map[string]string{}
		for _, r := range stack.Spec.StackResources {
			if r.Name != "web" || r.ExecutionConfig == nil {
				continue
			}
			web = &r.Name
			for _, e := range r.ExecutionConfig.EnvironmentVariables {
				got[e.Name] = e.GetValue()
			}
		}
		Expect(web).ToNot(BeNil(), "web resource must be present")
		Expect(got).To(HaveKeyWithValue("FOO", "cfg"))
		Expect(got).To(HaveKeyWithValue("BAR", "repo"))
	})
})
