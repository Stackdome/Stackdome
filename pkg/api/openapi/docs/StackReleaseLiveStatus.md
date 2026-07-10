# StackReleaseLiveStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Health** | Pointer to [**ReleaseHealth**](ReleaseHealth.md) |  | [optional] 
**Resources** | Pointer to [**map[string]StackResourceStatus**](StackResourceStatus.md) |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**TargetRevision** | Pointer to **string** |  | [optional] 
**ObservedRevision** | Pointer to **string** |  | [optional] 

## Methods

### NewStackReleaseLiveStatus

`func NewStackReleaseLiveStatus() *StackReleaseLiveStatus`

NewStackReleaseLiveStatus instantiates a new StackReleaseLiveStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackReleaseLiveStatusWithDefaults

`func NewStackReleaseLiveStatusWithDefaults() *StackReleaseLiveStatus`

NewStackReleaseLiveStatusWithDefaults instantiates a new StackReleaseLiveStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHealth

`func (o *StackReleaseLiveStatus) GetHealth() ReleaseHealth`

GetHealth returns the Health field if non-nil, zero value otherwise.

### GetHealthOk

`func (o *StackReleaseLiveStatus) GetHealthOk() (*ReleaseHealth, bool)`

GetHealthOk returns a tuple with the Health field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHealth

`func (o *StackReleaseLiveStatus) SetHealth(v ReleaseHealth)`

SetHealth sets Health field to given value.

### HasHealth

`func (o *StackReleaseLiveStatus) HasHealth() bool`

HasHealth returns a boolean if a field has been set.

### GetResources

`func (o *StackReleaseLiveStatus) GetResources() map[string]StackResourceStatus`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *StackReleaseLiveStatus) GetResourcesOk() (*map[string]StackResourceStatus, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *StackReleaseLiveStatus) SetResources(v map[string]StackResourceStatus)`

SetResources sets Resources field to given value.

### HasResources

`func (o *StackReleaseLiveStatus) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetConditions

`func (o *StackReleaseLiveStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *StackReleaseLiveStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *StackReleaseLiveStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *StackReleaseLiveStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetTargetRevision

`func (o *StackReleaseLiveStatus) GetTargetRevision() string`

GetTargetRevision returns the TargetRevision field if non-nil, zero value otherwise.

### GetTargetRevisionOk

`func (o *StackReleaseLiveStatus) GetTargetRevisionOk() (*string, bool)`

GetTargetRevisionOk returns a tuple with the TargetRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTargetRevision

`func (o *StackReleaseLiveStatus) SetTargetRevision(v string)`

SetTargetRevision sets TargetRevision field to given value.

### HasTargetRevision

`func (o *StackReleaseLiveStatus) HasTargetRevision() bool`

HasTargetRevision returns a boolean if a field has been set.

### GetObservedRevision

`func (o *StackReleaseLiveStatus) GetObservedRevision() string`

GetObservedRevision returns the ObservedRevision field if non-nil, zero value otherwise.

### GetObservedRevisionOk

`func (o *StackReleaseLiveStatus) GetObservedRevisionOk() (*string, bool)`

GetObservedRevisionOk returns a tuple with the ObservedRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedRevision

`func (o *StackReleaseLiveStatus) SetObservedRevision(v string)`

SetObservedRevision sets ObservedRevision field to given value.

### HasObservedRevision

`func (o *StackReleaseLiveStatus) HasObservedRevision() bool`

HasObservedRevision returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


