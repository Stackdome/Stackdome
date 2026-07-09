package models

import "time"

const (
	ManagedByLabelKey   = "stackdome.io/managed-by"
	ManagedByLabelValue = "stackdome"
)

// Stack namespaces are generated as "<stack-name>-<uuid>" (see
// PrepareNamespaceForStack in pkg/services/namespace_service.go). A
// Kubernetes namespace name is an RFC 1123 DNS label capped at 63
// characters, so the stack name only gets whatever budget the UUID suffix
// leaves over. Namespace generation and stack-name validation both derive
// from these constants — they are the single source of truth for that
// budget.
const (
	// KubernetesDNSLabelMaxLength is the RFC 1123 DNS-label cap Kubernetes
	// applies to namespace names and label values.
	KubernetesDNSLabelMaxLength = 63
	// NamespaceUUIDSuffixLength is the length of the "-<uuid>" suffix
	// appended to the stack name: 1 separator + 36 characters of canonical
	// RFC 4122 UUID text (as produced by uuid.UUID.String()).
	NamespaceUUIDSuffixLength = 1 + 36
	// MaxStackNameLength is the stack-name budget left inside a generated
	// namespace name (63 - 37 = 26).
	MaxStackNameLength = KubernetesDNSLabelMaxLength - NamespaceUUIDSuffixLength
)

type Namespace struct {
	ID             string      `gorm:"primary_key;default:gen_random_uuid()"`
	Name           string      `gorm:"not null;unique"`
	OrganisationID string      `gorm:"not null"`
	Labels         Labels      `gorm:"type:jsonb"`
	Annotations    Annotations `gorm:"type:jsonb"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (n *Namespace) AddDefaultLabels() {
	if n.Labels == nil {
		n.Labels = make(Labels, 0)
	}
	n.Labels = append(n.Labels, Label{
		Key:   ManagedByLabelKey,
		Value: ManagedByLabelValue,
	})
}
