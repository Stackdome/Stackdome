# StackTopology

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Nodes** | [**[]TopologyNode**](TopologyNode.md) |  | 
**Edges** | [**[]TopologyEdge**](TopologyEdge.md) |  | 

## Methods

### NewStackTopology

`func NewStackTopology(nodes []TopologyNode, edges []TopologyEdge, ) *StackTopology`

NewStackTopology instantiates a new StackTopology object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackTopologyWithDefaults

`func NewStackTopologyWithDefaults() *StackTopology`

NewStackTopologyWithDefaults instantiates a new StackTopology object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodes

`func (o *StackTopology) GetNodes() []TopologyNode`

GetNodes returns the Nodes field if non-nil, zero value otherwise.

### GetNodesOk

`func (o *StackTopology) GetNodesOk() (*[]TopologyNode, bool)`

GetNodesOk returns a tuple with the Nodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodes

`func (o *StackTopology) SetNodes(v []TopologyNode)`

SetNodes sets Nodes field to given value.


### GetEdges

`func (o *StackTopology) GetEdges() []TopologyEdge`

GetEdges returns the Edges field if non-nil, zero value otherwise.

### GetEdgesOk

`func (o *StackTopology) GetEdgesOk() (*[]TopologyEdge, bool)`

GetEdgesOk returns a tuple with the Edges field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEdges

`func (o *StackTopology) SetEdges(v []TopologyEdge)`

SetEdges sets Edges field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


