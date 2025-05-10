# BuildSourceContext

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Volume** | Pointer to [**BuildSourceContextVolume**](BuildSourceContextVolume.md) |  | [optional] 
**GitRepo** | Pointer to [**BuildSourceContextGitRepo**](BuildSourceContextGitRepo.md) |  | [optional] 

## Methods

### NewBuildSourceContext

`func NewBuildSourceContext() *BuildSourceContext`

NewBuildSourceContext instantiates a new BuildSourceContext object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewBuildSourceContextWithDefaults

`func NewBuildSourceContextWithDefaults() *BuildSourceContext`

NewBuildSourceContextWithDefaults instantiates a new BuildSourceContext object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVolume

`func (o *BuildSourceContext) GetVolume() BuildSourceContextVolume`

GetVolume returns the Volume field if non-nil, zero value otherwise.

### GetVolumeOk

`func (o *BuildSourceContext) GetVolumeOk() (*BuildSourceContextVolume, bool)`

GetVolumeOk returns a tuple with the Volume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolume

`func (o *BuildSourceContext) SetVolume(v BuildSourceContextVolume)`

SetVolume sets Volume field to given value.

### HasVolume

`func (o *BuildSourceContext) HasVolume() bool`

HasVolume returns a boolean if a field has been set.

### GetGitRepo

`func (o *BuildSourceContext) GetGitRepo() BuildSourceContextGitRepo`

GetGitRepo returns the GitRepo field if non-nil, zero value otherwise.

### GetGitRepoOk

`func (o *BuildSourceContext) GetGitRepoOk() (*BuildSourceContextGitRepo, bool)`

GetGitRepoOk returns a tuple with the GitRepo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitRepo

`func (o *BuildSourceContext) SetGitRepo(v BuildSourceContextGitRepo)`

SetGitRepo sets GitRepo field to given value.

### HasGitRepo

`func (o *BuildSourceContext) HasGitRepo() bool`

HasGitRepo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


