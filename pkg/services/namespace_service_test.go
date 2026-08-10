package services

import (
	"context"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/config"
	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation"
)

func expectSharedComputeNamespaceLabels(namespace *models.Namespace, role string) {
	labels := namespace.Labels.ToMap()
	Expect(labels).To(HaveKeyWithValue(models.ManagedByLabelKey, models.ManagedByLabelValue))
	Expect(labels).To(HaveKeyWithValue(models.CloudTenantLabelKey, models.CloudTenantLabelValue))
	Expect(labels).To(HaveKeyWithValue(models.OrganizationIDLabelKey, namespace.OrganisationID))
	Expect(labels).To(HaveKeyWithValue(models.NamespaceRoleLabelKey, role))
}

func expectBringYourOwnNamespaceLabels(namespace *models.Namespace) {
	Expect(namespace.Labels.ToMap()).To(Equal(map[string]string{
		models.ManagedByLabelKey: models.ManagedByLabelValue,
	}))
}

type fakeAddon struct {
	addonType string
	addonName string
}

const addonType = "postgres"

func (f fakeAddon) Type() string      { return f.addonType }
func (f fakeAddon) AddonName() string { return f.addonName }

var _ = Describe("namespaceNameForAddon", func() {
	It("truncates a max-length name to a valid RFC 1123 label keeping the entropy floor", func() {
		addon := fakeAddon{
			addonType: addonType,
			addonName: strings.Repeat("a", models.MaxAddonNameLength),
		}
		s := &namespaceService{}

		ns, err := s.PrepareNamespaceForAddon(context.Background(), addon, "org-1")

		Expect(err).To(BeNil())
		Expect(len(ns.Name)).To(BeNumerically("<=", models.KubernetesDNSLabelMaxLength))
		Expect(validation.IsDNS1123Label(ns.Name)).To(BeEmpty())
		Expect(ns.Name).NotTo(HaveSuffix(models.NamespaceNameSeparator))

		prefix := addonPrefix(addonType, addon.addonName)
		Expect(ns.Name).To(HavePrefix(prefix))
		survivingUUID := strings.TrimPrefix(ns.Name, prefix)
		Expect(len(survivingUUID)).To(BeNumerically(">=", models.MinNamespaceUUIDSuffixLength))
	})

	It("generates different names for the same type and name", func() {
		addon := fakeAddon{addonType: addonType, addonName: "mydb"}
		s := &namespaceService{}

		first, err := s.PrepareNamespaceForAddon(context.Background(), addon, "org-1")
		Expect(err).To(BeNil())
		second, err := s.PrepareNamespaceForAddon(context.Background(), addon, "org-1")
		Expect(err).To(BeNil())

		Expect(first.Name).NotTo(Equal(second.Name))
	})

	It("keeps more UUID entropy for a short name than for a max-length name", func() {
		s := &namespaceService{}

		shortAddon := fakeAddon{addonType: addonType, addonName: "mydb"}
		short, err := s.PrepareNamespaceForAddon(context.Background(), shortAddon, "org-1")
		Expect(err).To(BeNil())
		shortSuffix := strings.TrimPrefix(short.Name, addonPrefix(addonType, shortAddon.addonName))

		maxAddon := fakeAddon{addonType: addonType, addonName: strings.Repeat("a", models.MaxAddonNameLength)}
		max, err := s.PrepareNamespaceForAddon(context.Background(), maxAddon, "org-1")
		Expect(err).To(BeNil())
		maxSuffix := strings.TrimPrefix(max.Name, addonPrefix(addonType, maxAddon.addonName))

		Expect(len(shortSuffix)).To(BeNumerically(">", len(maxSuffix)))
		Expect(len(shortSuffix)).To(BeNumerically(">=", models.MinNamespaceUUIDSuffixLength))
	})
})

