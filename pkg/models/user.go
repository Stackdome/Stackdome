package models

import (
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	CreatedForUserLabel = "user.stackdome.io/id"
)

// UserRole represents org-level roles stored on the User model.
// OrgAdmin = full org access. Empty string = permissions derived from project memberships.
// OrgMemberRole is not stored on User.Role; it is a Casbin-only grouping auto-managed by project membership lifecycle.
type UserRole string

func (r UserRole) String() string {
	return string(r)
}

const (
	NoRole        UserRole = ""
	OrgAdminRole  UserRole = "OrgAdmin"
	OrgMemberRole UserRole = "OrgMember"
)

type User struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Email          string `gorm:"unique"`
	Password       string
	Organisation   *Organisation `gorm:"foreignKey:OrganisationID"`
	Role           UserRole
	OrganisationID string
	GithubID       *string `gorm:"column:github_id;uniqueIndex"`
	AvatarURL      *string `gorm:"column:avatar_url"`
}

func (u *User) IsOrgAdmin() bool {
	return u.Role == OrgAdminRole
}

func (u *User) ClusterAccessRules() []rbacv1.PolicyRule {
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
}
