# ImageRepository

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ExternalImageRepoUrl** | Pointer to **string** |  | [optional] 
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

### GetExternalImageRepoUrl

`func (o *ImageRepository) GetExternalImageRepoUrl() string`

GetExternalImageRepoUrl returns the ExternalImageRepoUrl field if non-nil, zero value otherwise.

### GetExternalImageRepoUrlOk

`func (o *ImageRepository) GetExternalImageRepoUrlOk() (*string, bool)`

GetExternalImageRepoUrlOk returns a tuple with the ExternalImageRepoUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExternalImageRepoUrl

`func (o *ImageRepository) SetExternalImageRepoUrl(v string)`

SetExternalImageRepoUrl sets ExternalImageRepoUrl field to given value.

### HasExternalImageRepoUrl

`func (o *ImageRepository) HasExternalImageRepoUrl() bool`

HasExternalImageRepoUrl returns a boolean if a field has been set.

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


