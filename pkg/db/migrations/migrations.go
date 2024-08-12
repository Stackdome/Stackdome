package migrations

import "github.com/go-gormigrate/gormigrate/v2"

var MigrationList = []*gormigrate.Migration{
	createUserAndOrganisationTable(),
	createClustersTable(),
	createWorkspaceStorageAndVolumeTables(),
	addManagerRunningColumnToCluster(),
	createWorkspaceUserTable(),
	createWorkspaceNamespaceTable(),
	alterWorkspaceUserTableAddVersioningSupport(),
}
