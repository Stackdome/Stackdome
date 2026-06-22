# StackReleaseSnapshot

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Stack** | Pointer to [**StackReleaseSnapshotStack**](StackReleaseSnapshotStack.md) |  | [optional] 
**Resources** | Pointer to [**[]StackResource**](StackResource.md) |  | [optional] 
**Volumes** | Pointer to [**[]Volume**](Volume.md) |  | [optional] 
**Connections** | Pointer to [**[]StackConnection**](StackConnection.md) |  | [optional] 
**CapturedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewStackReleaseSnapshot

`func NewStackReleaseSnapshot() *StackReleaseSnapshot`

NewStackReleaseSnapshot instantiates a new StackReleaseSnapshot object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackReleaseSnapshotWithDefaults

`func NewStackReleaseSnapshotWithDefaults() *StackReleaseSnapshot`

NewStackReleaseSnapshotWithDefaults instantiates a new StackReleaseSnapshot object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStack

`func (o *StackReleaseSnapshot) GetStack() StackReleaseSnapshotStack`

GetStack returns the Stack field if non-nil, zero value otherwise.

### GetStackOk

`func (o *StackReleaseSnapshot) GetStackOk() (*StackReleaseSnapshotStack, bool)`

GetStackOk returns a tuple with the Stack field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStack

`func (o *StackReleaseSnapshot) SetStack(v StackReleaseSnapshotStack)`

SetStack sets Stack field to given value.

### HasStack

`func (o *StackReleaseSnapshot) HasStack() bool`

HasStack returns a boolean if a field has been set.

### GetResources

`func (o *StackReleaseSnapshot) GetResources() []StackResource`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *StackReleaseSnapshot) GetResourcesOk() (*[]StackResource, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *StackReleaseSnapshot) SetResources(v []StackResource)`

SetResources sets Resources field to given value.

### HasResources

`func (o *StackReleaseSnapshot) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetVolumes

`func (o *StackReleaseSnapshot) GetVolumes() []Volume`

GetVolumes returns the Volumes field if non-nil, zero value otherwise.

### GetVolumesOk

`func (o *StackReleaseSnapshot) GetVolumesOk() (*[]Volume, bool)`

GetVolumesOk returns a tuple with the Volumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumes

`func (o *StackReleaseSnapshot) SetVolumes(v []Volume)`

SetVolumes sets Volumes field to given value.

### HasVolumes

`func (o *StackReleaseSnapshot) HasVolumes() bool`

HasVolumes returns a boolean if a field has been set.

### GetConnections

`func (o *StackReleaseSnapshot) GetConnections() []StackConnection`

GetConnections returns the Connections field if non-nil, zero value otherwise.

### GetConnectionsOk

`func (o *StackReleaseSnapshot) GetConnectionsOk() (*[]StackConnection, bool)`

GetConnectionsOk returns a tuple with the Connections field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnections

`func (o *StackReleaseSnapshot) SetConnections(v []StackConnection)`

SetConnections sets Connections field to given value.

### HasConnections

`func (o *StackReleaseSnapshot) HasConnections() bool`

HasConnections returns a boolean if a field has been set.

### GetCapturedAt

`func (o *StackReleaseSnapshot) GetCapturedAt() time.Time`

GetCapturedAt returns the CapturedAt field if non-nil, zero value otherwise.

### GetCapturedAtOk

`func (o *StackReleaseSnapshot) GetCapturedAtOk() (*time.Time, bool)`

GetCapturedAtOk returns a tuple with the CapturedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCapturedAt

`func (o *StackReleaseSnapshot) SetCapturedAt(v time.Time)`

SetCapturedAt sets CapturedAt field to given value.

### HasCapturedAt

`func (o *StackReleaseSnapshot) HasCapturedAt() bool`

HasCapturedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


