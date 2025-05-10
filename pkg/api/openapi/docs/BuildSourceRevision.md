# BuildSourceRevision

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**VolumeSourceRevision** | Pointer to [**BuildSourceRevisionVolumeSourceRevision**](BuildSourceRevisionVolumeSourceRevision.md) |  | [optional] 
**GitRepoRevision** | Pointer to [**GitRepoRevision**](GitRepoRevision.md) |  | [optional] 

## Methods

### NewBuildSourceRevision

`func NewBuildSourceRevision() *BuildSourceRevision`

NewBuildSourceRevision instantiates a new BuildSourceRevision object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBuildSourceRevisionWithDefaults

`func NewBuildSourceRevisionWithDefaults() *BuildSourceRevision`

NewBuildSourceRevisionWithDefaults instantiates a new BuildSourceRevision object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVolumeSourceRevision

`func (o *BuildSourceRevision) GetVolumeSourceRevision() BuildSourceRevisionVolumeSourceRevision`

GetVolumeSourceRevision returns the VolumeSourceRevision field if non-nil, zero value otherwise.

### GetVolumeSourceRevisionOk

`func (o *BuildSourceRevision) GetVolumeSourceRevisionOk() (*BuildSourceRevisionVolumeSourceRevision, bool)`

GetVolumeSourceRevisionOk returns a tuple with the VolumeSourceRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeSourceRevision

`func (o *BuildSourceRevision) SetVolumeSourceRevision(v BuildSourceRevisionVolumeSourceRevision)`

SetVolumeSourceRevision sets VolumeSourceRevision field to given value.

### HasVolumeSourceRevision

`func (o *BuildSourceRevision) HasVolumeSourceRevision() bool`

HasVolumeSourceRevision returns a boolean if a field has been set.

### GetGitRepoRevision

`func (o *BuildSourceRevision) GetGitRepoRevision() GitRepoRevision`

GetGitRepoRevision returns the GitRepoRevision field if non-nil, zero value otherwise.

### GetGitRepoRevisionOk

`func (o *BuildSourceRevision) GetGitRepoRevisionOk() (*GitRepoRevision, bool)`

GetGitRepoRevisionOk returns a tuple with the GitRepoRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitRepoRevision

`func (o *BuildSourceRevision) SetGitRepoRevision(v GitRepoRevision)`

SetGitRepoRevision sets GitRepoRevision field to given value.

### HasGitRepoRevision

`func (o *BuildSourceRevision) HasGitRepoRevision() bool`

HasGitRepoRevision returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


