# ClusterImageRegistrySpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**BackendStorageSize** | Pointer to **string** |  | [optional] 
**BackendStorageClass** | Pointer to **string** |  | [optional] 

## Methods

### NewClusterImageRegistrySpec

`func NewClusterImageRegistrySpec() *ClusterImageRegistrySpec`

NewClusterImageRegistrySpec instantiates a new ClusterImageRegistrySpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterImageRegistrySpecWithDefaults

`func NewClusterImageRegistrySpecWithDefaults() *ClusterImageRegistrySpec`

NewClusterImageRegistrySpecWithDefaults instantiates a new ClusterImageRegistrySpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBackendStorageSize

`func (o *ClusterImageRegistrySpec) GetBackendStorageSize() string`

GetBackendStorageSize returns the BackendStorageSize field if non-nil, zero value otherwise.

### GetBackendStorageSizeOk

`func (o *ClusterImageRegistrySpec) GetBackendStorageSizeOk() (*string, bool)`

GetBackendStorageSizeOk returns a tuple with the BackendStorageSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackendStorageSize

`func (o *ClusterImageRegistrySpec) SetBackendStorageSize(v string)`

SetBackendStorageSize sets BackendStorageSize field to given value.

### HasBackendStorageSize

`func (o *ClusterImageRegistrySpec) HasBackendStorageSize() bool`

HasBackendStorageSize returns a boolean if a field has been set.

### GetBackendStorageClass

`func (o *ClusterImageRegistrySpec) GetBackendStorageClass() string`

GetBackendStorageClass returns the BackendStorageClass field if non-nil, zero value otherwise.

### GetBackendStorageClassOk

`func (o *ClusterImageRegistrySpec) GetBackendStorageClassOk() (*string, bool)`

GetBackendStorageClassOk returns a tuple with the BackendStorageClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackendStorageClass

`func (o *ClusterImageRegistrySpec) SetBackendStorageClass(v string)`

SetBackendStorageClass sets BackendStorageClass field to given value.

### HasBackendStorageClass

`func (o *ClusterImageRegistrySpec) HasBackendStorageClass() bool`

HasBackendStorageClass returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


