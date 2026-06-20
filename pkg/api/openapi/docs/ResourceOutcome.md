# ResourceOutcome

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | Pointer to **string** |  | [optional] 
**ReadyReplicas** | Pointer to **int32** |  | [optional] 
**Replicas** | Pointer to **int32** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewResourceOutcome

`func NewResourceOutcome() *ResourceOutcome`

NewResourceOutcome instantiates a new ResourceOutcome object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceOutcomeWithDefaults

`func NewResourceOutcomeWithDefaults() *ResourceOutcome`

NewResourceOutcomeWithDefaults instantiates a new ResourceOutcome object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *ResourceOutcome) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *ResourceOutcome) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *ResourceOutcome) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *ResourceOutcome) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetReadyReplicas

`func (o *ResourceOutcome) GetReadyReplicas() int32`

GetReadyReplicas returns the ReadyReplicas field if non-nil, zero value otherwise.

### GetReadyReplicasOk

`func (o *ResourceOutcome) GetReadyReplicasOk() (*int32, bool)`

GetReadyReplicasOk returns a tuple with the ReadyReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadyReplicas

`func (o *ResourceOutcome) SetReadyReplicas(v int32)`

SetReadyReplicas sets ReadyReplicas field to given value.

### HasReadyReplicas

`func (o *ResourceOutcome) HasReadyReplicas() bool`

HasReadyReplicas returns a boolean if a field has been set.

### GetReplicas

`func (o *ResourceOutcome) GetReplicas() int32`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *ResourceOutcome) GetReplicasOk() (*int32, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *ResourceOutcome) SetReplicas(v int32)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *ResourceOutcome) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### GetMessage

`func (o *ResourceOutcome) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ResourceOutcome) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ResourceOutcome) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ResourceOutcome) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


