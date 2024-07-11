# VolumeSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceType** | [**VolumeSourceTypes**](VolumeSourceTypes.md) |  | 
**LocalSource** | Pointer to [**LocalSource**](LocalSource.md) |  | [optional] 
**BuildSource** | Pointer to [**[]BuildArtifact**](BuildArtifact.md) |  | [optional] 

## Methods

### NewVolumeSource

`func NewVolumeSource(sourceType VolumeSourceTypes, ) *VolumeSource`

NewVolumeSource instantiates a new VolumeSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeSourceWithDefaults

`func NewVolumeSourceWithDefaults() *VolumeSource`

NewVolumeSourceWithDefaults instantiates a new VolumeSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceType

`func (o *VolumeSource) GetSourceType() VolumeSourceTypes`

GetSourceType returns the SourceType field if non-nil, zero value otherwise.

### GetSourceTypeOk

`func (o *VolumeSource) GetSourceTypeOk() (*VolumeSourceTypes, bool)`

GetSourceTypeOk returns a tuple with the SourceType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceType

`func (o *VolumeSource) SetSourceType(v VolumeSourceTypes)`

SetSourceType sets SourceType field to given value.


### GetLocalSource

`func (o *VolumeSource) GetLocalSource() LocalSource`

GetLocalSource returns the LocalSource field if non-nil, zero value otherwise.

### GetLocalSourceOk

`func (o *VolumeSource) GetLocalSourceOk() (*LocalSource, bool)`

GetLocalSourceOk returns a tuple with the LocalSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLocalSource

`func (o *VolumeSource) SetLocalSource(v LocalSource)`

SetLocalSource sets LocalSource field to given value.

### HasLocalSource

`func (o *VolumeSource) HasLocalSource() bool`

HasLocalSource returns a boolean if a field has been set.

### GetBuildSource

`func (o *VolumeSource) GetBuildSource() []BuildArtifact`

GetBuildSource returns the BuildSource field if non-nil, zero value otherwise.

### GetBuildSourceOk

`func (o *VolumeSource) GetBuildSourceOk() (*[]BuildArtifact, bool)`

GetBuildSourceOk returns a tuple with the BuildSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildSource

`func (o *VolumeSource) SetBuildSource(v []BuildArtifact)`

SetBuildSource sets BuildSource field to given value.

### HasBuildSource

`func (o *VolumeSource) HasBuildSource() bool`

HasBuildSource returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


