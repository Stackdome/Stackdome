# PostgresDatabase

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Database name | 
**Extensions** | Pointer to **[]string** | PostgreSQL extensions to enable | [optional] 

## Methods

### NewPostgresDatabase

`func NewPostgresDatabase(name string, ) *PostgresDatabase`

NewPostgresDatabase instantiates a new PostgresDatabase object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresDatabaseWithDefaults

`func NewPostgresDatabaseWithDefaults() *PostgresDatabase`

NewPostgresDatabaseWithDefaults instantiates a new PostgresDatabase object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *PostgresDatabase) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PostgresDatabase) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PostgresDatabase) SetName(v string)`

SetName sets Name field to given value.


### GetExtensions

`func (o *PostgresDatabase) GetExtensions() []string`

GetExtensions returns the Extensions field if non-nil, zero value otherwise.

### GetExtensionsOk

`func (o *PostgresDatabase) GetExtensionsOk() (*[]string, bool)`

GetExtensionsOk returns a tuple with the Extensions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExtensions

`func (o *PostgresDatabase) SetExtensions(v []string)`

SetExtensions sets Extensions field to given value.

### HasExtensions

`func (o *PostgresDatabase) HasExtensions() bool`

HasExtensions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


