package models

type WorkspaceNamespace struct {
	WorkspaceUserID string `gorm:"type:text;not null"`
	UserID          string `gorm:"type:text;not null;primaryKey"`
	Namespace       string `gorm:"type:text;uniqueIndex;not null"`
	Workspace       string `gorm:"type:text;not null; primaryKey"`
	Enabled         bool
}
