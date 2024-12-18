# WorkspaceResourceBuild

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Namespace** | Pointer to **string** |  | [optional] 
**WorkspaceId** | Pointer to **string** |  | [optional] 
**WorkspaceResourceId** | Pointer to **string** |  | [optional] 
**WorkspaceResourceName** | Pointer to **string** |  | [optional] 
**SourceHash** | Pointer to **string** |  | [optional] 
**ImageRegistry** | Pointer to **string** |  | [optional] 
**Status** | Pointer to [**ResourceBuildStatus**](ResourceBuildStatus.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewWorkspaceResourceBuild

`func NewWorkspaceResourceBuild() *WorkspaceResourceBuild`

NewWorkspaceResourceBuild instantiates a new WorkspaceResourceBuild object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceResourceBuildWithDefaults

`func NewWorkspaceResourceBuildWithDefaults() *WorkspaceResourceBuild`

NewWorkspaceResourceBuildWithDefaults instantiates a new WorkspaceResourceBuild object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkspaceResourceBuild) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkspaceResourceBuild) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkspaceResourceBuild) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *WorkspaceResourceBuild) HasId() bool`

HasId returns a boolean if a field has been set.

### GetNamespace

`func (o *WorkspaceResourceBuild) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *WorkspaceResourceBuild) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *WorkspaceResourceBuild) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *WorkspaceResourceBuild) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetWorkspaceId

`func (o *WorkspaceResourceBuild) GetWorkspaceId() string`

GetWorkspaceId returns the WorkspaceId field if non-nil, zero value otherwise.

### GetWorkspaceIdOk

`func (o *WorkspaceResourceBuild) GetWorkspaceIdOk() (*string, bool)`

GetWorkspaceIdOk returns a tuple with the WorkspaceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceId

`func (o *WorkspaceResourceBuild) SetWorkspaceId(v string)`

SetWorkspaceId sets WorkspaceId field to given value.

### HasWorkspaceId

`func (o *WorkspaceResourceBuild) HasWorkspaceId() bool`

HasWorkspaceId returns a boolean if a field has been set.

### GetWorkspaceResourceId

`func (o *WorkspaceResourceBuild) GetWorkspaceResourceId() string`

GetWorkspaceResourceId returns the WorkspaceResourceId field if non-nil, zero value otherwise.

### GetWorkspaceResourceIdOk

`func (o *WorkspaceResourceBuild) GetWorkspaceResourceIdOk() (*string, bool)`

GetWorkspaceResourceIdOk returns a tuple with the WorkspaceResourceId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceResourceId

`func (o *WorkspaceResourceBuild) SetWorkspaceResourceId(v string)`

SetWorkspaceResourceId sets WorkspaceResourceId field to given value.

### HasWorkspaceResourceId

`func (o *WorkspaceResourceBuild) HasWorkspaceResourceId() bool`

HasWorkspaceResourceId returns a boolean if a field has been set.

### GetWorkspaceResourceName

`func (o *WorkspaceResourceBuild) GetWorkspaceResourceName() string`

GetWorkspaceResourceName returns the WorkspaceResourceName field if non-nil, zero value otherwise.

### GetWorkspaceResourceNameOk

`func (o *WorkspaceResourceBuild) GetWorkspaceResourceNameOk() (*string, bool)`

GetWorkspaceResourceNameOk returns a tuple with the WorkspaceResourceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceResourceName

`func (o *WorkspaceResourceBuild) SetWorkspaceResourceName(v string)`

SetWorkspaceResourceName sets WorkspaceResourceName field to given value.

### HasWorkspaceResourceName

`func (o *WorkspaceResourceBuild) HasWorkspaceResourceName() bool`

HasWorkspaceResourceName returns a boolean if a field has been set.

### GetSourceHash

`func (o *WorkspaceResourceBuild) GetSourceHash() string`

GetSourceHash returns the SourceHash field if non-nil, zero value otherwise.

### GetSourceHashOk

`func (o *WorkspaceResourceBuild) GetSourceHashOk() (*string, bool)`

GetSourceHashOk returns a tuple with the SourceHash field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceHash

`func (o *WorkspaceResourceBuild) SetSourceHash(v string)`

SetSourceHash sets SourceHash field to given value.

### HasSourceHash

`func (o *WorkspaceResourceBuild) HasSourceHash() bool`

HasSourceHash returns a boolean if a field has been set.

### GetImageRegistry

`func (o *WorkspaceResourceBuild) GetImageRegistry() string`

GetImageRegistry returns the ImageRegistry field if non-nil, zero value otherwise.

### GetImageRegistryOk

`func (o *WorkspaceResourceBuild) GetImageRegistryOk() (*string, bool)`

GetImageRegistryOk returns a tuple with the ImageRegistry field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageRegistry

`func (o *WorkspaceResourceBuild) SetImageRegistry(v string)`

SetImageRegistry sets ImageRegistry field to given value.

### HasImageRegistry

`func (o *WorkspaceResourceBuild) HasImageRegistry() bool`

HasImageRegistry returns a boolean if a field has been set.

### GetStatus

`func (o *WorkspaceResourceBuild) GetStatus() ResourceBuildStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *WorkspaceResourceBuild) GetStatusOk() (*ResourceBuildStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *WorkspaceResourceBuild) SetStatus(v ResourceBuildStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *WorkspaceResourceBuild) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetCreatedAt

`func (o *WorkspaceResourceBuild) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkspaceResourceBuild) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkspaceResourceBuild) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkspaceResourceBuild) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkspaceResourceBuild) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkspaceResourceBuild) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkspaceResourceBuild) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkspaceResourceBuild) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


