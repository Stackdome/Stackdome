# PostgresConnectionInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Host** | Pointer to **string** | Primary host endpoint | [optional] 
**Port** | Pointer to **int32** |  | [optional] [default to 5432]
**Databases** | Pointer to [**[]PostgresConnectionInfoDatabasesInner**](PostgresConnectionInfoDatabasesInner.md) |  | [optional] 
**Credentials** | Pointer to [**PostgresConnectionInfoCredentials**](PostgresConnectionInfoCredentials.md) |  | [optional] 

## Methods

### NewPostgresConnectionInfo

`func NewPostgresConnectionInfo() *PostgresConnectionInfo`

NewPostgresConnectionInfo instantiates a new PostgresConnectionInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresConnectionInfoWithDefaults

`func NewPostgresConnectionInfoWithDefaults() *PostgresConnectionInfo`

NewPostgresConnectionInfoWithDefaults instantiates a new PostgresConnectionInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHost

`func (o *PostgresConnectionInfo) GetHost() string`

GetHost returns the Host field if non-nil, zero value otherwise.

### GetHostOk

`func (o *PostgresConnectionInfo) GetHostOk() (*string, bool)`

GetHostOk returns a tuple with the Host field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHost

`func (o *PostgresConnectionInfo) SetHost(v string)`

SetHost sets Host field to given value.

### HasHost

`func (o *PostgresConnectionInfo) HasHost() bool`

HasHost returns a boolean if a field has been set.

### GetPort

`func (o *PostgresConnectionInfo) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *PostgresConnectionInfo) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *PostgresConnectionInfo) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *PostgresConnectionInfo) HasPort() bool`

HasPort returns a boolean if a field has been set.

### GetDatabases

`func (o *PostgresConnectionInfo) GetDatabases() []PostgresConnectionInfoDatabasesInner`

GetDatabases returns the Databases field if non-nil, zero value otherwise.

### GetDatabasesOk

`func (o *PostgresConnectionInfo) GetDatabasesOk() (*[]PostgresConnectionInfoDatabasesInner, bool)`

GetDatabasesOk returns a tuple with the Databases field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatabases

`func (o *PostgresConnectionInfo) SetDatabases(v []PostgresConnectionInfoDatabasesInner)`

SetDatabases sets Databases field to given value.

### HasDatabases

`func (o *PostgresConnectionInfo) HasDatabases() bool`

HasDatabases returns a boolean if a field has been set.

### GetCredentials

`func (o *PostgresConnectionInfo) GetCredentials() PostgresConnectionInfoCredentials`

GetCredentials returns the Credentials field if non-nil, zero value otherwise.

### GetCredentialsOk

`func (o *PostgresConnectionInfo) GetCredentialsOk() (*PostgresConnectionInfoCredentials, bool)`

GetCredentialsOk returns a tuple with the Credentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentials

`func (o *PostgresConnectionInfo) SetCredentials(v PostgresConnectionInfoCredentials)`

SetCredentials sets Credentials field to given value.

### HasCredentials

`func (o *PostgresConnectionInfo) HasCredentials() bool`

HasCredentials returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


