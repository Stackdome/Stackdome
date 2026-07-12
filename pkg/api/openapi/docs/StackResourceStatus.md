# StackResourceStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PublicIngress** | Pointer to [**[]Ingress**](Ingress.md) |  | [optional] 
**InternalServiceName** | Pointer to **string** |  | [optional] 
**LastRestartRequestProcessedAt** | Pointer to **time.Time** |  | [optional] 
**State** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**ObservedRevision** | Pointer to **string** |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**LastFailure** | Pointer to [**StackResourceFailure**](StackResourceFailure.md) |  | [optional] 
**Replicas** | Pointer to **int32** |  | [optional] [readonly] 
**AvailableReplicas** | Pointer to **int32** |  | [optional] [readonly] 
**UpdatedReplicas** | Pointer to **int32** |  | [optional] [readonly] 
**LastRunTime** | Pointer to **time.Time** |  | [optional] [readonly] 
**LastRunSucceeded** | Pointer to **bool** |  | [optional] [readonly] 

## Methods

### NewStackResourceStatus

`func NewStackResourceStatus() *StackResourceStatus`

NewStackResourceStatus instantiates a new StackResourceStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackResourceStatusWithDefaults

`func NewStackResourceStatusWithDefaults() *StackResourceStatus`

NewStackResourceStatusWithDefaults instantiates a new StackResourceStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPublicIngress

`func (o *StackResourceStatus) GetPublicIngress() []Ingress`

GetPublicIngress returns the PublicIngress field if non-nil, zero value otherwise.

### GetPublicIngressOk

`func (o *StackResourceStatus) GetPublicIngressOk() (*[]Ingress, bool)`

GetPublicIngressOk returns a tuple with the PublicIngress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicIngress

`func (o *StackResourceStatus) SetPublicIngress(v []Ingress)`

SetPublicIngress sets PublicIngress field to given value.

### HasPublicIngress

`func (o *StackResourceStatus) HasPublicIngress() bool`

HasPublicIngress returns a boolean if a field has been set.

### GetInternalServiceName

`func (o *StackResourceStatus) GetInternalServiceName() string`

GetInternalServiceName returns the InternalServiceName field if non-nil, zero value otherwise.

### GetInternalServiceNameOk

`func (o *StackResourceStatus) GetInternalServiceNameOk() (*string, bool)`

GetInternalServiceNameOk returns a tuple with the InternalServiceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInternalServiceName

`func (o *StackResourceStatus) SetInternalServiceName(v string)`

SetInternalServiceName sets InternalServiceName field to given value.

### HasInternalServiceName

`func (o *StackResourceStatus) HasInternalServiceName() bool`

HasInternalServiceName returns a boolean if a field has been set.

### GetLastRestartRequestProcessedAt

`func (o *StackResourceStatus) GetLastRestartRequestProcessedAt() time.Time`

GetLastRestartRequestProcessedAt returns the LastRestartRequestProcessedAt field if non-nil, zero value otherwise.

### GetLastRestartRequestProcessedAtOk

`func (o *StackResourceStatus) GetLastRestartRequestProcessedAtOk() (*time.Time, bool)`

GetLastRestartRequestProcessedAtOk returns a tuple with the LastRestartRequestProcessedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastRestartRequestProcessedAt

`func (o *StackResourceStatus) SetLastRestartRequestProcessedAt(v time.Time)`

SetLastRestartRequestProcessedAt sets LastRestartRequestProcessedAt field to given value.

### HasLastRestartRequestProcessedAt

`func (o *StackResourceStatus) HasLastRestartRequestProcessedAt() bool`

HasLastRestartRequestProcessedAt returns a boolean if a field has been set.

### GetState

