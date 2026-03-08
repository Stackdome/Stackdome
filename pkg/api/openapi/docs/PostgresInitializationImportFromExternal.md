# PostgresInitializationImportFromExternal

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Host** | **string** |  | 
**Port** | **int32** |  | [default to 5432]
**Database** | **string** |  | 
**Username** | **string** |  | 
**PasswordSecretId** | **string** | Secret containing the password | 
**SslMode** | Pointer to **string** |  | [optional] [default to "require"]
**DatabasesToImport** | Pointer to **[]string** | Specific databases to import | [optional] 

## Methods

### NewPostgresInitializationImportFromExternal

`func NewPostgresInitializationImportFromExternal(host string, port int32, database string, username string, passwordSecretId string, ) *PostgresInitializationImportFromExternal`

NewPostgresInitializationImportFromExternal instantiates a new PostgresInitializationImportFromExternal object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresInitializationImportFromExternalWithDefaults

`func NewPostgresInitializationImportFromExternalWithDefaults() *PostgresInitializationImportFromExternal`

NewPostgresInitializationImportFromExternalWithDefaults instantiates a new PostgresInitializationImportFromExternal object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHost

`func (o *PostgresInitializationImportFromExternal) GetHost() string`

GetHost returns the Host field if non-nil, zero value otherwise.

### GetHostOk

`func (o *PostgresInitializationImportFromExternal) GetHostOk() (*string, bool)`

GetHostOk returns a tuple with the Host field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHost

`func (o *PostgresInitializationImportFromExternal) SetHost(v string)`

SetHost sets Host field to given value.


### GetPort

`func (o *PostgresInitializationImportFromExternal) GetPort() int32`

GetPort returns the Port field if non-nil, zero value otherwise.

### GetPortOk

`func (o *PostgresInitializationImportFromExternal) GetPortOk() (*int32, bool)`

GetPortOk returns a tuple with the Port field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPort

`func (o *PostgresInitializationImportFromExternal) SetPort(v int32)`

SetPort sets Port field to given value.


### GetDatabase

`func (o *PostgresInitializationImportFromExternal) GetDatabase() string`

GetDatabase returns the Database field if non-nil, zero value otherwise.

### GetDatabaseOk

`func (o *PostgresInitializationImportFromExternal) GetDatabaseOk() (*string, bool)`

GetDatabaseOk returns a tuple with the Database field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatabase

`func (o *PostgresInitializationImportFromExternal) SetDatabase(v string)`

SetDatabase sets Database field to given value.


### GetUsername

`func (o *PostgresInitializationImportFromExternal) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *PostgresInitializationImportFromExternal) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *PostgresInitializationImportFromExternal) SetUsername(v string)`

SetUsername sets Username field to given value.


### GetPasswordSecretId

`func (o *PostgresInitializationImportFromExternal) GetPasswordSecretId() string`

GetPasswordSecretId returns the PasswordSecretId field if non-nil, zero value otherwise.

### GetPasswordSecretIdOk

`func (o *PostgresInitializationImportFromExternal) GetPasswordSecretIdOk() (*string, bool)`

GetPasswordSecretIdOk returns a tuple with the PasswordSecretId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPasswordSecretId

`func (o *PostgresInitializationImportFromExternal) SetPasswordSecretId(v string)`

SetPasswordSecretId sets PasswordSecretId field to given value.


### GetSslMode

`func (o *PostgresInitializationImportFromExternal) GetSslMode() string`

GetSslMode returns the SslMode field if non-nil, zero value otherwise.

### GetSslModeOk

`func (o *PostgresInitializationImportFromExternal) GetSslModeOk() (*string, bool)`

GetSslModeOk returns a tuple with the SslMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSslMode

`func (o *PostgresInitializationImportFromExternal) SetSslMode(v string)`

SetSslMode sets SslMode field to given value.

### HasSslMode

`func (o *PostgresInitializationImportFromExternal) HasSslMode() bool`

HasSslMode returns a boolean if a field has been set.

### GetDatabasesToImport

`func (o *PostgresInitializationImportFromExternal) GetDatabasesToImport() []string`

GetDatabasesToImport returns the DatabasesToImport field if non-nil, zero value otherwise.

### GetDatabasesToImportOk

`func (o *PostgresInitializationImportFromExternal) GetDatabasesToImportOk() (*[]string, bool)`

GetDatabasesToImportOk returns a tuple with the DatabasesToImport field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatabasesToImport

`func (o *PostgresInitializationImportFromExternal) SetDatabasesToImport(v []string)`

SetDatabasesToImport sets DatabasesToImport field to given value.

### HasDatabasesToImport

`func (o *PostgresInitializationImportFromExternal) HasDatabasesToImport() bool`

HasDatabasesToImport returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


