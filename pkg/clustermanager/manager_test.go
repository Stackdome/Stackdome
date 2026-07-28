package clustermanager

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/Stackdome/stackdome/pkg/logger"
	"github.com/Stackdome/stackdome/pkg/models"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	certutil "k8s.io/client-go/util/cert"
)

func TestClusterManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ClusterManager Suite")
}

// identityDecryptor returns the input bytes unchanged, standing in for the real
// EncryptionService in tests.
type identityDecryptor struct{}

func (identityDecryptor) DecryptData(in string) ([]byte, error) {
	return []byte(in), nil
}

var _ = Describe("ReRegisterCluster", func() {
	const clusterID = "cluster-1"

	var (
		cm      *ClusterManagerImpl
		caData  string
		makeClu func(token string) *models.Cluster
	)

	BeforeEach(func() {
		certPEM, _, err := certutil.GenerateSelfSignedCertKey("localhost", nil, nil)
		Expect(err).NotTo(HaveOccurred())
		caData = base64.StdEncoding.EncodeToString(certPEM)

		cm = NewClusterManager(ClusterManagerConfig{
			CredentialDecryptor: identityDecryptor{},
			Logger:              logger.NewLogger(),
		}).(*ClusterManagerImpl)

		makeClu = func(token string) *models.Cluster {
			return &models.Cluster{
				ID:                     clusterID,
				Name:                   "test-cluster",
				ClusterURL:             "https://example.com:6443",
				EncryptedToken:         base64.StdEncoding.EncodeToString([]byte(token)),
				EncryptedClusterCAData: caData,
			}
		}

		// Serve the underlying supervisor so teardown can stop per-cluster
		// services, mirroring the running server.
		ctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)
		go func() { _ = cm.supervisor.Serve(ctx) }()
	})

	It("registers the cluster and exposes the rotated bearer token", func() {
		Expect(cm.RegisterCluster(makeClu("token-v1"))).To(Succeed())
		Expect(cm.IsClusterRegistered(clusterID)).To(BeTrue())

		restConfig, err := cm.GetRestConfig(clusterID)
		Expect(err).NotTo(HaveOccurred())
		Expect(restConfig.BearerToken).To(Equal("token-v1"))

		Eventually(func() error {
			return cm.ReRegisterCluster(makeClu("token-v2"))
		}).Should(Succeed())
		Expect(cm.IsClusterRegistered(clusterID)).To(BeTrue())

		rotated, err := cm.GetRestConfig(clusterID)
		Expect(err).NotTo(HaveOccurred())
		Expect(rotated.BearerToken).To(Equal("token-v2"))
	})

	It("registers a cluster that was never registered before", func() {
		Expect(cm.IsClusterRegistered(clusterID)).To(BeFalse())
		Expect(cm.ReRegisterCluster(makeClu("token-v1"))).To(Succeed())
		Expect(cm.IsClusterRegistered(clusterID)).To(BeTrue())
	})
})
