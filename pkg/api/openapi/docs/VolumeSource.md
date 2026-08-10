# VolumeSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GitRepoSource** | [**GitRepoSource**](GitRepoSource.md) |  |

## Methods

### NewVolumeSource

`func NewVolumeSource(gitRepoSource GitRepoSource, ) *VolumeSource`

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



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

