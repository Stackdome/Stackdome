package clients

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("mostRecentPod", func() {
	It("returns nil when there are no pods", func() {
		Expect(mostRecentPod(nil)).To(BeNil())
	})

	It("returns the pod with the latest creation timestamp", func() {
		pods := []corev1.Pod{
			{ObjectMeta: metav1.ObjectMeta{Name: "old", CreationTimestamp: metav1.Unix(100, 0)}},
			{ObjectMeta: metav1.ObjectMeta{Name: "new", CreationTimestamp: metav1.Unix(200, 0)}},
			{ObjectMeta: metav1.ObjectMeta{Name: "mid", CreationTimestamp: metav1.Unix(150, 0)}},
		}
		got := mostRecentPod(pods)
		Expect(got).NotTo(BeNil())
		Expect(got.Name).To(Equal("new"))
	})
})
