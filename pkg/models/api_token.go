package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringSlice) Scan(src any) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, s)
	case string:
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}

type APIToken struct {
	ID          string      `gorm:"primary_key;default:gen_random_uuid()" json:"id"`
	Name        string      `gorm:"not null" json:"name"`
	UserID      string      `gorm:"not null" json:"userId"`
	TokenHash   string      `gorm:"not null" json:"-"`
	TokenPrefix string      `gorm:"not null" json:"tokenPrefix"`
	Scopes      StringSlice `gorm:"type:jsonb;not null" json:"scopes"`
	ResourceIDs StringSlice `gorm:"type:jsonb" json:"resourceIds,omitempty"`
	OrgID       string      `gorm:"not null" json:"orgId"`
	ExpiresAt   *time.Time  `json:"expiresAt,omitempty"`
	LastUsedAt  *time.Time  `json:"lastUsedAt,omitempty"`
	CreatedAt   time.Time   `gorm:"not null" json:"createdAt"`
	RevokedAt   *time.Time  `json:"revokedAt,omitempty"`
}

func (APIToken) TableName() string {
	return "api_tokens"
}