`func (o *StackResourceStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *StackResourceStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *StackResourceStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *StackResourceStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetMessage

`func (o *StackResourceStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *StackResourceStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *StackResourceStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *StackResourceStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetObservedRevision

`func (o *StackResourceStatus) GetObservedRevision() string`

GetObservedRevision returns the ObservedRevision field if non-nil, zero value otherwise.

### GetObservedRevisionOk

`func (o *StackResourceStatus) GetObservedRevisionOk() (*string, bool)`

GetObservedRevisionOk returns a tuple with the ObservedRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedRevision

`func (o *StackResourceStatus) SetObservedRevision(v string)`

SetObservedRevision sets ObservedRevision field to given value.

### HasObservedRevision

`func (o *StackResourceStatus) HasObservedRevision() bool`

HasObservedRevision returns a boolean if a field has been set.

### GetConditions

`func (o *StackResourceStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *StackResourceStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *StackResourceStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *StackResourceStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetLastFailure

`func (o *StackResourceStatus) GetLastFailure() StackResourceFailure`

GetLastFailure returns the LastFailure field if non-nil, zero value otherwise.

### GetLastFailureOk

`func (o *StackResourceStatus) GetLastFailureOk() (*StackResourceFailure, bool)`

GetLastFailureOk returns a tuple with the LastFailure field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastFailure

`func (o *StackResourceStatus) SetLastFailure(v StackResourceFailure)`

SetLastFailure sets LastFailure field to given value.

### HasLastFailure

`func (o *StackResourceStatus) HasLastFailure() bool`

HasLastFailure returns a boolean if a field has been set.

### GetReplicas

`func (o *StackResourceStatus) GetReplicas() int32`

GetReplicas returns the Replicas field if non-nil, zero value otherwise.

### GetReplicasOk

`func (o *StackResourceStatus) GetReplicasOk() (*int32, bool)`

GetReplicasOk returns a tuple with the Replicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReplicas

`func (o *StackResourceStatus) SetReplicas(v int32)`

SetReplicas sets Replicas field to given value.

### HasReplicas

`func (o *StackResourceStatus) HasReplicas() bool`

HasReplicas returns a boolean if a field has been set.

### GetAvailableReplicas

`func (o *StackResourceStatus) GetAvailableReplicas() int32`

GetAvailableReplicas returns the AvailableReplicas field if non-nil, zero value otherwise.

### GetAvailableReplicasOk

`func (o *StackResourceStatus) GetAvailableReplicasOk() (*int32, bool)`

GetAvailableReplicasOk returns a tuple with the AvailableReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAvailableReplicas

`func (o *StackResourceStatus) SetAvailableReplicas(v int32)`

SetAvailableReplicas sets AvailableReplicas field to given value.

### HasAvailableReplicas

`func (o *StackResourceStatus) HasAvailableReplicas() bool`

HasAvailableReplicas returns a boolean if a field has been set.

### GetUpdatedReplicas

`func (o *StackResourceStatus) GetUpdatedReplicas() int32`

GetUpdatedReplicas returns the UpdatedReplicas field if non-nil, zero value otherwise.

### GetUpdatedReplicasOk

`func (o *StackResourceStatus) GetUpdatedReplicasOk() (*int32, bool)`

GetUpdatedReplicasOk returns a tuple with the UpdatedReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedReplicas

`func (o *StackResourceStatus) SetUpdatedReplicas(v int32)`

SetUpdatedReplicas sets UpdatedReplicas field to given value.

### HasUpdatedReplicas

`func (o *StackResourceStatus) HasUpdatedReplicas() bool`

HasUpdatedReplicas returns a boolean if a field has been set.

### GetLastRunTime

`func (o *StackResourceStatus) GetLastRunTime() time.Time`

GetLastRunTime returns the LastRunTime field if non-nil, zero value otherwise.

### GetLastRunTimeOk

`func (o *StackResourceStatus) GetLastRunTimeOk() (*time.Time, bool)`

GetLastRunTimeOk returns a tuple with the LastRunTime field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastRunTime

`func (o *StackResourceStatus) SetLastRunTime(v time.Time)`

SetLastRunTime sets LastRunTime field to given value.

### HasLastRunTime

`func (o *StackResourceStatus) HasLastRunTime() bool`

HasLastRunTime returns a boolean if a field has been set.

### GetLastRunSucceeded

`func (o *StackResourceStatus) GetLastRunSucceeded() bool`

GetLastRunSucceeded returns the LastRunSucceeded field if non-nil, zero value otherwise.

### GetLastRunSucceededOk

`func (o *StackResourceStatus) GetLastRunSucceededOk() (*bool, bool)`

GetLastRunSucceededOk returns a tuple with the LastRunSucceeded field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLastRunSucceeded

`func (o *StackResourceStatus) SetLastRunSucceeded(v bool)`

SetLastRunSucceeded sets LastRunSucceeded field to given value.

### HasLastRunSucceeded

`func (o *StackResourceStatus) HasLastRunSucceeded() bool`

HasLastRunSucceeded returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


