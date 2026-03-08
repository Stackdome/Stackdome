package presenters

import (
	"time"

	"github.com/ashishmax31/stackdome-api-server/pkg/api/openapi"
	"github.com/ashishmax31/stackdome-api-server/pkg/models"
)

// ConvertPostgresAddon converts API PostgresAddon to domain model
func ConvertPostgresAddon(in *openapi.PostgresAddon) *models.PostgresAddon {
	if in == nil {
		return nil
	}

	res := &models.PostgresAddon{
		Name:            in.GetName(),
		Labels:          convertLabels(in.GetLabels()),
		Annotations:     convertAnnotations(in.GetAnnotations()),
		PostgresVersion: convertPostgresVersion(in.Spec.GetVersion()),
		Instances:       convertPostgresInstances(in.Spec.GetInstances()),
		Resources:       convertPostgresResources(in.Spec.Resources),
		Storage:         convertPostgresStorage(in.Spec.GetStorage()),
		Configuration:   convertPostgresConfiguration(in.Spec.Configuration),
		Initialization:  convertPostgresInitialization(in.Spec.Initialization),
		BackupConfig:    convertPostgresBackupConfig(in.Spec.Backup),
		Databases:       convertPostgresDatabases(in.Spec.GetDatabases()),
	}

	return res
}

// PresentPostgresAddon converts domain model to API PostgresAddon
func PresentPostgresAddon(in *models.PostgresAddon) openapi.PostgresAddon {
	res := openapi.PostgresAddon{}

	res.SetId(in.ID)
	res.SetOrganisationId(in.OrganisationID)
	res.SetUserId(in.UserID)
	res.SetClusterId(in.ClusterID)
	res.SetName(in.Name)
	res.SetNamespace(in.Namespace)
	res.SetLabels(presentLabels(in.Labels))
	res.SetAnnotations(presentAnnotations(in.Annotations))
	res.SetRevision(in.Revision)
	res.SetCreatedAt(in.CreatedAt)
	res.SetUpdatedAt(in.UpdatedAt)

	// Set spec
	spec := openapi.PostgresAddonSpec{
		Version:   presentPostgresVersion(in.PostgresVersion),
		Instances: presentPostgresInstances(in.Instances),
		Storage:   presentPostgresStorage(in.Storage),
		Databases: presentPostgresDatabases(in.Databases),
	}

	if in.Resources != (models.PostgresResources{}) {
		spec.Resources = presentPostgresResources(in.Resources)
	}

	if in.Configuration.EnableSuperuserAccess || len(in.Configuration.Parameters) > 0 {
		spec.Configuration = presentPostgresConfiguration(in.Configuration)
	}

	if in.Initialization != (models.PostgresInitialization{}) {
		spec.Initialization = presentPostgresInitialization(in.Initialization)
	}

	if in.BackupConfig != (models.PostgresBackupConfig{}) {
		spec.Backup = presentPostgresBackupConfig(in.BackupConfig)
	}

	res.SetSpec(spec)

	// Set status if available
	if in.Status.State != "" || len(in.Status.Conditions) > 0 {
		res.SetStatus(presentPostgresAddonStatus(in.Status))
	}

	return res
}

// PresentPostgresAddonList converts list of domain models to API list
func PresentPostgresAddonList(in []*models.PostgresAddon) []openapi.PostgresAddon {
	if len(in) == 0 {
		return []openapi.PostgresAddon{}
	}

	result := make([]openapi.PostgresAddon, len(in))
	for i, addon := range in {
		result[i] = PresentPostgresAddon(addon)
	}
	return result
}

// Helper functions for nested conversions

func convertPostgresVersion(in openapi.PostgresVersion) models.PostgresVersion {
	return models.PostgresVersion{
		Major:                     int(in.GetMajor()),
		Minor:                     int(in.GetMinor()),
		EnableMajorVersionUpgrade: in.GetEnableAutoMajorUpgrade(),
		EnableMinorVersionUpgrade: in.GetEnableAutoMinorUpgrade(),
	}
}

