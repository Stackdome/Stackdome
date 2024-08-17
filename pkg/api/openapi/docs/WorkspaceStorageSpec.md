# WorkspaceStorageSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkspaceName** | **string** |  | 
**Volumes** | Pointer to [**[]Volume**](Volume.md) |  | [optional] 

## Methods

### NewWorkspaceStorageSpec

`func NewWorkspaceStorageSpec(workspaceName string, ) *WorkspaceStorageSpec`

NewWorkspaceStorageSpec instantiates a new WorkspaceStorageSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceStorageSpecWithDefaults

`func NewWorkspaceStorageSpecWithDefaults() *WorkspaceStorageSpec`

NewWorkspaceStorageSpecWithDefaults instantiates a new WorkspaceStorageSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkspaceName

`func (o *WorkspaceStorageSpec) GetWorkspaceName() string`

GetWorkspaceName returns the WorkspaceName field if non-nil, zero value otherwise.

### GetWorkspaceNameOk

`func (o *WorkspaceStorageSpec) GetWorkspaceNameOk() (*string, bool)`

GetWorkspaceNameOk returns a tuple with the WorkspaceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceName

`func (o *WorkspaceStorageSpec) SetWorkspaceName(v string)`

SetWorkspaceName sets WorkspaceName field to given value.


### GetVolumes

`func (o *WorkspaceStorageSpec) GetVolumes() []Volume`

GetVolumes returns the Volumes field if non-nil, zero value otherwise.

### GetVolumesOk

`func (o *WorkspaceStorageSpec) GetVolumesOk() (*[]Volume, bool)`

GetVolumesOk returns a tuple with the Volumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumes

`func (o *WorkspaceStorageSpec) SetVolumes(v []Volume)`

SetVolumes sets Volumes field to given value.

### HasVolumes

`func (o *WorkspaceStorageSpec) HasVolumes() bool`

HasVolumes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


