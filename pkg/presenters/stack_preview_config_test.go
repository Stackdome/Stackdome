package presenters_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/Stackdome/stackdome/pkg/presenters"
)

var _ = Describe("StackPreviewConfig env presenter", func() {
	It("round-trips env vars including a secret ref", func() {
		req := &openapi.StackPreviewConfigCreate{
			Name:          "web",
			GitRepository: openapi.PreviewGitRepository{RepoUrl: "https://github.com/acme/app"},
			Env: []openapi.EnvVar{
				{Name: "FOO", Value: openapi.PtrString("bar")},
				{Name: "SECRET_TOKEN", Value: openapi.PtrString("{{ secret.api-token }}")},
			},
		}

		model := presenters.ConvertStackPreviewConfigCreate(req)
		Expect(model.Env).To(HaveLen(2))
		Expect(model.Env[0].Name).To(Equal("FOO"))
		Expect(model.Env[0].Value).To(Equal("bar"))
		Expect(model.Env[1].Value).To(Equal("{{ secret.api-token }}"))

		out := presenters.PresentStackPreviewConfig(&models.StackPreviewConfig{Env: model.Env})
		Expect(out.Env).To(HaveLen(2))
		Expect(out.Env[0].GetValue()).To(Equal("bar"))
		Expect(out.Env[1].GetName()).To(Equal("SECRET_TOKEN"))
	})
})
