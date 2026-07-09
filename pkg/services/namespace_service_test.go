package services

import (
	"context"
	"strings"
	"testing"

	"github.com/Stackdome/stackdome/pkg/models"
)

// TestPrepareNamespaceForStackFitsDNSLabelAtMaxNameLength pins the contract
// behind models.MaxStackNameLength: a stack name at the validator's cap must
// generate a namespace name of exactly models.KubernetesDNSLabelMaxLength
// characters. If the "<stack-name>-<uuid>" format or the UUID library's
// textual form ever drifts from models.NamespaceUUIDSuffixLength, this fails.
func TestPrepareNamespaceForStackFitsDNSLabelAtMaxNameLength(t *testing.T) {
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
		t.Fatalf("generated namespace %q is %d characters, want exactly %d (name budget %d + uuid suffix %d)",
			ns.Name, got, want, models.MaxStackNameLength, models.NamespaceUUIDSuffixLength)
	}
	if !strings.HasPrefix(ns.Name, stack.Name+"-") {
		t.Fatalf("generated namespace %q does not start with %q", ns.Name, stack.Name+"-")
	}
}
