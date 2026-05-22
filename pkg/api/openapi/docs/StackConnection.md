# StackConnection

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | Stable connection identifier. Generated when omitted. | [optional] 
**Kind** | **string** | The relationship type. &#x60;env&#x60; injects values into environment variables, &#x60;secret_mount&#x60; mounts secret values as files, &#x60;volume_mount&#x60; mounts a volume into a resource, and &#x60;build_artifact_source&#x60; seeds a volume from build output.  | 
**From** | [**TopologyNodeRef**](TopologyNodeRef.md) |  | 
**To** | [**TopologyNodeRef**](TopologyNodeRef.md) |  | 
**Mappings** | Pointer to [**[]ConnectionMapping**](ConnectionMapping.md) | Target/value mappings for kinds that move values, such as &#x60;env&#x60; and &#x60;secret_mount&#x60;.  | [optional] 
**Config** | Pointer to **map[string]interface{}** | Kind-specific connection configuration. Expected keys depend on &#x60;kind&#x60;. Examples: for &#x60;env&#x60; from &#x60;addon/postgres&#x60;, use fields such as &#x60;database&#x60; and &#x60;credential_scope&#x60;; for &#x60;volume_mount&#x60;, use &#x60;mount_path&#x60;, &#x60;sub_path&#x60;, and &#x60;read_only&#x60;; for &#x60;build_artifact_source&#x60;, use &#x60;source_path&#x60; and &#x60;destination_path&#x60;.  | [optional] 

## Methods

### NewStackConnection

`func NewStackConnection(kind string, from TopologyNodeRef, to TopologyNodeRef, ) *StackConnection`

NewStackConnection instantiates a new StackConnection object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackConnectionWithDefaults

`func NewStackConnectionWithDefaults() *StackConnection`

NewStackConnectionWithDefaults instantiates a new StackConnection object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *StackConnection) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *StackConnection) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *StackConnection) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *StackConnection) HasId() bool`

HasId returns a boolean if a field has been set.

### GetKind

`func (o *StackConnection) GetKind() string`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *StackConnection) GetKindOk() (*string, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *StackConnection) SetKind(v string)`

SetKind sets Kind field to given value.


### GetFrom

`func (o *StackConnection) GetFrom() TopologyNodeRef`

GetFrom returns the From field if non-nil, zero value otherwise.

### GetFromOk

`func (o *StackConnection) GetFromOk() (*TopologyNodeRef, bool)`

GetFromOk returns a tuple with the From field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFrom

`func (o *StackConnection) SetFrom(v TopologyNodeRef)`

SetFrom sets From field to given value.


### GetTo

`func (o *StackConnection) GetTo() TopologyNodeRef`

GetTo returns the To field if non-nil, zero value otherwise.

### GetToOk

`func (o *StackConnection) GetToOk() (*TopologyNodeRef, bool)`

GetToOk returns a tuple with the To field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTo

`func (o *StackConnection) SetTo(v TopologyNodeRef)`

SetTo sets To field to given value.


### GetMappings

`func (o *StackConnection) GetMappings() []ConnectionMapping`

GetMappings returns the Mappings field if non-nil, zero value otherwise.

### GetMappingsOk

`func (o *StackConnection) GetMappingsOk() (*[]ConnectionMapping, bool)`

GetMappingsOk returns a tuple with the Mappings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMappings

`func (o *StackConnection) SetMappings(v []ConnectionMapping)`

SetMappings sets Mappings field to given value.

### HasMappings

`func (o *StackConnection) HasMappings() bool`

HasMappings returns a boolean if a field has been set.

### GetConfig

`func (o *StackConnection) GetConfig() map[string]interface{}`

GetConfig returns the Config field if non-nil, zero value otherwise.

### GetConfigOk

`func (o *StackConnection) GetConfigOk() (*map[string]interface{}, bool)`

GetConfigOk returns a tuple with the Config field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfig

`func (o *StackConnection) SetConfig(v map[string]interface{})`

SetConfig sets Config field to given value.

### HasConfig

`func (o *StackConnection) HasConfig() bool`

HasConfig returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


