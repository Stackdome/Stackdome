# PostgresAddonEnvSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AddonId** | **string** |  | 
**Database** | **string** |  | 
**EnvMapping** | **map[string]string** | Maps addon credential fields to environment variable names. Valid fields are host, port, username, password, database, sslmode, connectionString, caCertificate. | 

## Methods

### NewPostgresAddonEnvSource

`func NewPostgresAddonEnvSource(addonId string, database string, envMapping map[string]string, ) *PostgresAddonEnvSource`

NewPostgresAddonEnvSource instantiates a new PostgresAddonEnvSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresAddonEnvSourceWithDefaults

`func NewPostgresAddonEnvSourceWithDefaults() *PostgresAddonEnvSource`

NewPostgresAddonEnvSourceWithDefaults instantiates a new PostgresAddonEnvSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAddonId

`func (o *PostgresAddonEnvSource) GetAddonId() string`

GetAddonId returns the AddonId field if non-nil, zero value otherwise.

### GetAddonIdOk

`func (o *PostgresAddonEnvSource) GetAddonIdOk() (*string, bool)`

GetAddonIdOk returns a tuple with the AddonId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAddonId

`func (o *PostgresAddonEnvSource) SetAddonId(v string)`

SetAddonId sets AddonId field to given value.


### GetDatabase

`func (o *PostgresAddonEnvSource) GetDatabase() string`

GetDatabase returns the Database field if non-nil, zero value otherwise.

### GetDatabaseOk

`func (o *PostgresAddonEnvSource) GetDatabaseOk() (*string, bool)`

GetDatabaseOk returns a tuple with the Database field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatabase

`func (o *PostgresAddonEnvSource) SetDatabase(v string)`

SetDatabase sets Database field to given value.


### GetEnvMapping

`func (o *PostgresAddonEnvSource) GetEnvMapping() map[string]string`

GetEnvMapping returns the EnvMapping field if non-nil, zero value otherwise.

### GetEnvMappingOk

`func (o *PostgresAddonEnvSource) GetEnvMappingOk() (*map[string]string, bool)`

GetEnvMappingOk returns a tuple with the EnvMapping field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvMapping

`func (o *PostgresAddonEnvSource) SetEnvMapping(v map[string]string)`

SetEnvMapping sets EnvMapping field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


