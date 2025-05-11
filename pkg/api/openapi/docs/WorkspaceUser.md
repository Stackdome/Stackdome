# WorkspaceUser

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**UserId** | Pointer to **string** |  | [optional] 
**OrgId** | Pointer to **string** |  | [optional] 
**Workspaces** | **[]string** |  | 
**Version** | Pointer to **int32** |  | [optional] [readonly] 
**Status** | Pointer to [**WorkspaceUserStatus**](WorkspaceUserStatus.md) |  | [optional] 
**State** | Pointer to [**WorkspaceUserState**](WorkspaceUserState.md) |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewWorkspaceUser

`func NewWorkspaceUser(workspaces []string, ) *WorkspaceUser`

NewWorkspaceUser instantiates a new WorkspaceUser object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceUserWithDefaults

`func NewWorkspaceUserWithDefaults() *WorkspaceUser`

NewWorkspaceUserWithDefaults instantiates a new WorkspaceUser object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkspaceUser) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkspaceUser) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkspaceUser) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *WorkspaceUser) HasId() bool`

HasId returns a boolean if a field has been set.

### GetUserId

`func (o *WorkspaceUser) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *WorkspaceUser) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *WorkspaceUser) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *WorkspaceUser) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetOrgId

`func (o *WorkspaceUser) GetOrgId() string`

GetOrgId returns the OrgId field if non-nil, zero value otherwise.

### GetOrgIdOk

`func (o *WorkspaceUser) GetOrgIdOk() (*string, bool)`

GetOrgIdOk returns a tuple with the OrgId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgId

`func (o *WorkspaceUser) SetOrgId(v string)`

SetOrgId sets OrgId field to given value.

### HasOrgId

`func (o *WorkspaceUser) HasOrgId() bool`

HasOrgId returns a boolean if a field has been set.

### GetWorkspaces

`func (o *WorkspaceUser) GetWorkspaces() []string`

GetWorkspaces returns the Workspaces field if non-nil, zero value otherwise.

### GetWorkspacesOk

`func (o *WorkspaceUser) GetWorkspacesOk() (*[]string, bool)`

GetWorkspacesOk returns a tuple with the Workspaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaces

`func (o *WorkspaceUser) SetWorkspaces(v []string)`

SetWorkspaces sets Workspaces field to given value.


### GetVersion

`func (o *WorkspaceUser) GetVersion() int32`

GetVersion returns the Version field if non-nil, zero value otherwise.

### GetVersionOk

`func (o *WorkspaceUser) GetVersionOk() (*int32, bool)`

GetVersionOk returns a tuple with the Version field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVersion

`func (o *WorkspaceUser) SetVersion(v int32)`

SetVersion sets Version field to given value.

### HasVersion

`func (o *WorkspaceUser) HasVersion() bool`

HasVersion returns a boolean if a field has been set.

### GetStatus

`func (o *WorkspaceUser) GetStatus() WorkspaceUserStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *WorkspaceUser) GetStatusOk() (*WorkspaceUserStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *WorkspaceUser) SetStatus(v WorkspaceUserStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *WorkspaceUser) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetState

`func (o *WorkspaceUser) GetState() WorkspaceUserState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *WorkspaceUser) GetStateOk() (*WorkspaceUserState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *WorkspaceUser) SetState(v WorkspaceUserState)`

SetState sets State field to given value.

### HasState

`func (o *WorkspaceUser) HasState() bool`

HasState returns a boolean if a field has been set.

### GetMessage

`func (o *WorkspaceUser) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *WorkspaceUser) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *WorkspaceUser) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *WorkspaceUser) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetCreatedAt

`func (o *WorkspaceUser) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkspaceUser) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkspaceUser) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkspaceUser) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkspaceUser) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkspaceUser) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkspaceUser) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkspaceUser) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


