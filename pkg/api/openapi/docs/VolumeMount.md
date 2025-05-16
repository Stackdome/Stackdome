# VolumeMount

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StackResourceId** | Pointer to **string** |  | [optional] [readonly] 
**SourceVolumeType** | Pointer to [**VolumeMountSourceType**](VolumeMountSourceType.md) |  | [optional] 
**SourceVolumeName** | **string** |  | 
**SourceSubPath** | Pointer to **string** |  | [optional] 
**TargetPath** | **string** |  | 

## Methods

### NewVolumeMount

`func NewVolumeMount(sourceVolumeName string, targetPath string, ) *VolumeMount`

NewVolumeMount instantiates a new VolumeMount object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeMountWithDefaults

`func NewVolumeMountWithDefaults() *VolumeMount`

NewVolumeMountWithDefaults instantiates a new VolumeMount object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStackResourceId

`func (o *VolumeMount) GetStackResourceId() string`

GetStackResourceId returns the StackResourceId field if non-nil, zero value otherwise.

### GetStackResourceIdOk

`func (o *VolumeMount) GetStackResourceIdOk() (*string, bool)`

GetStackResourceIdOk returns a tuple with the StackResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackResourceId

`func (o *VolumeMount) SetStackResourceId(v string)`

SetStackResourceId sets StackResourceId field to given value.

### HasStackResourceId

`func (o *VolumeMount) HasStackResourceId() bool`

HasStackResourceId returns a boolean if a field has been set.

### GetSourceVolumeType

`func (o *VolumeMount) GetSourceVolumeType() VolumeMountSourceType`

GetSourceVolumeType returns the SourceVolumeType field if non-nil, zero value otherwise.

### GetSourceVolumeTypeOk

`func (o *VolumeMount) GetSourceVolumeTypeOk() (*VolumeMountSourceType, bool)`

GetSourceVolumeTypeOk returns a tuple with the SourceVolumeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceVolumeType

`func (o *VolumeMount) SetSourceVolumeType(v VolumeMountSourceType)`

SetSourceVolumeType sets SourceVolumeType field to given value.

### HasSourceVolumeType

`func (o *VolumeMount) HasSourceVolumeType() bool`

HasSourceVolumeType returns a boolean if a field has been set.

### GetSourceVolumeName

`func (o *VolumeMount) GetSourceVolumeName() string`

GetSourceVolumeName returns the SourceVolumeName field if non-nil, zero value otherwise.

### GetSourceVolumeNameOk

`func (o *VolumeMount) GetSourceVolumeNameOk() (*string, bool)`

GetSourceVolumeNameOk returns a tuple with the SourceVolumeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceVolumeName

`func (o *VolumeMount) SetSourceVolumeName(v string)`

SetSourceVolumeName sets SourceVolumeName field to given value.


### GetSourceSubPath

`func (o *VolumeMount) GetSourceSubPath() string`

GetSourceSubPath returns the SourceSubPath field if non-nil, zero value otherwise.

### GetSourceSubPathOk

`func (o *VolumeMount) GetSourceSubPathOk() (*string, bool)`

GetSourceSubPathOk returns a tuple with the SourceSubPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceSubPath

`func (o *VolumeMount) SetSourceSubPath(v string)`

SetSourceSubPath sets SourceSubPath field to given value.

### HasSourceSubPath

`func (o *VolumeMount) HasSourceSubPath() bool`

HasSourceSubPath returns a boolean if a field has been set.

### GetTargetPath

`func (o *VolumeMount) GetTargetPath() string`

GetTargetPath returns the TargetPath field if non-nil, zero value otherwise.

### GetTargetPathOk

`func (o *VolumeMount) GetTargetPathOk() (*string, bool)`

GetTargetPathOk returns a tuple with the TargetPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetPath

`func (o *VolumeMount) SetTargetPath(v string)`

SetTargetPath sets TargetPath field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


