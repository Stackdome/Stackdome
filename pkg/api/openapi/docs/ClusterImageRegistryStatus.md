# ClusterImageRegistryStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to [**ClusterImageRegistryState**](ClusterImageRegistryState.md) |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 

## Methods

### NewClusterImageRegistryStatus

`func NewClusterImageRegistryStatus() *ClusterImageRegistryStatus`

NewClusterImageRegistryStatus instantiates a new ClusterImageRegistryStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterImageRegistryStatusWithDefaults

`func NewClusterImageRegistryStatusWithDefaults() *ClusterImageRegistryStatus`

NewClusterImageRegistryStatusWithDefaults instantiates a new ClusterImageRegistryStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *ClusterImageRegistryStatus) GetState() ClusterImageRegistryState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ClusterImageRegistryStatus) GetStateOk() (*ClusterImageRegistryState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ClusterImageRegistryStatus) SetState(v ClusterImageRegistryState)`

SetState sets State field to given value.

### HasState

`func (o *ClusterImageRegistryStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetConditions

`func (o *ClusterImageRegistryStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ClusterImageRegistryStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ClusterImageRegistryStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ClusterImageRegistryStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


