# StackResourceStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PublicIngress** | Pointer to [**[]Ingress**](Ingress.md) |  | [optional] 
**InternalServiceName** | Pointer to **string** |  | [optional] 
**LastRestartRequestProcessedAt** | Pointer to **time.Time** |  | [optional] 
**State** | Pointer to **string** |  | [optional] 
**ObservedRevision** | Pointer to **string** |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 

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


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


