# StackStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**ObservedRevision** | Pointer to **string** |  | [optional] 
**TargetRevision** | Pointer to **string** |  | [optional] 
**LastConverged** | Pointer to [**StackConvergenceRecord**](StackConvergenceRecord.md) |  | [optional] 
**Resources** | Pointer to [**[]StackResourceSummary**](StackResourceSummary.md) |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 

## Methods

### NewStackStatus

`func NewStackStatus() *StackStatus`

NewStackStatus instantiates a new StackStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackStatusWithDefaults

`func NewStackStatusWithDefaults() *StackStatus`

NewStackStatusWithDefaults instantiates a new StackStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *StackStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *StackStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *StackStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *StackStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetMessage

`func (o *StackStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *StackStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *StackStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *StackStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetObservedRevision

`func (o *StackStatus) GetObservedRevision() string`

GetObservedRevision returns the ObservedRevision field if non-nil, zero value otherwise.

### GetObservedRevisionOk

`func (o *StackStatus) GetObservedRevisionOk() (*string, bool)`

GetObservedRevisionOk returns a tuple with the ObservedRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedRevision

`func (o *StackStatus) SetObservedRevision(v string)`

SetObservedRevision sets ObservedRevision field to given value.

### HasObservedRevision

`func (o *StackStatus) HasObservedRevision() bool`

HasObservedRevision returns a boolean if a field has been set.

### GetTargetRevision

`func (o *StackStatus) GetTargetRevision() string`

GetTargetRevision returns the TargetRevision field if non-nil, zero value otherwise.

### GetTargetRevisionOk

`func (o *StackStatus) GetTargetRevisionOk() (*string, bool)`

GetTargetRevisionOk returns a tuple with the TargetRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetRevision

`func (o *StackStatus) SetTargetRevision(v string)`

SetTargetRevision sets TargetRevision field to given value.

### HasTargetRevision

`func (o *StackStatus) HasTargetRevision() bool`

HasTargetRevision returns a boolean if a field has been set.

### GetLastConverged

`func (o *StackStatus) GetLastConverged() StackConvergenceRecord`

GetLastConverged returns the LastConverged field if non-nil, zero value otherwise.

### GetLastConvergedOk

`func (o *StackStatus) GetLastConvergedOk() (*StackConvergenceRecord, bool)`

GetLastConvergedOk returns a tuple with the LastConverged field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastConverged

`func (o *StackStatus) SetLastConverged(v StackConvergenceRecord)`

SetLastConverged sets LastConverged field to given value.

### HasLastConverged

`func (o *StackStatus) HasLastConverged() bool`

HasLastConverged returns a boolean if a field has been set.

### GetResources

`func (o *StackStatus) GetResources() []StackResourceSummary`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *StackStatus) GetResourcesOk() (*[]StackResourceSummary, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *StackStatus) SetResources(v []StackResourceSummary)`

SetResources sets Resources field to given value.

### HasResources

`func (o *StackStatus) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetConditions

`func (o *StackStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *StackStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *StackStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *StackStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


