# TopologyNode

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Ref** | [**TopologyNodeRef**](TopologyNodeRef.md) |  | 
**Label** | **string** |  | 
**Outputs** | Pointer to [**[]OutputDescriptor**](OutputDescriptor.md) |  | [optional] 
**State** | Pointer to **string** | Optional runtime state for nodes that have status, such as stack resources or addons. | [optional] 

## Methods

### NewTopologyNode

`func NewTopologyNode(ref TopologyNodeRef, label string, ) *TopologyNode`

NewTopologyNode instantiates a new TopologyNode object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTopologyNodeWithDefaults

`func NewTopologyNodeWithDefaults() *TopologyNode`

NewTopologyNodeWithDefaults instantiates a new TopologyNode object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRef

`func (o *TopologyNode) GetRef() TopologyNodeRef`

GetRef returns the Ref field if non-nil, zero value otherwise.

### GetRefOk

`func (o *TopologyNode) GetRefOk() (*TopologyNodeRef, bool)`

GetRefOk returns a tuple with the Ref field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRef

`func (o *TopologyNode) SetRef(v TopologyNodeRef)`

SetRef sets Ref field to given value.


### GetLabel

`func (o *TopologyNode) GetLabel() string`

GetLabel returns the Label field if non-nil, zero value otherwise.

### GetLabelOk

`func (o *TopologyNode) GetLabelOk() (*string, bool)`

GetLabelOk returns a tuple with the Label field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabel

`func (o *TopologyNode) SetLabel(v string)`

SetLabel sets Label field to given value.


### GetOutputs

`func (o *TopologyNode) GetOutputs() []OutputDescriptor`

GetOutputs returns the Outputs field if non-nil, zero value otherwise.

### GetOutputsOk

`func (o *TopologyNode) GetOutputsOk() (*[]OutputDescriptor, bool)`

GetOutputsOk returns a tuple with the Outputs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutputs

`func (o *TopologyNode) SetOutputs(v []OutputDescriptor)`

SetOutputs sets Outputs field to given value.

### HasOutputs

`func (o *TopologyNode) HasOutputs() bool`

HasOutputs returns a boolean if a field has been set.

### GetState

`func (o *TopologyNode) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *TopologyNode) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *TopologyNode) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *TopologyNode) HasState() bool`

HasState returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


