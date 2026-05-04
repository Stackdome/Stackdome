package models

import (
	"time"

	rbacv1 "k8s.io/api/rbac/v1"
)

const (
	CreatedForUserLabel = "user.stackdome.io/id"
)

type Role string

func (r Role) String() string {
	return string(r)
}

const (
	OrgAdminRole  Role = "OrgAdmin"
	DeveloperRole Role = "Developer"
	ViewerRole    Role = "Viewer"
	OrgMemberRole Role = "OrgMember"
)

type User struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Name           string
	Email          string `gorm:"unique"`
	Password       string
	Organisation   *Organisation `gorm:"foreignKey:OrganisationID"`
	Role           Role
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