func presentPostgresVersion(in models.PostgresVersion) openapi.PostgresVersion {
	res := openapi.PostgresVersion{}
	res.SetMajor(int32(in.Major))
	res.SetMinor(int32(in.Minor))
	res.SetEnableAutoMajorUpgrade(in.EnableMajorVersionUpgrade)
	res.SetEnableAutoMinorUpgrade(in.EnableMinorVersionUpgrade)
	return res
}

func convertPostgresInstances(in openapi.PostgresInstances) models.PostgresInstances {
	res := models.PostgresInstances{
		Count: int(in.GetCount()),
	}

	if in.Placement != nil {
		res.Placement = &models.PostgresInstancePlacement{
			TopologyKey:  in.Placement.GetTopologyKey(),
			Policy:       in.Placement.GetPolicy(),
			NodeSelector: in.Placement.GetNodeSelector(),
			Tolerations:  convertTolerations(in.Placement.GetTolerations()),
		}
	}

	return res
}

func presentPostgresInstances(in models.PostgresInstances) openapi.PostgresInstances {
	res := openapi.PostgresInstances{}
	res.SetCount(int32(in.Count))

	if in.Placement != nil {
		placement := openapi.PostgresInstancesPlacement{}
		placement.SetTopologyKey(in.Placement.TopologyKey)
		placement.SetPolicy(in.Placement.Policy)
		placement.SetNodeSelector(in.Placement.NodeSelector)
		placement.SetTolerations(presentTolerations(in.Placement.Tolerations))
		res.SetPlacement(placement)
	}

	return res
}

func convertTolerations(in []openapi.PostgresInstancesPlacementTolerationsInner) []models.PostgresToleration {
	if len(in) == 0 {
		return nil
	}

	result := make([]models.PostgresToleration, len(in))
	for i, toleration := range in {
		result[i] = models.PostgresToleration{
			Key:      toleration.GetKey(),
			Operator: toleration.GetOperator(),
			Value:    toleration.GetValue(),
			Effect:   toleration.GetEffect(),
		}
	}
	return result
}

func presentTolerations(in []models.PostgresToleration) []openapi.PostgresInstancesPlacementTolerationsInner {
	if len(in) == 0 {
		return nil
	}

	result := make([]openapi.PostgresInstancesPlacementTolerationsInner, len(in))
	for i, toleration := range in {
		item := openapi.PostgresInstancesPlacementTolerationsInner{}
		item.SetKey(toleration.Key)
		item.SetOperator(toleration.Operator)
		item.SetValue(toleration.Value)
		item.SetEffect(toleration.Effect)
		result[i] = item
	}
	return result
}

func convertPostgresResources(in *openapi.PostgresResources) models.PostgresResources {
	if in == nil {
		return models.PostgresResources{}
	}

	res := models.PostgresResources{}

	if in.Cpu != nil {
		res.CPU = models.PostgresCPUResource{
			Request: in.Cpu.GetRequest(),
			Limit:   in.Cpu.GetLimit(),
		}
	}

	if in.Memory != nil {
		res.Memory = models.PostgresMemoryResource{
			Request: in.Memory.GetRequest(),
			Limit:   in.Memory.GetLimit(),
		}
	}

	return res
}

func presentPostgresResources(in models.PostgresResources) *openapi.PostgresResources {
	res := &openapi.PostgresResources{}

	cpu := openapi.PostgresResourcesCpu{}
	cpu.SetRequest(in.CPU.Request)
	cpu.SetLimit(in.CPU.Limit)
	res.SetCpu(cpu)

	memory := openapi.PostgresResourcesMemory{}
	memory.SetRequest(in.Memory.Request)
	memory.SetLimit(in.Memory.Limit)
	res.SetMemory(memory)

	return res
}

