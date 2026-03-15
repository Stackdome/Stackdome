# SecretReference

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SecretId** | **string** | UUID of the Stackdome secret | 
**Key** | **string** | Key within the secret containing the value | 

## Methods

### NewSecretReference

`func NewSecretReference(secretId string, key string, ) *SecretReference`

NewSecretReference instantiates a new SecretReference object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretReferenceWithDefaults

`func NewSecretReferenceWithDefaults() *SecretReference`

NewSecretReferenceWithDefaults instantiates a new SecretReference object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSecretId

`func (o *SecretReference) GetSecretId() string`

GetSecretId returns the SecretId field if non-nil, zero value otherwise.

### GetSecretIdOk

`func (o *SecretReference) GetSecretIdOk() (*string, bool)`

GetSecretIdOk returns a tuple with the SecretId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretId

`func (o *SecretReference) SetSecretId(v string)`

SetSecretId sets SecretId field to given value.


### GetKey

`func (o *SecretReference) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *SecretReference) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *SecretReference) SetKey(v string)`

SetKey sets Key field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


