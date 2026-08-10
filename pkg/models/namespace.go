package models

import "time"

const (
	ManagedByLabelKey     = "stackdome.io/managed-by"
	ManagedByLabelValue   = "stackdome"
	CloudTenantLabelKey   = "stackdome.io/cloud-tenant"
	CloudTenantLabelValue = "true"
	// CloudPolicyReadyLabelKey is set by the private cloud guard only after it
	// verifies the namespace's exact quota, limit range, and network policies.
	CloudPolicyReadyLabelKey = "stackdome.io/cloud-policy-ready"
	OrganizationIDLabelKey   = "stackdome.io/organization-id"
	NamespaceRoleLabelKey    = "stackdome.io/namespace-role"
	NamespaceRoleStack       = "stack"
	NamespaceRoleAddon       = "addon"
)

// Stack namespaces are generated as "<stack-name>-<uuid>" and then truncated
// from the end to the DNS-label cap (see PrepareNamespaceForStack in
// pkg/services/namespace_service.go). A Kubernetes namespace name is an RFC
// 1123 DNS label capped at 63 characters; truncation cuts the UUID tail, so
// the stack name only has to leave room for the separator plus the minimum
// surviving slice of UUID entropy — not the full UUID. Namespace generation
// and stack-name validation both derive from these constants — they are the
// single source of truth for that budget.
const (
	// KubernetesDNSLabelMaxLength is the RFC 1123 DNS-label cap Kubernetes
	// applies to namespace names and label values.
	KubernetesDNSLabelMaxLength = 63
	// NamespaceNameSeparator joins the stack name and the UUID suffix in a
	// generated namespace name.
	NamespaceNameSeparator = "-"
	// NamespaceUUIDLength is the length of canonical RFC 4122 UUID text (as
	// produced by uuid.UUID.String()).
	NamespaceUUIDLength = 36
	// NamespaceUUIDSuffixLength is the length of the full
	// "<separator><uuid>" suffix appended to the stack name before
	// truncation. Short stack names keep it whole.
	NamespaceUUIDSuffixLength = len(NamespaceNameSeparator) + NamespaceUUIDLength
	// MinNamespaceUUIDSuffixLength is the entropy floor: at least this many
	// UUID characters must survive truncation so generated namespace names
	// stay effectively collision-free.
	MinNamespaceUUIDSuffixLength = 8
	// MaxStackNameLength is the stack-name budget inside a generated
	// namespace name: the DNS-label cap minus the separator and the entropy
	// floor (63 - 1 - 8 = 54).
	MaxStackNameLength = KubernetesDNSLabelMaxLength - len(NamespaceNameSeparator) - MinNamespaceUUIDSuffixLength
)

// Addon namespaces are generated as
// "stackdome-addons-<type>-<name>-<uuid>" and truncated from the end to the
// DNS-label cap (see PrepareNamespaceForAddon in
// pkg/services/namespace_service.go). These constants are the single source
// of truth for the addon-name budget, mirroring the stack-name budget above.
const (
	// AddonNamespacePrefix prefixes every generated addon namespace name.
	AddonNamespacePrefix = "stackdome-addons"
	// MaxAddonTypeLength reserves room for the longest addon type slug
	// ("postgres") in the addon-name budget.
	MaxAddonTypeLength = 8
	// MaxAddonNameLength is the addon-name budget inside a generated
	// namespace name: the DNS-label cap minus the prefix, the type reservation,
	// the three separators (prefix-type, type-name, name-uuid), and the entropy
	// floor (63 - 16 - 8 - 3 - 8 = 28).
	MaxAddonNameLength = KubernetesDNSLabelMaxLength - len(AddonNamespacePrefix) - MaxAddonTypeLength - 3*len(NamespaceNameSeparator) - MinNamespaceUUIDSuffixLength
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

// AddSharedComputeTenantLabels preserves the legacy cloud-tenant label because
// the hardening controller uses it to select tenant namespaces.
func (n *Namespace) AddSharedComputeTenantLabels(role string) {
	n.Labels = append(n.Labels,
		Label{Key: CloudTenantLabelKey, Value: CloudTenantLabelValue},
		Label{Key: OrganizationIDLabelKey, Value: n.OrganisationID},
		Label{Key: NamespaceRoleLabelKey, Value: role},
	)
}
