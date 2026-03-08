# PostgresBackupConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | Pointer to **bool** | Enable backup functionality | [optional] [default to false]
**ObjectStoreId** | Pointer to **string** | Reference to ObjectStore for backup storage | [optional] 
**Schedule** | Pointer to **string** | Cron schedule for automated backups | [optional] [default to "0 0 0 * * 0"]
**WalArchiving** | Pointer to **bool** | Enable WAL archiving for point-in-time recovery | [optional] [default to true]

## Methods

### NewPostgresBackupConfig

`func NewPostgresBackupConfig() *PostgresBackupConfig`

NewPostgresBackupConfig instantiates a new PostgresBackupConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresBackupConfigWithDefaults

`func NewPostgresBackupConfigWithDefaults() *PostgresBackupConfig`

NewPostgresBackupConfigWithDefaults instantiates a new PostgresBackupConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *PostgresBackupConfig) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *PostgresBackupConfig) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *PostgresBackupConfig) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.

### HasEnabled

`func (o *PostgresBackupConfig) HasEnabled() bool`

HasEnabled returns a boolean if a field has been set.

### GetObjectStoreId

`func (o *PostgresBackupConfig) GetObjectStoreId() string`

GetObjectStoreId returns the ObjectStoreId field if non-nil, zero value otherwise.

### GetObjectStoreIdOk

`func (o *PostgresBackupConfig) GetObjectStoreIdOk() (*string, bool)`

GetObjectStoreIdOk returns a tuple with the ObjectStoreId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObjectStoreId

`func (o *PostgresBackupConfig) SetObjectStoreId(v string)`

SetObjectStoreId sets ObjectStoreId field to given value.

### HasObjectStoreId

`func (o *PostgresBackupConfig) HasObjectStoreId() bool`

HasObjectStoreId returns a boolean if a field has been set.

### GetSchedule

`func (o *PostgresBackupConfig) GetSchedule() string`

GetSchedule returns the Schedule field if non-nil, zero value otherwise.

### GetScheduleOk

`func (o *PostgresBackupConfig) GetScheduleOk() (*string, bool)`

GetScheduleOk returns a tuple with the Schedule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSchedule

`func (o *PostgresBackupConfig) SetSchedule(v string)`

SetSchedule sets Schedule field to given value.

### HasSchedule

`func (o *PostgresBackupConfig) HasSchedule() bool`

HasSchedule returns a boolean if a field has been set.

### GetWalArchiving

`func (o *PostgresBackupConfig) GetWalArchiving() bool`

GetWalArchiving returns the WalArchiving field if non-nil, zero value otherwise.

### GetWalArchivingOk

`func (o *PostgresBackupConfig) GetWalArchivingOk() (*bool, bool)`

GetWalArchivingOk returns a tuple with the WalArchiving field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWalArchiving

`func (o *PostgresBackupConfig) SetWalArchiving(v bool)`

SetWalArchiving sets WalArchiving field to given value.

### HasWalArchiving

`func (o *PostgresBackupConfig) HasWalArchiving() bool`

HasWalArchiving returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


