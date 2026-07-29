package models

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = ginkgo.Describe("ClusterInfo.DefaultStorageClass", func() {
	ginkgo.It("returns the storage class flagged default", func() {
		info := &ClusterInfo{StorageClasses: []ClusterStorageClass{
			{Name: "slow"},
			{Name: "local-path", IsDefault: true},
		}}
		gomega.Expect(info.DefaultStorageClass()).To(gomega.Equal("local-path"))
	})

	ginkgo.It("returns empty when no class is default", func() {
		info := &ClusterInfo{StorageClasses: []ClusterStorageClass{{Name: "slow"}}}
		gomega.Expect(info.DefaultStorageClass()).To(gomega.BeEmpty())
	})

	ginkgo.It("returns empty for a nil receiver", func() {
		var info *ClusterInfo
		gomega.Expect(info.DefaultStorageClass()).To(gomega.BeEmpty())
	})

	ginkgo.It("round-trips through Value and Scan, quantities included", func() {
		cpu := resource.MustParse("3800m")
		mem := resource.MustParse("7168Mi")
		in := ClusterInfo{
			KubernetesVersion: "v1.31.2",
			StorageClasses:    []ClusterStorageClass{{Name: "local-path", IsDefault: true}},
			Nodes: []ClusterNode{{
				Name:              "node-1",
				Ready:             true,
				AllocatableCPU:    &cpu,
				AllocatableMemory: &mem,
			}},
		}
		raw, err := in.Value()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())

		out := &ClusterInfo{}
		gomega.Expect(out.Scan(raw.([]byte))).To(gomega.Succeed())
		gomega.Expect(out.DefaultStorageClass()).To(gomega.Equal("local-path"))
		gomega.Expect(out.Nodes[0].AllocatableCPU.MilliValue()).To(gomega.Equal(int64(3800)))
		gomega.Expect(out.Nodes[0].AllocatableMemory.Value()).To(gomega.Equal(int64(7168 * 1024 * 1024)))
		gomega.Expect(out.Nodes[0].AllocatableEphemeralDisk).To(gomega.BeNil())
	})
})
