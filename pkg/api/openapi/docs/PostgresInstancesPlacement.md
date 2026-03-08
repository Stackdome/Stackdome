# PostgresInstancesPlacement

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TopologyKey** | Pointer to **string** | Topology key for instance spreading | [optional] [default to "kubernetes.io/hostname"]
**Policy** | Pointer to **string** | Instance placement policy | [optional] [default to "preferred"]
**NodeSelector** | Pointer to **map[string]string** | Node selector for instance placement | [optional] 
**Tolerations** | Pointer to [**[]PostgresInstancesPlacementTolerationsInner**](PostgresInstancesPlacementTolerationsInner.md) |  | [optional] 

## Methods

### NewPostgresInstancesPlacement

`func NewPostgresInstancesPlacement() *PostgresInstancesPlacement`

NewPostgresInstancesPlacement instantiates a new PostgresInstancesPlacement object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresInstancesPlacementWithDefaults

`func NewPostgresInstancesPlacementWithDefaults() *PostgresInstancesPlacement`

NewPostgresInstancesPlacementWithDefaults instantiates a new PostgresInstancesPlacement object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTopologyKey

`func (o *PostgresInstancesPlacement) GetTopologyKey() string`

GetTopologyKey returns the TopologyKey field if non-nil, zero value otherwise.

### GetTopologyKeyOk

`func (o *PostgresInstancesPlacement) GetTopologyKeyOk() (*string, bool)`

GetTopologyKeyOk returns a tuple with the TopologyKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTopologyKey

`func (o *PostgresInstancesPlacement) SetTopologyKey(v string)`

SetTopologyKey sets TopologyKey field to given value.

### HasTopologyKey

`func (o *PostgresInstancesPlacement) HasTopologyKey() bool`

HasTopologyKey returns a boolean if a field has been set.

### GetPolicy

`func (o *PostgresInstancesPlacement) GetPolicy() string`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *PostgresInstancesPlacement) GetPolicyOk() (*string, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *PostgresInstancesPlacement) SetPolicy(v string)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *PostgresInstancesPlacement) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### GetNodeSelector

`func (o *PostgresInstancesPlacement) GetNodeSelector() map[string]string`

GetNodeSelector returns the NodeSelector field if non-nil, zero value otherwise.

### GetNodeSelectorOk

`func (o *PostgresInstancesPlacement) GetNodeSelectorOk() (*map[string]string, bool)`

GetNodeSelectorOk returns a tuple with the NodeSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeSelector

`func (o *PostgresInstancesPlacement) SetNodeSelector(v map[string]string)`

SetNodeSelector sets NodeSelector field to given value.

### HasNodeSelector

`func (o *PostgresInstancesPlacement) HasNodeSelector() bool`

HasNodeSelector returns a boolean if a field has been set.

### GetTolerations

`func (o *PostgresInstancesPlacement) GetTolerations() []PostgresInstancesPlacementTolerationsInner`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *PostgresInstancesPlacement) GetTolerationsOk() (*[]PostgresInstancesPlacementTolerationsInner, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *PostgresInstancesPlacement) SetTolerations(v []PostgresInstancesPlacementTolerationsInner)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *PostgresInstancesPlacement) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


