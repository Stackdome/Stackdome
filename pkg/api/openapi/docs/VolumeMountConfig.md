# VolumeMountConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**MountPath** | **string** | Absolute path where the volume is mounted in the container. | 
**SubPath** | Pointer to **string** | Sub-path within the volume to mount. | [optional] 
**ReadOnly** | Pointer to **bool** | Mount the volume read-only. | [optional] 

## Methods

### NewVolumeMountConfig

`func NewVolumeMountConfig(mountPath string, ) *VolumeMountConfig`

NewVolumeMountConfig instantiates a new VolumeMountConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeMountConfigWithDefaults

`func NewVolumeMountConfigWithDefaults() *VolumeMountConfig`

NewVolumeMountConfigWithDefaults instantiates a new VolumeMountConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMountPath

`func (o *VolumeMountConfig) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *VolumeMountConfig) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *VolumeMountConfig) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.


### GetSubPath

`func (o *VolumeMountConfig) GetSubPath() string`

GetSubPath returns the SubPath field if non-nil, zero value otherwise.

### GetSubPathOk

`func (o *VolumeMountConfig) GetSubPathOk() (*string, bool)`

GetSubPathOk returns a tuple with the SubPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubPath

`func (o *VolumeMountConfig) SetSubPath(v string)`

SetSubPath sets SubPath field to given value.

### HasSubPath

`func (o *VolumeMountConfig) HasSubPath() bool`

HasSubPath returns a boolean if a field has been set.

### GetReadOnly

`func (o *VolumeMountConfig) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *VolumeMountConfig) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *VolumeMountConfig) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *VolumeMountConfig) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


