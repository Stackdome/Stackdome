# ResourceMetrics

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AssignedNodes** | Pointer to **[]string** |  | [optional] 
**CpuUsage** | Pointer to **string** | CPU usage in millicores | [optional] 
**MemoryUsage** | Pointer to **string** | Memory usage in bytes | [optional] 
**NodeCapacities** | Pointer to [**[]ResourceMetricsNodeCapacitiesInner**](ResourceMetricsNodeCapacitiesInner.md) |  | [optional] 
**Timestamp** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewResourceMetrics

`func NewResourceMetrics() *ResourceMetrics`

NewResourceMetrics instantiates a new ResourceMetrics object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceMetricsWithDefaults

`func NewResourceMetricsWithDefaults() *ResourceMetrics`

NewResourceMetricsWithDefaults instantiates a new ResourceMetrics object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAssignedNodes

`func (o *ResourceMetrics) GetAssignedNodes() []string`

GetAssignedNodes returns the AssignedNodes field if non-nil, zero value otherwise.

### GetAssignedNodesOk

`func (o *ResourceMetrics) GetAssignedNodesOk() (*[]string, bool)`

GetAssignedNodesOk returns a tuple with the AssignedNodes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssignedNodes

`func (o *ResourceMetrics) SetAssignedNodes(v []string)`

SetAssignedNodes sets AssignedNodes field to given value.

### HasAssignedNodes

`func (o *ResourceMetrics) HasAssignedNodes() bool`

HasAssignedNodes returns a boolean if a field has been set.

### GetCpuUsage

`func (o *ResourceMetrics) GetCpuUsage() string`

GetCpuUsage returns the CpuUsage field if non-nil, zero value otherwise.

### GetCpuUsageOk

`func (o *ResourceMetrics) GetCpuUsageOk() (*string, bool)`

GetCpuUsageOk returns a tuple with the CpuUsage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpuUsage

`func (o *ResourceMetrics) SetCpuUsage(v string)`

SetCpuUsage sets CpuUsage field to given value.

### HasCpuUsage

`func (o *ResourceMetrics) HasCpuUsage() bool`

HasCpuUsage returns a boolean if a field has been set.

### GetMemoryUsage

`func (o *ResourceMetrics) GetMemoryUsage() string`

GetMemoryUsage returns the MemoryUsage field if non-nil, zero value otherwise.

### GetMemoryUsageOk

`func (o *ResourceMetrics) GetMemoryUsageOk() (*string, bool)`

GetMemoryUsageOk returns a tuple with the MemoryUsage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMemoryUsage

`func (o *ResourceMetrics) SetMemoryUsage(v string)`

SetMemoryUsage sets MemoryUsage field to given value.

### HasMemoryUsage

`func (o *ResourceMetrics) HasMemoryUsage() bool`

HasMemoryUsage returns a boolean if a field has been set.

### GetNodeCapacities

`func (o *ResourceMetrics) GetNodeCapacities() []ResourceMetricsNodeCapacitiesInner`

GetNodeCapacities returns the NodeCapacities field if non-nil, zero value otherwise.

### GetNodeCapacitiesOk

`func (o *ResourceMetrics) GetNodeCapacitiesOk() (*[]ResourceMetricsNodeCapacitiesInner, bool)`

GetNodeCapacitiesOk returns a tuple with the NodeCapacities field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeCapacities

`func (o *ResourceMetrics) SetNodeCapacities(v []ResourceMetricsNodeCapacitiesInner)`

SetNodeCapacities sets NodeCapacities field to given value.

### HasNodeCapacities

`func (o *ResourceMetrics) HasNodeCapacities() bool`

HasNodeCapacities returns a boolean if a field has been set.

### GetTimestamp

`func (o *ResourceMetrics) GetTimestamp() time.Time`

GetTimestamp returns the Timestamp field if non-nil, zero value otherwise.

### GetTimestampOk

`func (o *ResourceMetrics) GetTimestampOk() (*time.Time, bool)`

GetTimestampOk returns a tuple with the Timestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTimestamp

`func (o *ResourceMetrics) SetTimestamp(v time.Time)`

SetTimestamp sets Timestamp field to given value.

### HasTimestamp

`func (o *ResourceMetrics) HasTimestamp() bool`

HasTimestamp returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


