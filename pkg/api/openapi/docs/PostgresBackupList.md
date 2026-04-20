# PostgresBackupList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | Pointer to [**[]PostgresBackup**](PostgresBackup.md) |  | [optional] 
**Total** | Pointer to **int32** |  | [optional] 

## Methods

### NewPostgresBackupList

`func NewPostgresBackupList() *PostgresBackupList`

NewPostgresBackupList instantiates a new PostgresBackupList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresBackupListWithDefaults

`func NewPostgresBackupListWithDefaults() *PostgresBackupList`

NewPostgresBackupListWithDefaults instantiates a new PostgresBackupList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *PostgresBackupList) GetItems() []PostgresBackup`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *PostgresBackupList) GetItemsOk() (*[]PostgresBackup, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *PostgresBackupList) SetItems(v []PostgresBackup)`

SetItems sets Items field to given value.

### HasItems

`func (o *PostgresBackupList) HasItems() bool`

HasItems returns a boolean if a field has been set.

### GetTotal

`func (o *PostgresBackupList) GetTotal() int32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *PostgresBackupList) GetTotalOk() (*int32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *PostgresBackupList) SetTotal(v int32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *PostgresBackupList) HasTotal() bool`

HasTotal returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


