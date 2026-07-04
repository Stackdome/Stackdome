# SourceSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Git** | Pointer to [**GitSource**](GitSource.md) |  | [optional] 
**Image** | Pointer to [**ImageSource**](ImageSource.md) |  | [optional] 
**Volume** | Pointer to [**VolumeBuildSource**](VolumeBuildSource.md) |  | [optional] 

## Methods

### NewSourceSpec

`func NewSourceSpec() *SourceSpec`

NewSourceSpec instantiates a new SourceSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSourceSpecWithDefaults

`func NewSourceSpecWithDefaults() *SourceSpec`

NewSourceSpecWithDefaults instantiates a new SourceSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGit

`func (o *SourceSpec) GetGit() GitSource`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *SourceSpec) GetGitOk() (*GitSource, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *SourceSpec) SetGit(v GitSource)`

SetGit sets Git field to given value.

### HasGit

`func (o *SourceSpec) HasGit() bool`

HasGit returns a boolean if a field has been set.

### GetImage

`func (o *SourceSpec) GetImage() ImageSource`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *SourceSpec) GetImageOk() (*ImageSource, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *SourceSpec) SetImage(v ImageSource)`

SetImage sets Image field to given value.

### HasImage

`func (o *SourceSpec) HasImage() bool`

HasImage returns a boolean if a field has been set.

### GetVolume

`func (o *SourceSpec) GetVolume() VolumeBuildSource`

GetVolume returns the Volume field if non-nil, zero value otherwise.

### GetVolumeOk

`func (o *SourceSpec) GetVolumeOk() (*VolumeBuildSource, bool)`

GetVolumeOk returns a tuple with the Volume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolume

`func (o *SourceSpec) SetVolume(v VolumeBuildSource)`

SetVolume sets Volume field to given value.

### HasVolume

`func (o *SourceSpec) HasVolume() bool`

HasVolume returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


