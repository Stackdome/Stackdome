# ImageSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ref** | **string** |  | 
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


