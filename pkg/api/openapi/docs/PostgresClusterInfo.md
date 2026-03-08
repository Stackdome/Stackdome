# PostgresClusterInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Version** | Pointer to **string** | Current PostgreSQL version | [optional] 

## Methods

### NewPostgresClusterInfo

`func NewPostgresClusterInfo() *PostgresClusterInfo`

NewPostgresClusterInfo instantiates a new PostgresClusterInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresClusterInfoWithDefaults

`func NewPostgresClusterInfoWithDefaults() *PostgresClusterInfo`

NewPostgresClusterInfoWithDefaults instantiates a new PostgresClusterInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVersion

`func (o *PostgresClusterInfo) GetVersion() string`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *PostgresClusterInfo) GetVersionOk() (*string, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *PostgresClusterInfo) SetVersion(v string)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *PostgresClusterInfo) HasVersion() bool`

HasVersion returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


