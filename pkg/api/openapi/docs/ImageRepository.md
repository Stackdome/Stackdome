# ImageRepository

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExternalImageRef** | Pointer to **string** |  | [optional] 
**UseInternalRegistry** | Pointer to **bool** |  | [optional] 

## Methods

### NewImageRepository

`func NewImageRepository() *ImageRepository`

NewImageRepository instantiates a new ImageRepository object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImageRepositoryWithDefaults

`func NewImageRepositoryWithDefaults() *ImageRepository`

NewImageRepositoryWithDefaults instantiates a new ImageRepository object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetExternalImageRef

`func (o *ImageRepository) GetExternalImageRef() string`

GetExternalImageRef returns the ExternalImageRef field if non-nil, zero value otherwise.

### GetExternalImageRefOk

`func (o *ImageRepository) GetExternalImageRefOk() (*string, bool)`

GetExternalImageRefOk returns a tuple with the ExternalImageRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalImageRef

`func (o *ImageRepository) SetExternalImageRef(v string)`

SetExternalImageRef sets ExternalImageRef field to given value.

### HasExternalImageRef

`func (o *ImageRepository) HasExternalImageRef() bool`

HasExternalImageRef returns a boolean if a field has been set.

### GetUseInternalRegistry

`func (o *ImageRepository) GetUseInternalRegistry() bool`

GetUseInternalRegistry returns the UseInternalRegistry field if non-nil, zero value otherwise.

### GetUseInternalRegistryOk

`func (o *ImageRepository) GetUseInternalRegistryOk() (*bool, bool)`

GetUseInternalRegistryOk returns a tuple with the UseInternalRegistry field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseInternalRegistry

`func (o *ImageRepository) SetUseInternalRegistry(v bool)`

SetUseInternalRegistry sets UseInternalRegistry field to given value.

### HasUseInternalRegistry

`func (o *ImageRepository) HasUseInternalRegistry() bool`

HasUseInternalRegistry returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


