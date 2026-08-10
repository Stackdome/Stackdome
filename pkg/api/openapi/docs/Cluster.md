# Cluster

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**OrganisationId** | Pointer to **string** |  | [optional] 
**SharedCompute** | Pointer to **bool** |  | [optional] [readonly]
**Platform** | Pointer to **bool** | Deprecated alias for shared_compute. Both fields have the same value. | [optional] [readonly]
**ClusterUrl** | **string** |  | 
**ClusterCaData** | **string** |  | 
**ClusterSaToken** | **string** |  | 
**ClusterImageRegistry** | Pointer to [**ClusterImageRegistry**](ClusterImageRegistry.md) |  | [optional] 

## Methods

### NewCluster

`func NewCluster(name string, clusterUrl string, clusterCaData string, clusterSaToken string, ) *Cluster`

NewCluster instantiates a new Cluster object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterWithDefaults

`func NewClusterWithDefaults() *Cluster`

NewClusterWithDefaults instantiates a new Cluster object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Cluster) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Cluster) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Cluster) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *Cluster) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *Cluster) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Cluster) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Cluster) SetName(v string)`

SetName sets Name field to given value.


### GetOrganisationId

`func (o *Cluster) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *Cluster) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *Cluster) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *Cluster) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetSharedCompute

`func (o *Cluster) GetSharedCompute() bool`

GetSharedCompute returns the SharedCompute field if non-nil, zero value otherwise.

### GetSharedComputeOk

`func (o *Cluster) GetSharedComputeOk() (*bool, bool)`

GetSharedComputeOk returns a tuple with the SharedCompute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSharedCompute

`func (o *Cluster) SetSharedCompute(v bool)`

SetSharedCompute sets SharedCompute field to given value.

### HasSharedCompute

`func (o *Cluster) HasSharedCompute() bool`

HasSharedCompute returns a boolean if a field has been set.

### GetPlatform

`func (o *Cluster) GetPlatform() bool`

GetPlatform returns the Platform field if non-nil, zero value otherwise.

### GetPlatformOk

`func (o *Cluster) GetPlatformOk() (*bool, bool)`

GetPlatformOk returns a tuple with the Platform field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlatform

`func (o *Cluster) SetPlatform(v bool)`

SetPlatform sets Platform field to given value.

### HasPlatform

`func (o *Cluster) HasPlatform() bool`

HasPlatform returns a boolean if a field has been set.

### GetClusterUrl

`func (o *Cluster) GetClusterUrl() string`

GetClusterUrl returns the ClusterUrl field if non-nil, zero value otherwise.

### GetClusterUrlOk

`func (o *Cluster) GetClusterUrlOk() (*string, bool)`

GetClusterUrlOk returns a tuple with the ClusterUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUrl

`func (o *Cluster) SetClusterUrl(v string)`

SetClusterUrl sets ClusterUrl field to given value.


### GetClusterCaData

`func (o *Cluster) GetClusterCaData() string`

GetClusterCaData returns the ClusterCaData field if non-nil, zero value otherwise.

### GetClusterCaDataOk

`func (o *Cluster) GetClusterCaDataOk() (*string, bool)`

GetClusterCaDataOk returns a tuple with the ClusterCaData field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterCaData

`func (o *Cluster) SetClusterCaData(v string)`

SetClusterCaData sets ClusterCaData field to given value.


### GetClusterSaToken

`func (o *Cluster) GetClusterSaToken() string`

GetClusterSaToken returns the ClusterSaToken field if non-nil, zero value otherwise.

### GetClusterSaTokenOk

`func (o *Cluster) GetClusterSaTokenOk() (*string, bool)`

GetClusterSaTokenOk returns a tuple with the ClusterSaToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterSaToken

`func (o *Cluster) SetClusterSaToken(v string)`

SetClusterSaToken sets ClusterSaToken field to given value.


### GetClusterImageRegistry

`func (o *Cluster) GetClusterImageRegistry() ClusterImageRegistry`

GetClusterImageRegistry returns the ClusterImageRegistry field if non-nil, zero value otherwise.

### GetClusterImageRegistryOk

`func (o *Cluster) GetClusterImageRegistryOk() (*ClusterImageRegistry, bool)`

GetClusterImageRegistryOk returns a tuple with the ClusterImageRegistry field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterImageRegistry

`func (o *Cluster) SetClusterImageRegistry(v ClusterImageRegistry)`

SetClusterImageRegistry sets ClusterImageRegistry field to given value.

### HasClusterImageRegistry

`func (o *Cluster) HasClusterImageRegistry() bool`

HasClusterImageRegistry returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
