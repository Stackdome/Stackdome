# ObjectStoreConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**S3Credentials** | Pointer to [**S3Credentials**](S3Credentials.md) |  | [optional] 
**AzureCredentials** | Pointer to [**AzureCredentials**](AzureCredentials.md) |  | [optional] 
**GcsCredentials** | Pointer to [**GCSCredentials**](GCSCredentials.md) |  | [optional] 

## Methods

### NewObjectStoreConfiguration

`func NewObjectStoreConfiguration() *ObjectStoreConfiguration`

NewObjectStoreConfiguration instantiates a new ObjectStoreConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewObjectStoreConfigurationWithDefaults

`func NewObjectStoreConfigurationWithDefaults() *ObjectStoreConfiguration`

NewObjectStoreConfigurationWithDefaults instantiates a new ObjectStoreConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetS3Credentials

`func (o *ObjectStoreConfiguration) GetS3Credentials() S3Credentials`

GetS3Credentials returns the S3Credentials field if non-nil, zero value otherwise.

### GetS3CredentialsOk

`func (o *ObjectStoreConfiguration) GetS3CredentialsOk() (*S3Credentials, bool)`

GetS3CredentialsOk returns a tuple with the S3Credentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3Credentials

`func (o *ObjectStoreConfiguration) SetS3Credentials(v S3Credentials)`

SetS3Credentials sets S3Credentials field to given value.

### HasS3Credentials

`func (o *ObjectStoreConfiguration) HasS3Credentials() bool`

HasS3Credentials returns a boolean if a field has been set.

### GetAzureCredentials

`func (o *ObjectStoreConfiguration) GetAzureCredentials() AzureCredentials`

GetAzureCredentials returns the AzureCredentials field if non-nil, zero value otherwise.

### GetAzureCredentialsOk

`func (o *ObjectStoreConfiguration) GetAzureCredentialsOk() (*AzureCredentials, bool)`

GetAzureCredentialsOk returns a tuple with the AzureCredentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAzureCredentials

`func (o *ObjectStoreConfiguration) SetAzureCredentials(v AzureCredentials)`

SetAzureCredentials sets AzureCredentials field to given value.

### HasAzureCredentials

`func (o *ObjectStoreConfiguration) HasAzureCredentials() bool`

HasAzureCredentials returns a boolean if a field has been set.

### GetGcsCredentials

`func (o *ObjectStoreConfiguration) GetGcsCredentials() GCSCredentials`

GetGcsCredentials returns the GcsCredentials field if non-nil, zero value otherwise.

### GetGcsCredentialsOk

`func (o *ObjectStoreConfiguration) GetGcsCredentialsOk() (*GCSCredentials, bool)`

GetGcsCredentialsOk returns a tuple with the GcsCredentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGcsCredentials

`func (o *ObjectStoreConfiguration) SetGcsCredentials(v GCSCredentials)`

SetGcsCredentials sets GcsCredentials field to given value.

### HasGcsCredentials

`func (o *ObjectStoreConfiguration) HasGcsCredentials() bool`

HasGcsCredentials returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


