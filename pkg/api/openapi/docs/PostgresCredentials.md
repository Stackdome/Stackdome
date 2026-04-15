# PostgresCredentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Database** | Pointer to **string** |  | [optional] 
**Host** | Pointer to **string** |  | [optional] 
**Port** | Pointer to **int32** |  | [optional] 
**Username** | Pointer to **string** |  | [optional] 
**Password** | Pointer to **string** |  | [optional] 
**SslMode** | Pointer to **string** |  | [optional] 
**ConnectionString** | Pointer to **string** |  | [optional] 
**CaCertificate** | Pointer to **string** |  | [optional] 

## Methods

### NewPostgresCredentials

`func NewPostgresCredentials() *PostgresCredentials`

NewPostgresCredentials instantiates a new PostgresCredentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresCredentialsWithDefaults

`func NewPostgresCredentialsWithDefaults() *PostgresCredentials`

NewPostgresCredentialsWithDefaults instantiates a new PostgresCredentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDatabase

`func (o *PostgresCredentials) GetDatabase() string`

GetDatabase returns the Database field if non-nil, zero value otherwise.

### GetDatabaseOk

`func (o *PostgresCredentials) GetDatabaseOk() (*string, bool)`

GetDatabaseOk returns a tuple with the Database field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatabase

`func (o *PostgresCredentials) SetDatabase(v string)`

SetDatabase sets Database field to given value.

### HasDatabase

`func (o *PostgresCredentials) HasDatabase() bool`

HasDatabase returns a boolean if a field has been set.

### GetHost

`func (o *PostgresCredentials) GetHost() string`

GetHost returns the Host field if non-nil, zero value otherwise.

### GetHostOk

`func (o *PostgresCredentials) GetHostOk() (*string, bool)`

GetHostOk returns a tuple with the Host field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHost

`func (o *PostgresCredentials) SetHost(v string)`

SetHost sets Host field to given value.

### HasHost

`func (o *PostgresCredentials) HasHost() bool`

HasHost returns a boolean if a field has been set.

### GetPort

`func (o *PostgresCredentials) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *PostgresCredentials) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *PostgresCredentials) SetPort(v int32)`

SetPort sets Port field to given value.

### HasPort

`func (o *PostgresCredentials) HasPort() bool`

HasPort returns a boolean if a field has been set.

### GetUsername

`func (o *PostgresCredentials) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *PostgresCredentials) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *PostgresCredentials) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *PostgresCredentials) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetPassword

`func (o *PostgresCredentials) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *PostgresCredentials) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *PostgresCredentials) SetPassword(v string)`

SetPassword sets Password field to given value.

### HasPassword

`func (o *PostgresCredentials) HasPassword() bool`

HasPassword returns a boolean if a field has been set.

### GetSslMode

`func (o *PostgresCredentials) GetSslMode() string`

GetSslMode returns the SslMode field if non-nil, zero value otherwise.

### GetSslModeOk

`func (o *PostgresCredentials) GetSslModeOk() (*string, bool)`

GetSslModeOk returns a tuple with the SslMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSslMode

`func (o *PostgresCredentials) SetSslMode(v string)`

SetSslMode sets SslMode field to given value.

### HasSslMode

`func (o *PostgresCredentials) HasSslMode() bool`

HasSslMode returns a boolean if a field has been set.

### GetConnectionString

`func (o *PostgresCredentials) GetConnectionString() string`

GetConnectionString returns the ConnectionString field if non-nil, zero value otherwise.

### GetConnectionStringOk

`func (o *PostgresCredentials) GetConnectionStringOk() (*string, bool)`

GetConnectionStringOk returns a tuple with the ConnectionString field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionString

`func (o *PostgresCredentials) SetConnectionString(v string)`

SetConnectionString sets ConnectionString field to given value.

### HasConnectionString

`func (o *PostgresCredentials) HasConnectionString() bool`

HasConnectionString returns a boolean if a field has been set.

### GetCaCertificate

`func (o *PostgresCredentials) GetCaCertificate() string`

GetCaCertificate returns the CaCertificate field if non-nil, zero value otherwise.

### GetCaCertificateOk

`func (o *PostgresCredentials) GetCaCertificateOk() (*string, bool)`

GetCaCertificateOk returns a tuple with the CaCertificate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCaCertificate

`func (o *PostgresCredentials) SetCaCertificate(v string)`

SetCaCertificate sets CaCertificate field to given value.

### HasCaCertificate

`func (o *PostgresCredentials) HasCaCertificate() bool`

HasCaCertificate returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


