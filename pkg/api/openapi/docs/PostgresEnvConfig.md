# PostgresEnvConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Database** | Pointer to **string** | Target database name within the addon. | [optional] 
**CredentialScope** | Pointer to **string** | Which credential set to inject. Mutually exclusive with superuser. | [optional] 
**Superuser** | Pointer to **bool** | Use superuser credentials. Mutually exclusive with credential_scope. | [optional] 

## Methods

### NewPostgresEnvConfig

`func NewPostgresEnvConfig() *PostgresEnvConfig`

NewPostgresEnvConfig instantiates a new PostgresEnvConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresEnvConfigWithDefaults

`func NewPostgresEnvConfigWithDefaults() *PostgresEnvConfig`

NewPostgresEnvConfigWithDefaults instantiates a new PostgresEnvConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDatabase

`func (o *PostgresEnvConfig) GetDatabase() string`

GetDatabase returns the Database field if non-nil, zero value otherwise.

### GetDatabaseOk

`func (o *PostgresEnvConfig) GetDatabaseOk() (*string, bool)`

GetDatabaseOk returns a tuple with the Database field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatabase

`func (o *PostgresEnvConfig) SetDatabase(v string)`

SetDatabase sets Database field to given value.

### HasDatabase

`func (o *PostgresEnvConfig) HasDatabase() bool`

HasDatabase returns a boolean if a field has been set.

### GetCredentialScope

`func (o *PostgresEnvConfig) GetCredentialScope() string`

GetCredentialScope returns the CredentialScope field if non-nil, zero value otherwise.

### GetCredentialScopeOk

`func (o *PostgresEnvConfig) GetCredentialScopeOk() (*string, bool)`

GetCredentialScopeOk returns a tuple with the CredentialScope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialScope

`func (o *PostgresEnvConfig) SetCredentialScope(v string)`

SetCredentialScope sets CredentialScope field to given value.

### HasCredentialScope

`func (o *PostgresEnvConfig) HasCredentialScope() bool`

HasCredentialScope returns a boolean if a field has been set.

### GetSuperuser

`func (o *PostgresEnvConfig) GetSuperuser() bool`

GetSuperuser returns the Superuser field if non-nil, zero value otherwise.

### GetSuperuserOk

`func (o *PostgresEnvConfig) GetSuperuserOk() (*bool, bool)`

GetSuperuserOk returns a tuple with the Superuser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuperuser

`func (o *PostgresEnvConfig) SetSuperuser(v bool)`

SetSuperuser sets Superuser field to given value.

### HasSuperuser

`func (o *PostgresEnvConfig) HasSuperuser() bool`

HasSuperuser returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


