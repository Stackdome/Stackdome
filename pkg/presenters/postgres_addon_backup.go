package presenters

import (
	"github.com/Stackdome/stackdome/pkg/api/openapi"
	"github.com/Stackdome/stackdome/pkg/models"
)

// PresentPostgresAddonBackup converts domain model to API PostgresBackup
// Note: No Convert function needed since backups are triggered via action endpoints
func PresentPostgresAddonBackup(in *models.PostgresBackup) openapi.PostgresBackup {
	res := openapi.PostgresBackup{}

	res.SetId(in.ID)
	res.SetPostgresAddonId(in.PostgresAddonID)
	res.SetName(in.Name)
	res.SetDescription(in.Description)
	res.SetType(string(in.Type))
	res.SetPhase(in.Phase)
	res.SetError(in.Error)
	res.SetCreatedAt(in.CreatedAt)
	// Note: API doesn't have UpdatedAt field, only CreatedAt

	if in.StartedAt != nil {
		res.SetStartedAt(*in.StartedAt)
	}

	if in.CompletedAt != nil {
		res.SetCompletedAt(*in.CompletedAt)
	}

	if in.SizeBytes != nil {
		res.SetSizeBytes(*in.SizeBytes)
	}

	return res
}

// PresentPostgresAddonBackupList converts list of domain models to API list
func PresentPostgresAddonBackupList(in []*models.PostgresBackup) []openapi.PostgresBackup {
	if len(in) == 0 {
		return []openapi.PostgresBackup{}
	}

	result := make([]openapi.PostgresBackup, len(in))
	for i, backup := range in {
		result[i] = PresentPostgresAddonBackup(backup)
	}
	return result
}

// PresentPostgresBackupList converts list of domain models to API list
func PresentPostgresBackupList(in []*models.PostgresBackup) []openapi.PostgresBackup {
	if len(in) == 0 {
		return []openapi.PostgresBackup{}
	}

	result := make([]openapi.PostgresBackup, len(in))
	for i, backup := range in {
		result[i] = PresentPostgresAddonBackup(backup)
	}
	return result
}
