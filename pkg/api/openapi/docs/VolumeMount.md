# VolumeMount

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkspaceResourceId** | Pointer to **string** |  | [optional] [readonly] 
**WorkspaceStorageId** | Pointer to **string** |  | [optional] [readonly] 
**SourceVolumeType** | Pointer to [**VolumeMountSourceType**](VolumeMountSourceType.md) |  | [optional] 
**SourceVolumeId** | **string** |  | 
**SourceSubPath** | Pointer to **string** |  | [optional] 
**TargetPath** | **string** |  | 

## Methods

### NewVolumeMount

`func NewVolumeMount(sourceVolumeId string, targetPath string, ) *VolumeMount`

NewVolumeMount instantiates a new VolumeMount object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeMountWithDefaults

`func NewVolumeMountWithDefaults() *VolumeMount`

NewVolumeMountWithDefaults instantiates a new VolumeMount object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkspaceResourceId

`func (o *VolumeMount) GetWorkspaceResourceId() string`

GetWorkspaceResourceId returns the WorkspaceResourceId field if non-nil, zero value otherwise.

### GetWorkspaceResourceIdOk

`func (o *VolumeMount) GetWorkspaceResourceIdOk() (*string, bool)`

GetWorkspaceResourceIdOk returns a tuple with the WorkspaceResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceResourceId

`func (o *VolumeMount) SetWorkspaceResourceId(v string)`

SetWorkspaceResourceId sets WorkspaceResourceId field to given value.

### HasWorkspaceResourceId

`func (o *VolumeMount) HasWorkspaceResourceId() bool`

HasWorkspaceResourceId returns a boolean if a field has been set.

### GetWorkspaceStorageId

`func (o *VolumeMount) GetWorkspaceStorageId() string`

GetWorkspaceStorageId returns the WorkspaceStorageId field if non-nil, zero value otherwise.

### GetWorkspaceStorageIdOk

`func (o *VolumeMount) GetWorkspaceStorageIdOk() (*string, bool)`

GetWorkspaceStorageIdOk returns a tuple with the WorkspaceStorageId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceStorageId

`func (o *VolumeMount) SetWorkspaceStorageId(v string)`

SetWorkspaceStorageId sets WorkspaceStorageId field to given value.

### HasWorkspaceStorageId

`func (o *VolumeMount) HasWorkspaceStorageId() bool`

HasWorkspaceStorageId returns a boolean if a field has been set.

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

### GetSourceVolumeId

`func (o *VolumeMount) GetSourceVolumeId() string`

GetSourceVolumeId returns the SourceVolumeId field if non-nil, zero value otherwise.

### GetSourceVolumeIdOk

`func (o *VolumeMount) GetSourceVolumeIdOk() (*string, bool)`

GetSourceVolumeIdOk returns a tuple with the SourceVolumeId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceVolumeId

`func (o *VolumeMount) SetSourceVolumeId(v string)`

SetSourceVolumeId sets SourceVolumeId field to given value.


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


