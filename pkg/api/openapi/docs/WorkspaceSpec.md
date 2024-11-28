# WorkspaceSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Resources** | [**[]WorkspaceResource**](WorkspaceResource.md) |  | 

## Methods

### NewWorkspaceSpec

`func NewWorkspaceSpec(resources []WorkspaceResource, ) *WorkspaceSpec`

NewWorkspaceSpec instantiates a new WorkspaceSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceSpecWithDefaults

`func NewWorkspaceSpecWithDefaults() *WorkspaceSpec`

NewWorkspaceSpecWithDefaults instantiates a new WorkspaceSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetResources

`func (o *WorkspaceSpec) GetResources() []WorkspaceResource`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *WorkspaceSpec) GetResourcesOk() (*[]WorkspaceResource, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *WorkspaceSpec) SetResources(v []WorkspaceResource)`

SetResources sets Resources field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