func convertPostgresStorage(in openapi.PostgresStorage) models.PostgresStorage {
	return models.PostgresStorage{
		Size:         in.GetSize(),
		StorageClass: in.GetStorageClass(),
	}
}

func presentPostgresStorage(in models.PostgresStorage) openapi.PostgresStorage {
	res := openapi.PostgresStorage{}
	res.SetSize(in.Size)
	res.SetStorageClass(in.StorageClass)
	return res
}

func convertPostgresConfiguration(in *openapi.PostgresConfiguration) models.PostgresConfiguration {
	if in == nil {
		return models.PostgresConfiguration{}
	}

	return models.PostgresConfiguration{
		EnableSuperuserAccess: in.GetEnableSuperuserAccess(),
		Parameters:            in.GetParameters(),
	}
}

func presentPostgresConfiguration(in models.PostgresConfiguration) *openapi.PostgresConfiguration {
	res := &openapi.PostgresConfiguration{}
	res.SetEnableSuperuserAccess(in.EnableSuperuserAccess)
	res.SetParameters(in.Parameters)
	return res
}

func convertPostgresInitialization(in *openapi.PostgresInitialization) models.PostgresInitialization {
	if in == nil {
		return models.PostgresInitialization{}
	}

	res := models.PostgresInitialization{}

	if in.RestoreFromBackup != nil {
		res.RestoreFromBackup = &models.PostgresRestoreFromBackup{
			BackupID: in.RestoreFromBackup.GetBackupId(),
		}
		return res
	}

	if in.RestoreFromObjectStore != nil {
		res.RestoreFromObjectStore = &models.PostgresRestoreFromObjectStore{
			ObjectStoreID:         in.RestoreFromObjectStore.GetObjectStoreId(),
			SourcePostgresAddonID: in.RestoreFromObjectStore.GetSourcePostgresAddonId(),
			RecoveryTargetTime: func() *time.Time {
				if in.RestoreFromObjectStore.HasRecoveryTargetTime() {
					t := in.RestoreFromObjectStore.GetRecoveryTargetTime()
					return &t
				}
				return nil
			}(),
		}
		return res
	}

	if in.ImportFromExternal != nil {
		res.ImportFromExternal = &models.PostgresImportFromExternal{
			Host:              in.ImportFromExternal.GetHost(),
			Port:              int(in.ImportFromExternal.GetPort()),
			Database:          in.ImportFromExternal.GetDatabase(),
			Username:          in.ImportFromExternal.GetUsername(),
			PasswordSecretID:  in.ImportFromExternal.GetPasswordSecretId(),
			SslMode:           in.ImportFromExternal.SslMode,
			DatabasesToImport: in.ImportFromExternal.GetDatabasesToImport(),
		}
	}

	return res
}

func presentPostgresInitialization(in models.PostgresInitialization) *openapi.PostgresInitialization {
	res := &openapi.PostgresInitialization{}

	if in.RestoreFromBackup != nil {
		backup := openapi.PostgresInitializationRestoreFromBackup{}
		backup.SetBackupId(in.RestoreFromBackup.BackupID)
		res.SetRestoreFromBackup(backup)
	}

	if in.RestoreFromObjectStore != nil {
		objectStore := openapi.PostgresInitializationRestoreFromObjectStore{}
		objectStore.SetObjectStoreId(in.RestoreFromObjectStore.ObjectStoreID)
		objectStore.SetSourcePostgresAddonId(in.RestoreFromObjectStore.SourcePostgresAddonID)
		if in.RestoreFromObjectStore.RecoveryTargetTime != nil {
			objectStore.SetRecoveryTargetTime(*in.RestoreFromObjectStore.RecoveryTargetTime)
		}
		res.SetRestoreFromObjectStore(objectStore)
	}

	if in.ImportFromExternal != nil {
		external := openapi.PostgresInitializationImportFromExternal{}
		external.SetHost(in.ImportFromExternal.Host)
		external.SetPort(int32(in.ImportFromExternal.Port))
		external.SetDatabase(in.ImportFromExternal.Database)
		external.SetUsername(in.ImportFromExternal.Username)
		external.SetPasswordSecretId(in.ImportFromExternal.PasswordSecretID)
		if in.ImportFromExternal.SslMode != nil {
			external.SetSslMode(*in.ImportFromExternal.SslMode)
		}
		if len(in.ImportFromExternal.DatabasesToImport) > 0 {
			external.SetDatabasesToImport(in.ImportFromExternal.DatabasesToImport)
		}
		res.SetImportFromExternal(external)
	}

	return res
}

