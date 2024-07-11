# WorkspaceVolumeSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | **string** |  | 
**StorageClass** | Pointer to **string** |  | [optional] 
**SyncBeforeUse** | Pointer to **bool** |  | [optional] 
**Source** | Pointer to [**VolumeSource**](VolumeSource.md) |  | [optional] 

## Methods

### NewWorkspaceVolumeSpec

`func NewWorkspaceVolumeSpec(size string, ) *WorkspaceVolumeSpec`

NewWorkspaceVolumeSpec instantiates a new WorkspaceVolumeSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceVolumeSpecWithDefaults

`func NewWorkspaceVolumeSpecWithDefaults() *WorkspaceVolumeSpec`

NewWorkspaceVolumeSpecWithDefaults instantiates a new WorkspaceVolumeSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSize

`func (o *WorkspaceVolumeSpec) GetSize() string`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *WorkspaceVolumeSpec) GetSizeOk() (*string, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *WorkspaceVolumeSpec) SetSize(v string)`

SetSize sets Size field to given value.


### GetStorageClass

`func (o *WorkspaceVolumeSpec) GetStorageClass() string`

GetStorageClass returns the StorageClass field if non-nil, zero value otherwise.

### GetStorageClassOk

`func (o *WorkspaceVolumeSpec) GetStorageClassOk() (*string, bool)`

GetStorageClassOk returns a tuple with the StorageClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClass

`func (o *WorkspaceVolumeSpec) SetStorageClass(v string)`

SetStorageClass sets StorageClass field to given value.

### HasStorageClass

`func (o *WorkspaceVolumeSpec) HasStorageClass() bool`

HasStorageClass returns a boolean if a field has been set.

### GetSyncBeforeUse

`func (o *WorkspaceVolumeSpec) GetSyncBeforeUse() bool`

GetSyncBeforeUse returns the SyncBeforeUse field if non-nil, zero value otherwise.

### GetSyncBeforeUseOk

`func (o *WorkspaceVolumeSpec) GetSyncBeforeUseOk() (*bool, bool)`

GetSyncBeforeUseOk returns a tuple with the SyncBeforeUse field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSyncBeforeUse

`func (o *WorkspaceVolumeSpec) SetSyncBeforeUse(v bool)`

SetSyncBeforeUse sets SyncBeforeUse field to given value.

### HasSyncBeforeUse

`func (o *WorkspaceVolumeSpec) HasSyncBeforeUse() bool`

HasSyncBeforeUse returns a boolean if a field has been set.

### GetSource

`func (o *WorkspaceVolumeSpec) GetSource() VolumeSource`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *WorkspaceVolumeSpec) GetSourceOk() (*VolumeSource, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *WorkspaceVolumeSpec) SetSource(v VolumeSource)`

SetSource sets Source field to given value.

### HasSource

`func (o *WorkspaceVolumeSpec) HasSource() bool`

HasSource returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


