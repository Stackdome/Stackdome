# PostgresInitialization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to **string** |  | [optional] [default to "new"]
**RestoreFromBackup** | Pointer to [**PostgresInitializationRestoreFromBackup**](PostgresInitializationRestoreFromBackup.md) |  | [optional] 
**RestoreFromObjectStore** | Pointer to [**PostgresInitializationRestoreFromObjectStore**](PostgresInitializationRestoreFromObjectStore.md) |  | [optional] 
**ImportFromExternal** | Pointer to [**PostgresInitializationImportFromExternal**](PostgresInitializationImportFromExternal.md) |  | [optional] 

## Methods

### NewPostgresInitialization

`func NewPostgresInitialization() *PostgresInitialization`

NewPostgresInitialization instantiates a new PostgresInitialization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresInitializationWithDefaults

`func NewPostgresInitializationWithDefaults() *PostgresInitialization`

NewPostgresInitializationWithDefaults instantiates a new PostgresInitialization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *PostgresInitialization) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *PostgresInitialization) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *PostgresInitialization) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *PostgresInitialization) HasType() bool`

HasType returns a boolean if a field has been set.

### GetRestoreFromBackup

`func (o *PostgresInitialization) GetRestoreFromBackup() PostgresInitializationRestoreFromBackup`

GetRestoreFromBackup returns the RestoreFromBackup field if non-nil, zero value otherwise.

### GetRestoreFromBackupOk

`func (o *PostgresInitialization) GetRestoreFromBackupOk() (*PostgresInitializationRestoreFromBackup, bool)`

GetRestoreFromBackupOk returns a tuple with the RestoreFromBackup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestoreFromBackup

`func (o *PostgresInitialization) SetRestoreFromBackup(v PostgresInitializationRestoreFromBackup)`

SetRestoreFromBackup sets RestoreFromBackup field to given value.

### HasRestoreFromBackup

`func (o *PostgresInitialization) HasRestoreFromBackup() bool`

HasRestoreFromBackup returns a boolean if a field has been set.

### GetRestoreFromObjectStore

`func (o *PostgresInitialization) GetRestoreFromObjectStore() PostgresInitializationRestoreFromObjectStore`

GetRestoreFromObjectStore returns the RestoreFromObjectStore field if non-nil, zero value otherwise.

### GetRestoreFromObjectStoreOk

`func (o *PostgresInitialization) GetRestoreFromObjectStoreOk() (*PostgresInitializationRestoreFromObjectStore, bool)`

GetRestoreFromObjectStoreOk returns a tuple with the RestoreFromObjectStore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestoreFromObjectStore

`func (o *PostgresInitialization) SetRestoreFromObjectStore(v PostgresInitializationRestoreFromObjectStore)`

SetRestoreFromObjectStore sets RestoreFromObjectStore field to given value.

### HasRestoreFromObjectStore

`func (o *PostgresInitialization) HasRestoreFromObjectStore() bool`

HasRestoreFromObjectStore returns a boolean if a field has been set.

### GetImportFromExternal

`func (o *PostgresInitialization) GetImportFromExternal() PostgresInitializationImportFromExternal`

GetImportFromExternal returns the ImportFromExternal field if non-nil, zero value otherwise.

### GetImportFromExternalOk

`func (o *PostgresInitialization) GetImportFromExternalOk() (*PostgresInitializationImportFromExternal, bool)`

GetImportFromExternalOk returns a tuple with the ImportFromExternal field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImportFromExternal

`func (o *PostgresInitialization) SetImportFromExternal(v PostgresInitializationImportFromExternal)`

SetImportFromExternal sets ImportFromExternal field to given value.

### HasImportFromExternal

`func (o *PostgresInitialization) HasImportFromExternal() bool`

HasImportFromExternal returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


