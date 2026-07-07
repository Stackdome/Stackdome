# StackSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StackResources** | Pointer to [**[]StackResource**](StackResource.md) |  | [optional] 
**Volumes** | Pointer to [**[]Volume**](Volume.md) |  | [optional] 
**Connections** | Pointer to [**[]StackConnection**](StackConnection.md) |  | [optional] 

## Methods

### NewStackSpec

`func NewStackSpec() *StackSpec`

NewStackSpec instantiates a new StackSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackSpecWithDefaults

`func NewStackSpecWithDefaults() *StackSpec`

NewStackSpecWithDefaults instantiates a new StackSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStackResources

`func (o *StackSpec) GetStackResources() []StackResource`

GetStackResources returns the StackResources field if non-nil, zero value otherwise.

### GetStackResourcesOk

`func (o *StackSpec) GetStackResourcesOk() (*[]StackResource, bool)`

GetStackResourcesOk returns a tuple with the StackResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackResources

`func (o *StackSpec) SetStackResources(v []StackResource)`

SetStackResources sets StackResources field to given value.

### HasStackResources

`func (o *StackSpec) HasStackResources() bool`

HasStackResources returns a boolean if a field has been set.

### GetVolumes

`func (o *StackSpec) GetVolumes() []Volume`

GetVolumes returns the Volumes field if non-nil, zero value otherwise.

### GetVolumesOk

`func (o *StackSpec) GetVolumesOk() (*[]Volume, bool)`

GetVolumesOk returns a tuple with the Volumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumes

`func (o *StackSpec) SetVolumes(v []Volume)`

SetVolumes sets Volumes field to given value.

### HasVolumes

`func (o *StackSpec) HasVolumes() bool`

HasVolumes returns a boolean if a field has been set.

### GetConnections

`func (o *StackSpec) GetConnections() []StackConnection`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *StackSpec) GetConnectionsOk() (*[]StackConnection, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *StackSpec) SetConnections(v []StackConnection)`

SetConnections sets Connections field to given value.

### HasConnections

`func (o *StackSpec) HasConnections() bool`

HasConnections returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


