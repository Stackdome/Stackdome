# PostgresAddonSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Version** | [**PostgresVersion**](PostgresVersion.md) |  | 
**Instances** | [**PostgresInstances**](PostgresInstances.md) |  | 
**Storage** | [**PostgresStorage**](PostgresStorage.md) |  | 
**Resources** | Pointer to [**PostgresResources**](PostgresResources.md) |  | [optional] 
**Backup** | Pointer to [**PostgresBackupConfig**](PostgresBackupConfig.md) |  | [optional] 
**Initialization** | Pointer to [**PostgresInitialization**](PostgresInitialization.md) |  | [optional] 
**Databases** | Pointer to [**[]PostgresDatabase**](PostgresDatabase.md) |  | [optional] 
**Configuration** | Pointer to [**PostgresConfiguration**](PostgresConfiguration.md) |  | [optional] 

## Methods

### NewPostgresAddonSpec

`func NewPostgresAddonSpec(version PostgresVersion, instances PostgresInstances, storage PostgresStorage, ) *PostgresAddonSpec`

NewPostgresAddonSpec instantiates a new PostgresAddonSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresAddonSpecWithDefaults

`func NewPostgresAddonSpecWithDefaults() *PostgresAddonSpec`

NewPostgresAddonSpecWithDefaults instantiates a new PostgresAddonSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVersion

`func (o *PostgresAddonSpec) GetVersion() PostgresVersion`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *PostgresAddonSpec) GetVersionOk() (*PostgresVersion, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *PostgresAddonSpec) SetVersion(v PostgresVersion)`

SetVersion sets Version field to given value.


### GetInstances

`func (o *PostgresAddonSpec) GetInstances() PostgresInstances`

GetInstances returns the Instances field if non-nil, zero value otherwise.

### GetInstancesOk

`func (o *PostgresAddonSpec) GetInstancesOk() (*PostgresInstances, bool)`

GetInstancesOk returns a tuple with the Instances field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstances

`func (o *PostgresAddonSpec) SetInstances(v PostgresInstances)`

SetInstances sets Instances field to given value.


### GetStorage

`func (o *PostgresAddonSpec) GetStorage() PostgresStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *PostgresAddonSpec) GetStorageOk() (*PostgresStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *PostgresAddonSpec) SetStorage(v PostgresStorage)`

SetStorage sets Storage field to given value.


### GetResources

`func (o *PostgresAddonSpec) GetResources() PostgresResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *PostgresAddonSpec) GetResourcesOk() (*PostgresResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *PostgresAddonSpec) SetResources(v PostgresResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *PostgresAddonSpec) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetBackup

`func (o *PostgresAddonSpec) GetBackup() PostgresBackupConfig`

GetBackup returns the Backup field if non-nil, zero value otherwise.

### GetBackupOk

`func (o *PostgresAddonSpec) GetBackupOk() (*PostgresBackupConfig, bool)`

GetBackupOk returns a tuple with the Backup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackup

`func (o *PostgresAddonSpec) SetBackup(v PostgresBackupConfig)`

SetBackup sets Backup field to given value.

### HasBackup

`func (o *PostgresAddonSpec) HasBackup() bool`

HasBackup returns a boolean if a field has been set.

### GetInitialization

`func (o *PostgresAddonSpec) GetInitialization() PostgresInitialization`

GetInitialization returns the Initialization field if non-nil, zero value otherwise.

### GetInitializationOk

`func (o *PostgresAddonSpec) GetInitializationOk() (*PostgresInitialization, bool)`

GetInitializationOk returns a tuple with the Initialization field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitialization

`func (o *PostgresAddonSpec) SetInitialization(v PostgresInitialization)`

SetInitialization sets Initialization field to given value.

### HasInitialization

`func (o *PostgresAddonSpec) HasInitialization() bool`

HasInitialization returns a boolean if a field has been set.

### GetDatabases

`func (o *PostgresAddonSpec) GetDatabases() []PostgresDatabase`

GetDatabases returns the Databases field if non-nil, zero value otherwise.

### GetDatabasesOk

`func (o *PostgresAddonSpec) GetDatabasesOk() (*[]PostgresDatabase, bool)`

GetDatabasesOk returns a tuple with the Databases field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatabases

`func (o *PostgresAddonSpec) SetDatabases(v []PostgresDatabase)`

SetDatabases sets Databases field to given value.

### HasDatabases

`func (o *PostgresAddonSpec) HasDatabases() bool`

HasDatabases returns a boolean if a field has been set.

### GetConfiguration

`func (o *PostgresAddonSpec) GetConfiguration() PostgresConfiguration`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *PostgresAddonSpec) GetConfigurationOk() (*PostgresConfiguration, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *PostgresAddonSpec) SetConfiguration(v PostgresConfiguration)`

SetConfiguration sets Configuration field to given value.

### HasConfiguration

`func (o *PostgresAddonSpec) HasConfiguration() bool`

HasConfiguration returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


