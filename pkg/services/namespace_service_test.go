package services

import (
	"context"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/models"
	"github.com/google/uuid"
)

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
