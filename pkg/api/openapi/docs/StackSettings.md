# StackSettings

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ReleaseRetentionLimit** | Pointer to **int32** | Maximum total releases to retain per stack | [optional] [default to 10]
**MinSuccessfulReleases** | Pointer to **int32** | Minimum number of successful releases to always keep | [optional] [default to 5]
**DeployTimeoutMinutes** | Pointer to **int32** | Minutes to wait for a deploy to converge before marking it failed | [optional] [default to 15]

## Methods

### NewStackSettings

`func NewStackSettings() *StackSettings`

NewStackSettings instantiates a new StackSettings object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackSettingsWithDefaults

`func NewStackSettingsWithDefaults() *StackSettings`

NewStackSettingsWithDefaults instantiates a new StackSettings object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetReleaseRetentionLimit

`func (o *StackSettings) GetReleaseRetentionLimit() int32`

GetReleaseRetentionLimit returns the ReleaseRetentionLimit field if non-nil, zero value otherwise.

### GetReleaseRetentionLimitOk

`func (o *StackSettings) GetReleaseRetentionLimitOk() (*int32, bool)`

GetReleaseRetentionLimitOk returns a tuple with the ReleaseRetentionLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReleaseRetentionLimit

`func (o *StackSettings) SetReleaseRetentionLimit(v int32)`

SetReleaseRetentionLimit sets ReleaseRetentionLimit field to given value.

### HasReleaseRetentionLimit

`func (o *StackSettings) HasReleaseRetentionLimit() bool`

HasReleaseRetentionLimit returns a boolean if a field has been set.

### GetMinSuccessfulReleases

`func (o *StackSettings) GetMinSuccessfulReleases() int32`

GetMinSuccessfulReleases returns the MinSuccessfulReleases field if non-nil, zero value otherwise.

### GetMinSuccessfulReleasesOk

`func (o *StackSettings) GetMinSuccessfulReleasesOk() (*int32, bool)`

GetMinSuccessfulReleasesOk returns a tuple with the MinSuccessfulReleases field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinSuccessfulReleases

`func (o *StackSettings) SetMinSuccessfulReleases(v int32)`

SetMinSuccessfulReleases sets MinSuccessfulReleases field to given value.

### HasMinSuccessfulReleases

`func (o *StackSettings) HasMinSuccessfulReleases() bool`

HasMinSuccessfulReleases returns a boolean if a field has been set.

### GetDeployTimeoutMinutes

`func (o *StackSettings) GetDeployTimeoutMinutes() int32`

GetDeployTimeoutMinutes returns the DeployTimeoutMinutes field if non-nil, zero value otherwise.

### GetDeployTimeoutMinutesOk

`func (o *StackSettings) GetDeployTimeoutMinutesOk() (*int32, bool)`

GetDeployTimeoutMinutesOk returns a tuple with the DeployTimeoutMinutes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeployTimeoutMinutes

`func (o *StackSettings) SetDeployTimeoutMinutes(v int32)`

SetDeployTimeoutMinutes sets DeployTimeoutMinutes field to given value.

### HasDeployTimeoutMinutes

`func (o *StackSettings) HasDeployTimeoutMinutes() bool`

HasDeployTimeoutMinutes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


