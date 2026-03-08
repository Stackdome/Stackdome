# S3Credentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessKeyId** | **string** | Reference to secret containing AWS access key ID | 
**SecretAccessKey** | **string** | Reference to secret containing AWS secret access key | 
**Region** | **string** | AWS region | 
**EndpointUrl** | Pointer to **string** | Custom S3 endpoint URL for S3-compatible storage | [optional] 

## Methods

### NewS3Credentials

`func NewS3Credentials(accessKeyId string, secretAccessKey string, region string, ) *S3Credentials`

NewS3Credentials instantiates a new S3Credentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewS3CredentialsWithDefaults

`func NewS3CredentialsWithDefaults() *S3Credentials`

NewS3CredentialsWithDefaults instantiates a new S3Credentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessKeyId

`func (o *S3Credentials) GetAccessKeyId() string`

GetAccessKeyId returns the AccessKeyId field if non-nil, zero value otherwise.

### GetAccessKeyIdOk

`func (o *S3Credentials) GetAccessKeyIdOk() (*string, bool)`

GetAccessKeyIdOk returns a tuple with the AccessKeyId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessKeyId

`func (o *S3Credentials) SetAccessKeyId(v string)`

SetAccessKeyId sets AccessKeyId field to given value.


### GetSecretAccessKey

`func (o *S3Credentials) GetSecretAccessKey() string`

GetSecretAccessKey returns the SecretAccessKey field if non-nil, zero value otherwise.

### GetSecretAccessKeyOk

`func (o *S3Credentials) GetSecretAccessKeyOk() (*string, bool)`

GetSecretAccessKeyOk returns a tuple with the SecretAccessKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretAccessKey

`func (o *S3Credentials) SetSecretAccessKey(v string)`

SetSecretAccessKey sets SecretAccessKey field to given value.


### GetRegion

`func (o *S3Credentials) GetRegion() string`

GetRegion returns the Region field if non-nil, zero value otherwise.

### GetRegionOk

`func (o *S3Credentials) GetRegionOk() (*string, bool)`

GetRegionOk returns a tuple with the Region field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegion

`func (o *S3Credentials) SetRegion(v string)`

SetRegion sets Region field to given value.


### GetEndpointUrl

`func (o *S3Credentials) GetEndpointUrl() string`

GetEndpointUrl returns the EndpointUrl field if non-nil, zero value otherwise.

### GetEndpointUrlOk

`func (o *S3Credentials) GetEndpointUrlOk() (*string, bool)`

GetEndpointUrlOk returns a tuple with the EndpointUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEndpointUrl

`func (o *S3Credentials) SetEndpointUrl(v string)`

SetEndpointUrl sets EndpointUrl field to given value.

### HasEndpointUrl

`func (o *S3Credentials) HasEndpointUrl() bool`

HasEndpointUrl returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


