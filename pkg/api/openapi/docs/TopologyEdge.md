# TopologyEdge

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Kind** | **string** | Edge kind. Explicit connections reuse connection kinds; derived edges use depends_on. | 
**Source** | [**TopologyNodeRef**](TopologyNodeRef.md) |  | 
**Target** | [**TopologyNodeRef**](TopologyNodeRef.md) |  | 
**Mappings** | Pointer to [**[]ConnectionMapping**](ConnectionMapping.md) |  | [optional] 
**Config** | Pointer to [**StackConnectionConfig**](StackConnectionConfig.md) |  | [optional] 
**SourceOfTruth** | **string** | Whether the edge came from an explicit connection or a derived relationship such as depends_on. | 

## Methods

### NewTopologyEdge

`func NewTopologyEdge(kind string, source TopologyNodeRef, target TopologyNodeRef, sourceOfTruth string, ) *TopologyEdge`

NewTopologyEdge instantiates a new TopologyEdge object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTopologyEdgeWithDefaults

`func NewTopologyEdgeWithDefaults() *TopologyEdge`

NewTopologyEdgeWithDefaults instantiates a new TopologyEdge object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *TopologyEdge) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TopologyEdge) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TopologyEdge) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *TopologyEdge) HasId() bool`

HasId returns a boolean if a field has been set.

### GetKind

`func (o *TopologyEdge) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *TopologyEdge) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *TopologyEdge) SetKind(v string)`

SetKind sets Kind field to given value.


### GetSource

`func (o *TopologyEdge) GetSource() TopologyNodeRef`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *TopologyEdge) GetSourceOk() (*TopologyNodeRef, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *TopologyEdge) SetSource(v TopologyNodeRef)`

SetSource sets Source field to given value.


### GetTarget

`func (o *TopologyEdge) GetTarget() TopologyNodeRef`

GetTarget returns the Target field if non-nil, zero value otherwise.

### GetTargetOk

`func (o *TopologyEdge) GetTargetOk() (*TopologyNodeRef, bool)`

GetTargetOk returns a tuple with the Target field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTarget

`func (o *TopologyEdge) SetTarget(v TopologyNodeRef)`

SetTarget sets Target field to given value.


### GetMappings

`func (o *TopologyEdge) GetMappings() []ConnectionMapping`

GetMappings returns the Mappings field if non-nil, zero value otherwise.

### GetMappingsOk

`func (o *TopologyEdge) GetMappingsOk() (*[]ConnectionMapping, bool)`

GetMappingsOk returns a tuple with the Mappings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMappings

`func (o *TopologyEdge) SetMappings(v []ConnectionMapping)`

SetMappings sets Mappings field to given value.

### HasMappings

`func (o *TopologyEdge) HasMappings() bool`

HasMappings returns a boolean if a field has been set.

### GetConfig

`func (o *TopologyEdge) GetConfig() StackConnectionConfig`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *TopologyEdge) GetConfigOk() (*StackConnectionConfig, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *TopologyEdge) SetConfig(v StackConnectionConfig)`

SetConfig sets Config field to given value.

### HasConfig

`func (o *TopologyEdge) HasConfig() bool`

HasConfig returns a boolean if a field has been set.

### GetSourceOfTruth

`func (o *TopologyEdge) GetSourceOfTruth() string`

GetSourceOfTruth returns the SourceOfTruth field if non-nil, zero value otherwise.

### GetSourceOfTruthOk

`func (o *TopologyEdge) GetSourceOfTruthOk() (*string, bool)`

GetSourceOfTruthOk returns a tuple with the SourceOfTruth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfTruth

`func (o *TopologyEdge) SetSourceOfTruth(v string)`

SetSourceOfTruth sets SourceOfTruth field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


