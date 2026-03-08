# PostgresInitializationRestoreFromObjectStore

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ObjectStoreId** | Pointer to **string** | Object store containing the backup | [optional] 
**SourcePostgresAddonId** | Pointer to **string** | Source PostgresAddon that created the backup | [optional] 
**RecoveryTargetTime** | Pointer to **time.Time** | Point-in-time recovery target | [optional] 

## Methods

### NewPostgresInitializationRestoreFromObjectStore

`func NewPostgresInitializationRestoreFromObjectStore() *PostgresInitializationRestoreFromObjectStore`

NewPostgresInitializationRestoreFromObjectStore instantiates a new PostgresInitializationRestoreFromObjectStore object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresInitializationRestoreFromObjectStoreWithDefaults

`func NewPostgresInitializationRestoreFromObjectStoreWithDefaults() *PostgresInitializationRestoreFromObjectStore`

NewPostgresInitializationRestoreFromObjectStoreWithDefaults instantiates a new PostgresInitializationRestoreFromObjectStore object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetObjectStoreId

`func (o *PostgresInitializationRestoreFromObjectStore) GetObjectStoreId() string`

GetObjectStoreId returns the ObjectStoreId field if non-nil, zero value otherwise.

### GetObjectStoreIdOk

`func (o *PostgresInitializationRestoreFromObjectStore) GetObjectStoreIdOk() (*string, bool)`

GetObjectStoreIdOk returns a tuple with the ObjectStoreId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectStoreId

`func (o *PostgresInitializationRestoreFromObjectStore) SetObjectStoreId(v string)`

SetObjectStoreId sets ObjectStoreId field to given value.

### HasObjectStoreId

`func (o *PostgresInitializationRestoreFromObjectStore) HasObjectStoreId() bool`

HasObjectStoreId returns a boolean if a field has been set.

### GetSourcePostgresAddonId

`func (o *PostgresInitializationRestoreFromObjectStore) GetSourcePostgresAddonId() string`

GetSourcePostgresAddonId returns the SourcePostgresAddonId field if non-nil, zero value otherwise.

### GetSourcePostgresAddonIdOk

`func (o *PostgresInitializationRestoreFromObjectStore) GetSourcePostgresAddonIdOk() (*string, bool)`

GetSourcePostgresAddonIdOk returns a tuple with the SourcePostgresAddonId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourcePostgresAddonId

`func (o *PostgresInitializationRestoreFromObjectStore) SetSourcePostgresAddonId(v string)`

SetSourcePostgresAddonId sets SourcePostgresAddonId field to given value.

### HasSourcePostgresAddonId

`func (o *PostgresInitializationRestoreFromObjectStore) HasSourcePostgresAddonId() bool`

HasSourcePostgresAddonId returns a boolean if a field has been set.

### GetRecoveryTargetTime

`func (o *PostgresInitializationRestoreFromObjectStore) GetRecoveryTargetTime() time.Time`

GetRecoveryTargetTime returns the RecoveryTargetTime field if non-nil, zero value otherwise.

### GetRecoveryTargetTimeOk

`func (o *PostgresInitializationRestoreFromObjectStore) GetRecoveryTargetTimeOk() (*time.Time, bool)`

GetRecoveryTargetTimeOk returns a tuple with the RecoveryTargetTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRecoveryTargetTime

`func (o *PostgresInitializationRestoreFromObjectStore) SetRecoveryTargetTime(v time.Time)`

SetRecoveryTargetTime sets RecoveryTargetTime field to given value.

### HasRecoveryTargetTime

`func (o *PostgresInitializationRestoreFromObjectStore) HasRecoveryTargetTime() bool`

HasRecoveryTargetTime returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


