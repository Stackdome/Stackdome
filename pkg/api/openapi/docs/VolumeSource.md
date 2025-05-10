# VolumeSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GitRepoSource** | Pointer to [**GitRepoSource**](GitRepoSource.md) |  | [optional] 
**SourceType** | [**VolumeSourceTypes**](VolumeSourceTypes.md) |  | 
**RemoteSource** | Pointer to [**RemoteSource**](RemoteSource.md) |  | [optional] 
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

### GetGitRepoSource

`func (o *VolumeSource) GetGitRepoSource() GitRepoSource`

GetGitRepoSource returns the GitRepoSource field if non-nil, zero value otherwise.

### GetGitRepoSourceOk

`func (o *VolumeSource) GetGitRepoSourceOk() (*GitRepoSource, bool)`

GetGitRepoSourceOk returns a tuple with the GitRepoSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitRepoSource

`func (o *VolumeSource) SetGitRepoSource(v GitRepoSource)`

SetGitRepoSource sets GitRepoSource field to given value.

### HasGitRepoSource

`func (o *VolumeSource) HasGitRepoSource() bool`

HasGitRepoSource returns a boolean if a field has been set.

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


### GetRemoteSource

`func (o *VolumeSource) GetRemoteSource() RemoteSource`

GetRemoteSource returns the RemoteSource field if non-nil, zero value otherwise.

### GetRemoteSourceOk

`func (o *VolumeSource) GetRemoteSourceOk() (*RemoteSource, bool)`

GetRemoteSourceOk returns a tuple with the RemoteSource field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRemoteSource

`func (o *VolumeSource) SetRemoteSource(v RemoteSource)`

SetRemoteSource sets RemoteSource field to given value.

### HasRemoteSource

`func (o *VolumeSource) HasRemoteSource() bool`

HasRemoteSource returns a boolean if a field has been set.

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


