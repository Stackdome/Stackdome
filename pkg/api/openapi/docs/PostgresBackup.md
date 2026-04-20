# PostgresBackup

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Unique backup identifier | [optional] 
**PostgresAddonId** | Pointer to **string** | ID of the PostgreSQL addon this backup belongs to | [optional] 
**Name** | Pointer to **string** | Human-readable backup name | [optional] 
**Description** | Pointer to **string** | Optional backup description | [optional] 
**Type** | Pointer to **string** | How the backup was initiated | [optional] 
**Phase** | Pointer to **string** | Current backup status | [optional] 
**StartedAt** | Pointer to **time.Time** | When the backup started | [optional] 
**CompletedAt** | Pointer to **time.Time** | When the backup completed (if finished) | [optional] 
**Error** | Pointer to **string** | Error message if backup failed | [optional] 
**SizeBytes** | Pointer to **int32** | Backup size in bytes (if available) | [optional] 
**CreatedAt** | Pointer to **time.Time** | When the backup record was created | [optional] 

## Methods

### NewPostgresBackup

`func NewPostgresBackup() *PostgresBackup`

NewPostgresBackup instantiates a new PostgresBackup object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresBackupWithDefaults

`func NewPostgresBackupWithDefaults() *PostgresBackup`

NewPostgresBackupWithDefaults instantiates a new PostgresBackup object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *PostgresBackup) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PostgresBackup) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PostgresBackup) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *PostgresBackup) HasId() bool`

HasId returns a boolean if a field has been set.

### GetPostgresAddonId

`func (o *PostgresBackup) GetPostgresAddonId() string`

GetPostgresAddonId returns the PostgresAddonId field if non-nil, zero value otherwise.

### GetPostgresAddonIdOk

`func (o *PostgresBackup) GetPostgresAddonIdOk() (*string, bool)`

GetPostgresAddonIdOk returns a tuple with the PostgresAddonId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPostgresAddonId

`func (o *PostgresBackup) SetPostgresAddonId(v string)`

SetPostgresAddonId sets PostgresAddonId field to given value.

### HasPostgresAddonId

`func (o *PostgresBackup) HasPostgresAddonId() bool`

HasPostgresAddonId returns a boolean if a field has been set.

### GetName

`func (o *PostgresBackup) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PostgresBackup) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PostgresBackup) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *PostgresBackup) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDescription

`func (o *PostgresBackup) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *PostgresBackup) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *PostgresBackup) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *PostgresBackup) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetType

`func (o *PostgresBackup) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *PostgresBackup) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *PostgresBackup) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *PostgresBackup) HasType() bool`

HasType returns a boolean if a field has been set.

### GetPhase

`func (o *PostgresBackup) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PostgresBackup) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PostgresBackup) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *PostgresBackup) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetStartedAt

`func (o *PostgresBackup) GetStartedAt() time.Time`

GetStartedAt returns the StartedAt field if non-nil, zero value otherwise.

### GetStartedAtOk

`func (o *PostgresBackup) GetStartedAtOk() (*time.Time, bool)`

GetStartedAtOk returns a tuple with the StartedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartedAt

`func (o *PostgresBackup) SetStartedAt(v time.Time)`

SetStartedAt sets StartedAt field to given value.

### HasStartedAt

`func (o *PostgresBackup) HasStartedAt() bool`

HasStartedAt returns a boolean if a field has been set.

### GetCompletedAt

`func (o *PostgresBackup) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *PostgresBackup) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *PostgresBackup) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *PostgresBackup) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.

### GetError

`func (o *PostgresBackup) GetError() string`

GetError returns the Error field if non-nil, zero value otherwise.

### GetErrorOk

`func (o *PostgresBackup) GetErrorOk() (*string, bool)`

GetErrorOk returns a tuple with the Error field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetError

`func (o *PostgresBackup) SetError(v string)`

SetError sets Error field to given value.

### HasError

`func (o *PostgresBackup) HasError() bool`

HasError returns a boolean if a field has been set.

### GetSizeBytes

`func (o *PostgresBackup) GetSizeBytes() int32`

GetSizeBytes returns the SizeBytes field if non-nil, zero value otherwise.

### GetSizeBytesOk

`func (o *PostgresBackup) GetSizeBytesOk() (*int32, bool)`

GetSizeBytesOk returns a tuple with the SizeBytes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSizeBytes

`func (o *PostgresBackup) SetSizeBytes(v int32)`

SetSizeBytes sets SizeBytes field to given value.

### HasSizeBytes

`func (o *PostgresBackup) HasSizeBytes() bool`

HasSizeBytes returns a boolean if a field has been set.

### GetCreatedAt

`func (o *PostgresBackup) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *PostgresBackup) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *PostgresBackup) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *PostgresBackup) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


