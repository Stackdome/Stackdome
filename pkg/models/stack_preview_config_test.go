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
	)
})
