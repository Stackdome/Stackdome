# StackResourceSummary

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** |  | [optional] 
**Phase** | Pointer to **string** |  | [optional] 
**ObservedRevision** | Pointer to **string** |  | [optional] 
**ConvergedRevision** | Pointer to **string** |  | [optional] 
**AvailableReplicas** | Pointer to **int32** |  | [optional] 
**UpdatedReplicas** | Pointer to **int32** |  | [optional] 
**Replicas** | Pointer to **int32** |  | [optional] 
**Missing** | Pointer to **bool** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewStackResourceSummary

`func NewStackResourceSummary() *StackResourceSummary`

NewStackResourceSummary instantiates a new StackResourceSummary object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackResourceSummaryWithDefaults

`func NewStackResourceSummaryWithDefaults() *StackResourceSummary`

NewStackResourceSummaryWithDefaults instantiates a new StackResourceSummary object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *StackResourceSummary) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *StackResourceSummary) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *StackResourceSummary) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *StackResourceSummary) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPhase

`func (o *StackResourceSummary) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *StackResourceSummary) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *StackResourceSummary) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *StackResourceSummary) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetObservedRevision

`func (o *StackResourceSummary) GetObservedRevision() string`

GetObservedRevision returns the ObservedRevision field if non-nil, zero value otherwise.

### GetObservedRevisionOk

`func (o *StackResourceSummary) GetObservedRevisionOk() (*string, bool)`

GetObservedRevisionOk returns a tuple with the ObservedRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedRevision

`func (o *StackResourceSummary) SetObservedRevision(v string)`

SetObservedRevision sets ObservedRevision field to given value.

### HasObservedRevision

`func (o *StackResourceSummary) HasObservedRevision() bool`

HasObservedRevision returns a boolean if a field has been set.

### GetConvergedRevision

`func (o *StackResourceSummary) GetConvergedRevision() string`

GetConvergedRevision returns the ConvergedRevision field if non-nil, zero value otherwise.

### GetConvergedRevisionOk

`func (o *StackResourceSummary) GetConvergedRevisionOk() (*string, bool)`

GetConvergedRevisionOk returns a tuple with the ConvergedRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConvergedRevision

`func (o *StackResourceSummary) SetConvergedRevision(v string)`

SetConvergedRevision sets ConvergedRevision field to given value.

### HasConvergedRevision

`func (o *StackResourceSummary) HasConvergedRevision() bool`

HasConvergedRevision returns a boolean if a field has been set.

### GetAvailableReplicas

`func (o *StackResourceSummary) GetAvailableReplicas() int32`

GetAvailableReplicas returns the AvailableReplicas field if non-nil, zero value otherwise.

### GetAvailableReplicasOk

`func (o *StackResourceSummary) GetAvailableReplicasOk() (*int32, bool)`

GetAvailableReplicasOk returns a tuple with the AvailableReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvailableReplicas

`func (o *StackResourceSummary) SetAvailableReplicas(v int32)`

SetAvailableReplicas sets AvailableReplicas field to given value.

### HasAvailableReplicas

`func (o *StackResourceSummary) HasAvailableReplicas() bool`

HasAvailableReplicas returns a boolean if a field has been set.

### GetUpdatedReplicas

`func (o *StackResourceSummary) GetUpdatedReplicas() int32`

GetUpdatedReplicas returns the UpdatedReplicas field if non-nil, zero value otherwise.

### GetUpdatedReplicasOk

`func (o *StackResourceSummary) GetUpdatedReplicasOk() (*int32, bool)`

GetUpdatedReplicasOk returns a tuple with the UpdatedReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedReplicas

`func (o *StackResourceSummary) SetUpdatedReplicas(v int32)`

SetUpdatedReplicas sets UpdatedReplicas field to given value.

### HasUpdatedReplicas

`func (o *StackResourceSummary) HasUpdatedReplicas() bool`

HasUpdatedReplicas returns a boolean if a field has been set.

### GetReplicas

`func (o *StackResourceSummary) GetReplicas() int32`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *StackResourceSummary) GetReplicasOk() (*int32, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *StackResourceSummary) SetReplicas(v int32)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *StackResourceSummary) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### GetMissing

`func (o *StackResourceSummary) GetMissing() bool`

GetMissing returns the Missing field if non-nil, zero value otherwise.

### GetMissingOk

`func (o *StackResourceSummary) GetMissingOk() (*bool, bool)`

GetMissingOk returns a tuple with the Missing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMissing

`func (o *StackResourceSummary) SetMissing(v bool)`

SetMissing sets Missing field to given value.

### HasMissing

`func (o *StackResourceSummary) HasMissing() bool`

HasMissing returns a boolean if a field has been set.

### GetMessage

`func (o *StackResourceSummary) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *StackResourceSummary) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *StackResourceSummary) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *StackResourceSummary) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


