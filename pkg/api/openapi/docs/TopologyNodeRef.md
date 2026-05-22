# TopologyNodeRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | **string** | The node category. | 
**Id** | Pointer to **string** | Stable ID for persisted resources such as addons or secrets. | [optional] 
**Name** | Pointer to **string** | Name-scoped reference for stack-local resources such as StackResources or Volumes. | [optional] 

## Methods

### NewTopologyNodeRef

`func NewTopologyNodeRef(type_ string, ) *TopologyNodeRef`

NewTopologyNodeRef instantiates a new TopologyNodeRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTopologyNodeRefWithDefaults

`func NewTopologyNodeRefWithDefaults() *TopologyNodeRef`

NewTopologyNodeRefWithDefaults instantiates a new TopologyNodeRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *TopologyNodeRef) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *TopologyNodeRef) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *TopologyNodeRef) SetType(v string)`

SetType sets Type field to given value.


### GetId

`func (o *TopologyNodeRef) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TopologyNodeRef) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TopologyNodeRef) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *TopologyNodeRef) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *TopologyNodeRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *TopologyNodeRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *TopologyNodeRef) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *TopologyNodeRef) HasName() bool`

HasName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


