package models

import "time"

const (
	DefaultProjectName = "default"
)

type ProjectRole string

func (r ProjectRole) String() string {
	return string(r)
}

const (
	DeveloperRole ProjectRole = "Developer"
	ViewerRole    ProjectRole = "Viewer"
)

type Project struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name           string `gorm:"not null" json:"name"`
	OrganisationID string `gorm:"not null" json:"organisation_id"`
	DefaultProject    bool   `gorm:"not null;default:false" json:"default_project"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ProjectMembership struct {
	ID        string   `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	ProjectID    string   `gorm:"not null" json:"project_id"`
	UserID    string   `gorm:"not null" json:"user_id"`
	Role      ProjectRole `gorm:"not null" json:"role"`
	CreatedAt time.Time
	Project      *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	User      *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
