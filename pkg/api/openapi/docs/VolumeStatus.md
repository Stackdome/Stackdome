# VolumeStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**Phase** | Pointer to **string** |  | [optional] 
**BuildArtifactSyncs** | Pointer to [**[]BuildArtifactSyncInfo**](BuildArtifactSyncInfo.md) |  | [optional] 

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


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


