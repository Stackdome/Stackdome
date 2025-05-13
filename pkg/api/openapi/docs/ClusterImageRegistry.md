# ClusterImageRegistry

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Name** | **string** |  | 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**ClusterId** | Pointer to **string** |  | [optional] [readonly] 
**Spec** | Pointer to [**ClusterImageRegistrySpec**](ClusterImageRegistrySpec.md) |  | [optional] 
**Status** | Pointer to [**ClusterImageRegistryStatus**](ClusterImageRegistryStatus.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewClusterImageRegistry

`func NewClusterImageRegistry(name string, ) *ClusterImageRegistry`

NewClusterImageRegistry instantiates a new ClusterImageRegistry object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClusterImageRegistryWithDefaults

`func NewClusterImageRegistryWithDefaults() *ClusterImageRegistry`

NewClusterImageRegistryWithDefaults instantiates a new ClusterImageRegistry object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ClusterImageRegistry) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ClusterImageRegistry) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ClusterImageRegistry) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ClusterImageRegistry) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *ClusterImageRegistry) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ClusterImageRegistry) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ClusterImageRegistry) SetName(v string)`

SetName sets Name field to given value.


### GetOrganisationId

`func (o *ClusterImageRegistry) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *ClusterImageRegistry) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *ClusterImageRegistry) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *ClusterImageRegistry) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetClusterId

`func (o *ClusterImageRegistry) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ClusterImageRegistry) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ClusterImageRegistry) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *ClusterImageRegistry) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetSpec

`func (o *ClusterImageRegistry) GetSpec() ClusterImageRegistrySpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *ClusterImageRegistry) GetSpecOk() (*ClusterImageRegistrySpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *ClusterImageRegistry) SetSpec(v ClusterImageRegistrySpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *ClusterImageRegistry) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetStatus

`func (o *ClusterImageRegistry) GetStatus() ClusterImageRegistryStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *ClusterImageRegistry) GetStatusOk() (*ClusterImageRegistryStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *ClusterImageRegistry) SetStatus(v ClusterImageRegistryStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *ClusterImageRegistry) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ClusterImageRegistry) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ClusterImageRegistry) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ClusterImageRegistry) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ClusterImageRegistry) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *ClusterImageRegistry) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *ClusterImageRegistry) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *ClusterImageRegistry) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *ClusterImageRegistry) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


