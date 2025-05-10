# VolumeStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**Phase** | Pointer to **string** |  | [optional] 
**BuildArtifactSyncs** | Pointer to [**[]BuildArtifactSyncInfo**](BuildArtifactSyncInfo.md) |  | [optional] 
**LastSyncedGitRevision** | Pointer to **string** |  | [optional] 
**LastRemoteSyncHash** | Pointer to **string** |  | [optional] 

## Methods

### NewVolumeStatus

`func NewVolumeStatus() *VolumeStatus`

NewVolumeStatus instantiates a new VolumeStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewVolumeStatusWithDefaults

`func NewVolumeStatusWithDefaults() *VolumeStatus`

NewVolumeStatusWithDefaults instantiates a new VolumeStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConditions

`func (o *VolumeStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *VolumeStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *VolumeStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *VolumeStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetPhase

`func (o *VolumeStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *VolumeStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *VolumeStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *VolumeStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetBuildArtifactSyncs

`func (o *VolumeStatus) GetBuildArtifactSyncs() []BuildArtifactSyncInfo`

GetBuildArtifactSyncs returns the BuildArtifactSyncs field if non-nil, zero value otherwise.

### GetBuildArtifactSyncsOk

`func (o *VolumeStatus) GetBuildArtifactSyncsOk() (*[]BuildArtifactSyncInfo, bool)`

GetBuildArtifactSyncsOk returns a tuple with the BuildArtifactSyncs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildArtifactSyncs

`func (o *VolumeStatus) SetBuildArtifactSyncs(v []BuildArtifactSyncInfo)`

SetBuildArtifactSyncs sets BuildArtifactSyncs field to given value.

### HasBuildArtifactSyncs

`func (o *VolumeStatus) HasBuildArtifactSyncs() bool`

HasBuildArtifactSyncs returns a boolean if a field has been set.

### GetLastSyncedGitRevision

`func (o *VolumeStatus) GetLastSyncedGitRevision() string`

GetLastSyncedGitRevision returns the LastSyncedGitRevision field if non-nil, zero value otherwise.

### GetLastSyncedGitRevisionOk

`func (o *VolumeStatus) GetLastSyncedGitRevisionOk() (*string, bool)`

GetLastSyncedGitRevisionOk returns a tuple with the LastSyncedGitRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastSyncedGitRevision

`func (o *VolumeStatus) SetLastSyncedGitRevision(v string)`

SetLastSyncedGitRevision sets LastSyncedGitRevision field to given value.

### HasLastSyncedGitRevision

`func (o *VolumeStatus) HasLastSyncedGitRevision() bool`

HasLastSyncedGitRevision returns a boolean if a field has been set.

### GetLastRemoteSyncHash

`func (o *VolumeStatus) GetLastRemoteSyncHash() string`

GetLastRemoteSyncHash returns the LastRemoteSyncHash field if non-nil, zero value otherwise.

### GetLastRemoteSyncHashOk

`func (o *VolumeStatus) GetLastRemoteSyncHashOk() (*string, bool)`

GetLastRemoteSyncHashOk returns a tuple with the LastRemoteSyncHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastRemoteSyncHash

`func (o *VolumeStatus) SetLastRemoteSyncHash(v string)`

SetLastRemoteSyncHash sets LastRemoteSyncHash field to given value.

### HasLastRemoteSyncHash

`func (o *VolumeStatus) HasLastRemoteSyncHash() bool`

HasLastRemoteSyncHash returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


