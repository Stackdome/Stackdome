# PostgresConnectionInfoCredentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SuperuserSecretId** | Pointer to **string** | Secret containing superuser credentials | [optional] 
**AppUserSecrets** | Pointer to **map[string]string** | Map of database to app user secret IDs | [optional] 
**CaCertificateSecretId** | Pointer to **string** | Secret containing CA certificate | [optional] 

## Methods

### NewPostgresConnectionInfoCredentials

`func NewPostgresConnectionInfoCredentials() *PostgresConnectionInfoCredentials`

NewPostgresConnectionInfoCredentials instantiates a new PostgresConnectionInfoCredentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresConnectionInfoCredentialsWithDefaults

`func NewPostgresConnectionInfoCredentialsWithDefaults() *PostgresConnectionInfoCredentials`

NewPostgresConnectionInfoCredentialsWithDefaults instantiates a new PostgresConnectionInfoCredentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSuperuserSecretId

`func (o *PostgresConnectionInfoCredentials) GetSuperuserSecretId() string`

GetSuperuserSecretId returns the SuperuserSecretId field if non-nil, zero value otherwise.

### GetSuperuserSecretIdOk

`func (o *PostgresConnectionInfoCredentials) GetSuperuserSecretIdOk() (*string, bool)`

GetSuperuserSecretIdOk returns a tuple with the SuperuserSecretId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuperuserSecretId

`func (o *PostgresConnectionInfoCredentials) SetSuperuserSecretId(v string)`

SetSuperuserSecretId sets SuperuserSecretId field to given value.

### HasSuperuserSecretId

`func (o *PostgresConnectionInfoCredentials) HasSuperuserSecretId() bool`

HasSuperuserSecretId returns a boolean if a field has been set.

### GetAppUserSecrets

`func (o *PostgresConnectionInfoCredentials) GetAppUserSecrets() map[string]string`

GetAppUserSecrets returns the AppUserSecrets field if non-nil, zero value otherwise.

### GetAppUserSecretsOk

`func (o *PostgresConnectionInfoCredentials) GetAppUserSecretsOk() (*map[string]string, bool)`

GetAppUserSecretsOk returns a tuple with the AppUserSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAppUserSecrets

`func (o *PostgresConnectionInfoCredentials) SetAppUserSecrets(v map[string]string)`

SetAppUserSecrets sets AppUserSecrets field to given value.

### HasAppUserSecrets

`func (o *PostgresConnectionInfoCredentials) HasAppUserSecrets() bool`

HasAppUserSecrets returns a boolean if a field has been set.

### GetCaCertificateSecretId

`func (o *PostgresConnectionInfoCredentials) GetCaCertificateSecretId() string`

GetCaCertificateSecretId returns the CaCertificateSecretId field if non-nil, zero value otherwise.

### GetCaCertificateSecretIdOk

`func (o *PostgresConnectionInfoCredentials) GetCaCertificateSecretIdOk() (*string, bool)`

GetCaCertificateSecretIdOk returns a tuple with the CaCertificateSecretId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCaCertificateSecretId

`func (o *PostgresConnectionInfoCredentials) SetCaCertificateSecretId(v string)`

SetCaCertificateSecretId sets CaCertificateSecretId field to given value.

### HasCaCertificateSecretId

`func (o *PostgresConnectionInfoCredentials) HasCaCertificateSecretId() bool`

HasCaCertificateSecretId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


