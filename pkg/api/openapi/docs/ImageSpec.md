# ImageSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Image** | **string** |  | 
**PullSecret** | Pointer to [**SecretRef**](SecretRef.md) |  | [optional] 

## Methods

### NewImageSpec

`func NewImageSpec(image string, ) *ImageSpec`

NewImageSpec instantiates a new ImageSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewImageSpecWithDefaults

`func NewImageSpecWithDefaults() *ImageSpec`

NewImageSpecWithDefaults instantiates a new ImageSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetImage

`func (o *ImageSpec) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *ImageSpec) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *ImageSpec) SetImage(v string)`

SetImage sets Image field to given value.


### GetPullSecret

`func (o *ImageSpec) GetPullSecret() SecretRef`

GetPullSecret returns the PullSecret field if non-nil, zero value otherwise.

### GetPullSecretOk

`func (o *ImageSpec) GetPullSecretOk() (*SecretRef, bool)`

GetPullSecretOk returns a tuple with the PullSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPullSecret

`func (o *ImageSpec) SetPullSecret(v SecretRef)`

SetPullSecret sets PullSecret field to given value.

### HasPullSecret

`func (o *ImageSpec) HasPullSecret() bool`

HasPullSecret returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


