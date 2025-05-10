# VolumeSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | **string** |  | 
**StorageClass** | Pointer to **string** |  | [optional] 
**NeedsSyncBeforeUse** | **bool** |  | 
**AccessMode** | [**VolumeAccessMode**](VolumeAccessMode.md) |  | 
**Source** | Pointer to [**VolumeSource**](VolumeSource.md) |  | [optional] 

## Methods

### NewVolumeSpec

`func NewVolumeSpec(size string, needsSyncBeforeUse bool, accessMode VolumeAccessMode, ) *VolumeSpec`

NewVolumeSpec instantiates a new VolumeSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeSpecWithDefaults

`func NewVolumeSpecWithDefaults() *VolumeSpec`

NewVolumeSpecWithDefaults instantiates a new VolumeSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSize

`func (o *VolumeSpec) GetSize() string`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *VolumeSpec) GetSizeOk() (*string, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *VolumeSpec) SetSize(v string)`

SetSize sets Size field to given value.


### GetStorageClass

`func (o *VolumeSpec) GetStorageClass() string`

GetStorageClass returns the StorageClass field if non-nil, zero value otherwise.

### GetStorageClassOk

`func (o *VolumeSpec) GetStorageClassOk() (*string, bool)`

GetStorageClassOk returns a tuple with the StorageClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClass

`func (o *VolumeSpec) SetStorageClass(v string)`

SetStorageClass sets StorageClass field to given value.

### HasStorageClass

`func (o *VolumeSpec) HasStorageClass() bool`

HasStorageClass returns a boolean if a field has been set.

### GetNeedsSyncBeforeUse

`func (o *VolumeSpec) GetNeedsSyncBeforeUse() bool`

GetNeedsSyncBeforeUse returns the NeedsSyncBeforeUse field if non-nil, zero value otherwise.

### GetNeedsSyncBeforeUseOk

`func (o *VolumeSpec) GetNeedsSyncBeforeUseOk() (*bool, bool)`

GetNeedsSyncBeforeUseOk returns a tuple with the NeedsSyncBeforeUse field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNeedsSyncBeforeUse

`func (o *VolumeSpec) SetNeedsSyncBeforeUse(v bool)`

SetNeedsSyncBeforeUse sets NeedsSyncBeforeUse field to given value.


### GetAccessMode

`func (o *VolumeSpec) GetAccessMode() VolumeAccessMode`

GetAccessMode returns the AccessMode field if non-nil, zero value otherwise.

### GetAccessModeOk

`func (o *VolumeSpec) GetAccessModeOk() (*VolumeAccessMode, bool)`

GetAccessModeOk returns a tuple with the AccessMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessMode

`func (o *VolumeSpec) SetAccessMode(v VolumeAccessMode)`

SetAccessMode sets AccessMode field to given value.


### GetSource

`func (o *VolumeSpec) GetSource() VolumeSource`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *VolumeSpec) GetSourceOk() (*VolumeSource, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *VolumeSpec) SetSource(v VolumeSource)`

SetSource sets Source field to given value.

### HasSource

`func (o *VolumeSpec) HasSource() bool`

HasSource returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


