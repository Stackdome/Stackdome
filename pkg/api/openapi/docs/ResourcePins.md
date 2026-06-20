# ResourcePins

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GitSha** | Pointer to **string** |  | [optional] 
**VolumeHash** | Pointer to **string** |  | [optional] 
**ImageDigest** | Pointer to **string** |  | [optional] 

## Methods

### NewResourcePins

`func NewResourcePins() *ResourcePins`

NewResourcePins instantiates a new ResourcePins object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourcePinsWithDefaults

`func NewResourcePinsWithDefaults() *ResourcePins`

NewResourcePinsWithDefaults instantiates a new ResourcePins object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGitSha

`func (o *ResourcePins) GetGitSha() string`

GetGitSha returns the GitSha field if non-nil, zero value otherwise.

### GetGitShaOk

`func (o *ResourcePins) GetGitShaOk() (*string, bool)`

GetGitShaOk returns a tuple with the GitSha field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitSha

`func (o *ResourcePins) SetGitSha(v string)`

SetGitSha sets GitSha field to given value.

### HasGitSha

`func (o *ResourcePins) HasGitSha() bool`

HasGitSha returns a boolean if a field has been set.

### GetVolumeHash

`func (o *ResourcePins) GetVolumeHash() string`

GetVolumeHash returns the VolumeHash field if non-nil, zero value otherwise.

### GetVolumeHashOk

`func (o *ResourcePins) GetVolumeHashOk() (*string, bool)`

GetVolumeHashOk returns a tuple with the VolumeHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeHash

`func (o *ResourcePins) SetVolumeHash(v string)`

SetVolumeHash sets VolumeHash field to given value.

### HasVolumeHash

`func (o *ResourcePins) HasVolumeHash() bool`

HasVolumeHash returns a boolean if a field has been set.

### GetImageDigest

`func (o *ResourcePins) GetImageDigest() string`

GetImageDigest returns the ImageDigest field if non-nil, zero value otherwise.

### GetImageDigestOk

`func (o *ResourcePins) GetImageDigestOk() (*string, bool)`

GetImageDigestOk returns a tuple with the ImageDigest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageDigest

`func (o *ResourcePins) SetImageDigest(v string)`

SetImageDigest sets ImageDigest field to given value.

### HasImageDigest

`func (o *ResourcePins) HasImageDigest() bool`

HasImageDigest returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


