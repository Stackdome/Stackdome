# EnvVarFromSecret

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**SecretRef** | [**SecretRef**](SecretRef.md) |  | 
**Key** | **string** | The key in the secret to use for this environment variable | 

## Methods

### NewEnvVarFromSecret

`func NewEnvVarFromSecret(name string, secretRef SecretRef, key string, ) *EnvVarFromSecret`

NewEnvVarFromSecret instantiates a new EnvVarFromSecret object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEnvVarFromSecretWithDefaults

`func NewEnvVarFromSecretWithDefaults() *EnvVarFromSecret`

NewEnvVarFromSecretWithDefaults instantiates a new EnvVarFromSecret object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *EnvVarFromSecret) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *EnvVarFromSecret) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *EnvVarFromSecret) SetName(v string)`

SetName sets Name field to given value.


### GetSecretRef

`func (o *EnvVarFromSecret) GetSecretRef() SecretRef`

GetSecretRef returns the SecretRef field if non-nil, zero value otherwise.

### GetSecretRefOk

`func (o *EnvVarFromSecret) GetSecretRefOk() (*SecretRef, bool)`

GetSecretRefOk returns a tuple with the SecretRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretRef

`func (o *EnvVarFromSecret) SetSecretRef(v SecretRef)`

SetSecretRef sets SecretRef field to given value.


### GetKey

`func (o *EnvVarFromSecret) GetKey() string`

GetKey returns the Key field if non-nil, zero value otherwise.

### GetKeyOk

`func (o *EnvVarFromSecret) GetKeyOk() (*string, bool)`

GetKeyOk returns a tuple with the Key field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKey

`func (o *EnvVarFromSecret) SetKey(v string)`

SetKey sets Key field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


