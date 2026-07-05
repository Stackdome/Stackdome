# VolumeBuildSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**VolumeId** | Pointer to **string** |  | [optional] 
**VolumeName** | Pointer to **string** | Name of a volume defined on the stack; either volume_id or volume_name is required | [optional] 
**CurrentVolumeHash** | Pointer to **string** | Content hash of the volume used as the build source revision | [optional] 
**DockerfilePath** | Pointer to **string** |  | [optional] [default to "Dockerfile"]
**BuildContext** | Pointer to **string** |  | [optional] [default to "."]

## Methods

### NewVolumeBuildSource

`func NewVolumeBuildSource() *VolumeBuildSource`

NewVolumeBuildSource instantiates a new VolumeBuildSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeBuildSourceWithDefaults

`func NewVolumeBuildSourceWithDefaults() *VolumeBuildSource`

NewVolumeBuildSourceWithDefaults instantiates a new VolumeBuildSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVolumeId

`func (o *VolumeBuildSource) GetVolumeId() string`

GetVolumeId returns the VolumeId field if non-nil, zero value otherwise.

### GetVolumeIdOk

`func (o *VolumeBuildSource) GetVolumeIdOk() (*string, bool)`

GetVolumeIdOk returns a tuple with the VolumeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeId

`func (o *VolumeBuildSource) SetVolumeId(v string)`

SetVolumeId sets VolumeId field to given value.

### HasVolumeId

`func (o *VolumeBuildSource) HasVolumeId() bool`

HasVolumeId returns a boolean if a field has been set.

### GetVolumeName

`func (o *VolumeBuildSource) GetVolumeName() string`

GetVolumeName returns the VolumeName field if non-nil, zero value otherwise.

### GetVolumeNameOk

`func (o *VolumeBuildSource) GetVolumeNameOk() (*string, bool)`

GetVolumeNameOk returns a tuple with the VolumeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeName

`func (o *VolumeBuildSource) SetVolumeName(v string)`

SetVolumeName sets VolumeName field to given value.

### HasVolumeName

`func (o *VolumeBuildSource) HasVolumeName() bool`

HasVolumeName returns a boolean if a field has been set.

### GetCurrentVolumeHash

`func (o *VolumeBuildSource) GetCurrentVolumeHash() string`

GetCurrentVolumeHash returns the CurrentVolumeHash field if non-nil, zero value otherwise.

### GetCurrentVolumeHashOk

`func (o *VolumeBuildSource) GetCurrentVolumeHashOk() (*string, bool)`

GetCurrentVolumeHashOk returns a tuple with the CurrentVolumeHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentVolumeHash

`func (o *VolumeBuildSource) SetCurrentVolumeHash(v string)`

SetCurrentVolumeHash sets CurrentVolumeHash field to given value.

### HasCurrentVolumeHash

`func (o *VolumeBuildSource) HasCurrentVolumeHash() bool`

HasCurrentVolumeHash returns a boolean if a field has been set.

### GetDockerfilePath

`func (o *VolumeBuildSource) GetDockerfilePath() string`

GetDockerfilePath returns the DockerfilePath field if non-nil, zero value otherwise.

### GetDockerfilePathOk

`func (o *VolumeBuildSource) GetDockerfilePathOk() (*string, bool)`

GetDockerfilePathOk returns a tuple with the DockerfilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDockerfilePath

`func (o *VolumeBuildSource) SetDockerfilePath(v string)`

SetDockerfilePath sets DockerfilePath field to given value.

### HasDockerfilePath

`func (o *VolumeBuildSource) HasDockerfilePath() bool`

HasDockerfilePath returns a boolean if a field has been set.

### GetBuildContext

`func (o *VolumeBuildSource) GetBuildContext() string`

GetBuildContext returns the BuildContext field if non-nil, zero value otherwise.

### GetBuildContextOk

`func (o *VolumeBuildSource) GetBuildContextOk() (*string, bool)`

GetBuildContextOk returns a tuple with the BuildContext field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildContext

`func (o *VolumeBuildSource) SetBuildContext(v string)`

SetBuildContext sets BuildContext field to given value.

### HasBuildContext

`func (o *VolumeBuildSource) HasBuildContext() bool`

HasBuildContext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


