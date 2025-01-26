package models

import (
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
)

type Role string

func (r Role) String() string {
	return string(r)
}

const (
	UserRole              Role = "User"
	OrganisationAdminRole Role = "OrganisationAdmin"
	PlatformAdminRole     Role = "PlatformAdmin"
)

type User struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Email          string `gorm:"unique"`
	Password       string
	Organisation   string
	Role           Role
	OrganisationID string
	DefaultUser    bool
}

func (u *User) ClusterAccessRules() []rbacv1.PolicyRule {
	switch u.Role {
	case UserRole:
		return []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/exec", "pods/portforward"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get"},
			},
		}
	case PlatformAdminRole:
		return []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/exec", "pods/portforward"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get"},
			},
		}
	default:
		panic("not implemented")
	}
}
