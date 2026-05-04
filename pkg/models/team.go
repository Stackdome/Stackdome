package models

import "time"

type Team struct {
	ID             string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name           string `gorm:"not null" json:"name"`
	OrganisationID string `gorm:"not null" json:"organisation_id"`
	DefaultTeam    bool   `gorm:"not null;default:false" json:"default_team"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TeamMembership struct {
	ID        string `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	TeamID    string `gorm:"not null" json:"team_id"`
	UserID    string `gorm:"not null" json:"user_id"`
	Role      Role   `gorm:"not null" json:"role"`
	CreatedAt time.Time

	Team *Team `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