var _ = Describe("shared compute namespace labels", func() {
	DescribeTable("follows compute mode independently of runtime mode",
		func(runtimeMode config.RuntimeMode, computeMode config.ComputeMode, expectTenantLabels bool) {
			applicationConfig := config.NewApplicationConfig()
			applicationConfig.RuntimeMode = runtimeMode
			applicationConfig.ComputeMode = computeMode
			service := NewNamespaceService(NamespaceServiceSpec{
				SharedCompute: applicationConfig.UsesSharedCompute(),
			})

			stackNamespace, err := service.PrepareNamespaceForStack(context.Background(), &models.Stack{
				Name: "api", OrganisationID: "organisation-1",
			})
			Expect(err).To(BeNil())
			addonNamespace, err := service.PrepareNamespaceForAddon(
				context.Background(), fakeAddon{addonType: addonType, addonName: "database"}, "organisation-1",
			)
			Expect(err).To(BeNil())

			if expectTenantLabels {
				expectSharedComputeNamespaceLabels(stackNamespace, models.NamespaceRoleStack)
				expectSharedComputeNamespaceLabels(addonNamespace, models.NamespaceRoleAddon)
				return
			}
			expectBringYourOwnNamespaceLabels(stackNamespace)
			expectBringYourOwnNamespaceLabels(addonNamespace)
		},
		Entry("self-hosted shared compute", config.RuntimeModeSelfHosted, config.ComputeModeShared, true),
		Entry("Stackdome Cloud shared compute", config.RuntimeModeStackdomeCloud, config.ComputeModeShared, true),
		Entry("self-hosted bring-your-own compute", config.RuntimeModeSelfHosted, config.ComputeModeBYOC, false),
	)
})

func addonPrefix(addonType, addonName string) string {
	return models.AddonNamespacePrefix +
		models.NamespaceNameSeparator + addonType +
		models.NamespaceNameSeparator + addonName +
		models.NamespaceNameSeparator
}

// TestPrepareNamespaceForStackTruncatesToDNSLabelAtMaxNameLength pins the
// contract behind models.MaxStackNameLength: a stack name at the validator's
// cap must generate a namespace name of exactly
// models.KubernetesDNSLabelMaxLength characters, keeping at least
// models.MinNamespaceUUIDSuffixLength UUID characters after truncation and
// never ending in a separator.
func TestPrepareNamespaceForStackTruncatesToDNSLabelAtMaxNameLength(t *testing.T) {
	s := &namespaceService{}
	stack := &models.Stack{
		Name:           strings.Repeat("a", models.MaxStackNameLength),
		OrganisationID: "org-1",
	}

	ns, err := s.PrepareNamespaceForStack(context.Background(), stack)
	if err != nil {
		t.Fatalf("PrepareNamespaceForStack returned error: %v", err)
	}
	if got, want := len(ns.Name), models.KubernetesDNSLabelMaxLength; got != want {
		t.Fatalf("generated namespace %q is %d characters, want exactly %d (name budget %d + separator + entropy floor %d)",
			ns.Name, got, want, models.MaxStackNameLength, models.MinNamespaceUUIDSuffixLength)
	}
	if !strings.HasPrefix(ns.Name, stack.Name+models.NamespaceNameSeparator) {
		t.Fatalf("generated namespace %q does not start with %q", ns.Name, stack.Name+models.NamespaceNameSeparator)
	}
	suffix := strings.TrimPrefix(ns.Name, stack.Name+models.NamespaceNameSeparator)
	if len(suffix) < models.MinNamespaceUUIDSuffixLength {
		t.Fatalf("generated namespace %q keeps only %d UUID characters, want at least %d",
			ns.Name, len(suffix), models.MinNamespaceUUIDSuffixLength)
	}
	if strings.HasSuffix(ns.Name, models.NamespaceNameSeparator) {
		t.Fatalf("generated namespace %q ends with %q, not a valid DNS label", ns.Name, models.NamespaceNameSeparator)
	}
}

// TestPrepareNamespaceForStackKeepsFullUUIDForShortNames pins backward
// compatibility: when the stack name is short enough that
// "<stack-name>-<uuid>" already fits the DNS-label cap, nothing is truncated
// and the suffix is a whole canonical UUID. If the UUID library's textual
// form ever drifts from models.NamespaceUUIDSuffixLength, this fails.
func TestPrepareNamespaceForStackKeepsFullUUIDForShortNames(t *testing.T) {
	s := &namespaceService{}
	stack := &models.Stack{
		Name:           "myapp",
		OrganisationID: "org-1",
	}

	ns, err := s.PrepareNamespaceForStack(context.Background(), stack)
	if err != nil {
		t.Fatalf("PrepareNamespaceForStack returned error: %v", err)
	}
	if got, want := len(ns.Name), len(stack.Name)+models.NamespaceUUIDSuffixLength; got != want {
		t.Fatalf("generated namespace %q is %d characters, want %d (full UUID suffix preserved)",
			ns.Name, got, want)
	}
	suffix := strings.TrimPrefix(ns.Name, stack.Name+models.NamespaceNameSeparator)
	if _, parseErr := uuid.Parse(suffix); parseErr != nil {
		t.Fatalf("generated namespace %q suffix %q is not a full canonical UUID: %v", ns.Name, suffix, parseErr)
	}
}
