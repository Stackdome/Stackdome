package models

import "time"

const (
	ManagedByLabelKey   = "stackdome.io/managed-by"
	ManagedByLabelValue = "stackdome"
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
