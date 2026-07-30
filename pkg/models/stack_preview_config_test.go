package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("StackPreviewConfig.NormalizedRepoURL", func() {
	ginkgo.DescribeTable("collapses equivalent repo URLs",
		func(a, b string) {
			ca := StackPreviewConfig{GitRepository: PreviewGitRepository{RepoURL: a}}
			cb := StackPreviewConfig{GitRepository: PreviewGitRepository{RepoURL: b}}
			gomega.Expect(ca.NormalizedRepoURL()).To(gomega.Equal(cb.NormalizedRepoURL()))
		},
		ginkgo.Entry(".git suffix", "https://github.com/acme/app.git", "https://github.com/acme/app"),
		ginkgo.Entry("trailing slash", "https://github.com/acme/app/", "https://github.com/acme/app"),
		ginkgo.Entry(".git plus trailing slash", "https://github.com/acme/app.git/", "https://github.com/acme/app"),
		ginkgo.Entry("http scheme", "http://github.com/acme/app", "https://github.com/acme/app"),
		ginkgo.Entry("scp form", "git@github.com:acme/app.git", "https://github.com/acme/app"),
		ginkgo.Entry("ssh scheme with userinfo", "ssh://git@github.com/acme/app", "https://github.com/acme/app"),
		ginkgo.Entry("case", "https://github.com/Acme/App", "https://github.com/acme/app"),
	)

	ginkgo.It("normalizes to host/owner/repo", func() {
		gomega.Expect(NormalizeRepoURL("https://github.com/Acme/App.git")).To(gomega.Equal("github.com/acme/app"))
	})
})
