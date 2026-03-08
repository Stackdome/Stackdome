# PostgresResources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Cpu** | Pointer to [**PostgresResourcesCpu**](PostgresResourcesCpu.md) |  | [optional] 
**Memory** | Pointer to [**PostgresResourcesMemory**](PostgresResourcesMemory.md) |  | [optional] 

## Methods

### NewPostgresResources

`func NewPostgresResources() *PostgresResources`

NewPostgresResources instantiates a new PostgresResources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresResourcesWithDefaults

`func NewPostgresResourcesWithDefaults() *PostgresResources`

NewPostgresResourcesWithDefaults instantiates a new PostgresResources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCpu

`func (o *PostgresResources) GetCpu() PostgresResourcesCpu`

GetCpu returns the Cpu field if non-nil, zero value otherwise.

### GetCpuOk

`func (o *PostgresResources) GetCpuOk() (*PostgresResourcesCpu, bool)`

GetCpuOk returns a tuple with the Cpu field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpu

`func (o *PostgresResources) SetCpu(v PostgresResourcesCpu)`

SetCpu sets Cpu field to given value.

### HasCpu

`func (o *PostgresResources) HasCpu() bool`

HasCpu returns a boolean if a field has been set.

### GetMemory

`func (o *PostgresResources) GetMemory() PostgresResourcesMemory`

GetMemory returns the Memory field if non-nil, zero value otherwise.

### GetMemoryOk

`func (o *PostgresResources) GetMemoryOk() (*PostgresResourcesMemory, bool)`

GetMemoryOk returns a tuple with the Memory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemory

`func (o *PostgresResources) SetMemory(v PostgresResourcesMemory)`

SetMemory sets Memory field to given value.

### HasMemory

`func (o *PostgresResources) HasMemory() bool`

HasMemory returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


