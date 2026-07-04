# ImageSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ref** | **string** |  | 
**Credentials** | Pointer to [**InlineCredentials**](InlineCredentials.md) |  | [optional] 
**CredentialsConfigured** | Pointer to **bool** |  | [optional] [readonly] 
**RegistryCredentialsId** | Pointer to **string** | Org-level registry credential override for pull auth | [optional] 

## Methods

### NewImageSource

`func NewImageSource(ref string, ) *ImageSource`

NewImageSource instantiates a new ImageSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImageSourceWithDefaults

`func NewImageSourceWithDefaults() *ImageSource`

NewImageSourceWithDefaults instantiates a new ImageSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRef

`func (o *ImageSource) GetRef() string`

GetRef returns the Ref field if non-nil, zero value otherwise.

### GetRefOk

`func (o *ImageSource) GetRefOk() (*string, bool)`

GetRefOk returns a tuple with the Ref field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRef

`func (o *ImageSource) SetRef(v string)`

SetRef sets Ref field to given value.


### GetCredentials

`func (o *ImageSource) GetCredentials() InlineCredentials`

GetCredentials returns the Credentials field if non-nil, zero value otherwise.

### GetCredentialsOk

`func (o *ImageSource) GetCredentialsOk() (*InlineCredentials, bool)`

GetCredentialsOk returns a tuple with the Credentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentials

`func (o *ImageSource) SetCredentials(v InlineCredentials)`

SetCredentials sets Credentials field to given value.

### HasCredentials

`func (o *ImageSource) HasCredentials() bool`

HasCredentials returns a boolean if a field has been set.

### GetCredentialsConfigured

`func (o *ImageSource) GetCredentialsConfigured() bool`

GetCredentialsConfigured returns the CredentialsConfigured field if non-nil, zero value otherwise.

### GetCredentialsConfiguredOk

`func (o *ImageSource) GetCredentialsConfiguredOk() (*bool, bool)`

GetCredentialsConfiguredOk returns a tuple with the CredentialsConfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialsConfigured

`func (o *ImageSource) SetCredentialsConfigured(v bool)`

SetCredentialsConfigured sets CredentialsConfigured field to given value.

### HasCredentialsConfigured

`func (o *ImageSource) HasCredentialsConfigured() bool`

HasCredentialsConfigured returns a boolean if a field has been set.

### GetRegistryCredentialsId

`func (o *ImageSource) GetRegistryCredentialsId() string`

GetRegistryCredentialsId returns the RegistryCredentialsId field if non-nil, zero value otherwise.

### GetRegistryCredentialsIdOk

`func (o *ImageSource) GetRegistryCredentialsIdOk() (*string, bool)`

GetRegistryCredentialsIdOk returns a tuple with the RegistryCredentialsId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRegistryCredentialsId

`func (o *ImageSource) SetRegistryCredentialsId(v string)`

SetRegistryCredentialsId sets RegistryCredentialsId field to given value.

### HasRegistryCredentialsId

`func (o *ImageSource) HasRegistryCredentialsId() bool`

HasRegistryCredentialsId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