func convertPostgresBackupConfig(in *openapi.PostgresBackupConfig) models.PostgresBackupConfig {
	if in == nil || !in.GetEnabled() {
		return models.PostgresBackupConfig{}
	}

	s := models.PostgresBackupConfig{
		ObjectStoreID: in.GetObjectStoreId(),
		Schedule:      in.GetSchedule(),
		WALArchiving:  in.GetWalArchiving(),
	}
	return s
}

func presentPostgresBackupConfig(in models.PostgresBackupConfig) *openapi.PostgresBackupConfig {
	res := &openapi.PostgresBackupConfig{}
	res.SetObjectStoreId(in.ObjectStoreID)
	res.SetSchedule(in.Schedule)
	res.SetWalArchiving(in.WALArchiving)
	return res
}

func convertPostgresDatabases(in []openapi.PostgresDatabase) []models.PostgresAddonDatabase {
	if len(in) == 0 {
		return nil
	}

	result := make([]models.PostgresAddonDatabase, len(in))
	for i, db := range in {
		result[i] = models.PostgresAddonDatabase{
			Name:       db.GetName(),
			Extensions: models.PostgresExtensions(db.GetExtensions()),
		}
	}
	return result
}

func presentPostgresDatabases(in []models.PostgresAddonDatabase) []openapi.PostgresDatabase {
	if len(in) == 0 {
		return nil
	}

	result := make([]openapi.PostgresDatabase, len(in))
	for i, db := range in {
		item := openapi.PostgresDatabase{}
		item.SetName(db.Name)
		item.SetExtensions([]string(db.Extensions))
		result[i] = item
	}
	return result
}

func presentPostgresAddonStatus(in models.PostgresAddonStatus) openapi.PostgresAddonStatus {
	res := openapi.PostgresAddonStatus{}
	res.SetState(in.State)
	res.SetMessage(in.Message)

	if len(in.Conditions) > 0 {
		res.SetConditions(presentConditions(in.Conditions))
	}

	if in.ClusterInfo != nil {
		clusterInfo := openapi.PostgresClusterInfo{}
		// Note: API only has Version field, domain model has different fields
		// This might need to be updated when API spec is aligned
		res.SetClusterInfo(clusterInfo)
	}

	if in.ConnectionInfo != nil {
		connectionInfo := openapi.PostgresConnectionInfo{}
		connectionInfo.SetHost(in.ConnectionInfo.Host)
		connectionInfo.SetPort(int32(in.ConnectionInfo.Port))

		if in.ConnectionInfo.Credentials != (models.PostgresCredentials{}) {
			credentials := openapi.PostgresConnectionInfoCredentials{}
			// Note: API credentials structure is different from domain model
			// Domain has username/password, API has secret IDs
			connectionInfo.SetCredentials(credentials)
		}

		if len(in.ConnectionInfo.Databases) > 0 {
			databases := make([]openapi.PostgresConnectionInfoDatabasesInner, len(in.ConnectionInfo.Databases))
			for i, db := range in.ConnectionInfo.Databases {
				databases[i] = openapi.PostgresConnectionInfoDatabasesInner{
					Name:  &db.Name,
					Owner: nil, // Not available in domain model
				}
			}
			connectionInfo.SetDatabases(databases)
		}

		res.SetConnectionInfo(connectionInfo)
	}

	return res
}
