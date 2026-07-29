# PostgresStorage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | **string** | Storage size (e.g., 10Gi, 100Gi) | 
**StorageClass** | Pointer to **string** | Kubernetes storage class name. Omit to use the cluster&#39;s default storage class. | [optional] 

## Methods

### NewPostgresStorage

`func NewPostgresStorage(size string, ) *PostgresStorage`

NewPostgresStorage instantiates a new PostgresStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresStorageWithDefaults

`func NewPostgresStorageWithDefaults() *PostgresStorage`

NewPostgresStorageWithDefaults instantiates a new PostgresStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSize

`func (o *PostgresStorage) GetSize() string`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *PostgresStorage) GetSizeOk() (*string, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *PostgresStorage) SetSize(v string)`

SetSize sets Size field to given value.


### GetStorageClass

`func (o *PostgresStorage) GetStorageClass() string`

GetStorageClass returns the StorageClass field if non-nil, zero value otherwise.

### GetStorageClassOk

`func (o *PostgresStorage) GetStorageClassOk() (*string, bool)`

GetStorageClassOk returns a tuple with the StorageClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClass

`func (o *PostgresStorage) SetStorageClass(v string)`

SetStorageClass sets StorageClass field to given value.

### HasStorageClass

`func (o *PostgresStorage) HasStorageClass() bool`

HasStorageClass returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


