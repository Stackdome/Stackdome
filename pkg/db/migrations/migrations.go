package migrations

import "github.com/go-gormigrate/gormigrate/v2"

var MigrationList = []*gormigrate.Migration{
	createUserAndOrganisationTable(),
	createClustersTable(),
	createStackStorageAndVolumeTables(),
	addManagerRunningColumnToCluster(),
	createWorkspaceUserTable(),
	createWorkspaceNamespaceTable(),
	alterWorkspaceUserTableAddVersioningSupport(),
	alterWorkspaceStorageTable(),
	createStackAndStackResourceTables(),
	createImageBuildsTable(),
	addSourceVolumeTypeToVolumeMounts(),
	addStackResourceNameToImageBuilds(),
	createDefaultOrganisation(),
	addDefaultUserColumnToUsersTable(),
	addWorkspaceNameColumnToStacksTable(),
	removeOrganisationColumnFromUsers(),
	addNamespaceToStackResourcesTable(),
}
