# PostgresConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EnableSuperuserAccess** | Pointer to **bool** | Enable superuser access to the database | [optional] [default to false]
**Parameters** | Pointer to **map[string]string** | PostgreSQL configuration parameters | [optional] 

## Methods

### NewPostgresConfiguration

`func NewPostgresConfiguration() *PostgresConfiguration`

NewPostgresConfiguration instantiates a new PostgresConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresConfigurationWithDefaults

`func NewPostgresConfigurationWithDefaults() *PostgresConfiguration`

NewPostgresConfigurationWithDefaults instantiates a new PostgresConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnableSuperuserAccess

`func (o *PostgresConfiguration) GetEnableSuperuserAccess() bool`

GetEnableSuperuserAccess returns the EnableSuperuserAccess field if non-nil, zero value otherwise.

### GetEnableSuperuserAccessOk

`func (o *PostgresConfiguration) GetEnableSuperuserAccessOk() (*bool, bool)`

GetEnableSuperuserAccessOk returns a tuple with the EnableSuperuserAccess field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnableSuperuserAccess

`func (o *PostgresConfiguration) SetEnableSuperuserAccess(v bool)`

SetEnableSuperuserAccess sets EnableSuperuserAccess field to given value.

### HasEnableSuperuserAccess

`func (o *PostgresConfiguration) HasEnableSuperuserAccess() bool`

HasEnableSuperuserAccess returns a boolean if a field has been set.

### GetParameters

`func (o *PostgresConfiguration) GetParameters() map[string]string`

GetParameters returns the Parameters field if non-nil, zero value otherwise.

### GetParametersOk

`func (o *PostgresConfiguration) GetParametersOk() (*map[string]string, bool)`

GetParametersOk returns a tuple with the Parameters field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParameters

`func (o *PostgresConfiguration) SetParameters(v map[string]string)`

SetParameters sets Parameters field to given value.

### HasParameters

`func (o *PostgresConfiguration) HasParameters() bool`

HasParameters returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


