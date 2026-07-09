package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.Describe("ReleaseEvent", func() {
	ginkgo.Describe("ReleaseEventLinks", func() {
		ginkgo.It("round-trip Value/Scan preserves data", func() {
			links := ReleaseEventLinks{{
				Kind:   ReleaseEventLinkKindBuildLogs,
				Label:  "View build logs",
				Target: map[string]string{"build_id": "b1", "resource_name": "api"},
			}}

			v, err := links.Value()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			var out ReleaseEventLinks
			err = out.Scan(v)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(out).To(gomega.Equal(links))
		})

		ginkgo.It("Scan(nil) yields nil without error", func() {
			var out ReleaseEventLinks
			err := out.Scan(nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(out).To(gomega.BeNil())
		})

		ginkgo.It("Value() on nil returns nil, nil", func() {
			var links ReleaseEventLinks
			v, err := links.Value()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(v).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("ReleaseEventMetadata", func() {
		ginkgo.It("round-trip Value/Scan preserves data", func() {
			m := ReleaseEventMetadata{ReleaseEventMetaReason: "image not found"}

			v, err := m.Value()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())

			var out ReleaseEventMetadata
			err = out.Scan(v)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(out).To(gomega.Equal(m))
		})

		ginkgo.It("Scan(nil) yields nil without error", func() {
			var out ReleaseEventMetadata
			err := out.Scan(nil)
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(out).To(gomega.BeNil())
		})

		ginkgo.It("Value() on nil returns nil, nil", func() {
			var m ReleaseEventMetadata
			v, err := m.Value()
			gomega.Expect(err).NotTo(gomega.HaveOccurred())
			gomega.Expect(v).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("ReleaseEvent", func() {
		ginkgo.It("TableName() returns 'release_events'", func() {
			event := ReleaseEvent{}
			gomega.Expect(event.TableName()).To(gomega.Equal("release_events"))
		})
	})
})
