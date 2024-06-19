# WorkspaceProvisionRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**UserId** | Pointer to **string** |  | [optional] 
**OrgId** | Pointer to **int32** |  | [optional] 
**SshPublicKey** | **string** |  | 
**Status** | Pointer to [**WorkspaceProvisionRequestStatus**](WorkspaceProvisionRequestStatus.md) |  | [optional] 
**State** | Pointer to [**WorkspaceProvisionRequestState**](WorkspaceProvisionRequestState.md) |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewWorkspaceProvisionRequest

`func NewWorkspaceProvisionRequest(sshPublicKey string, ) *WorkspaceProvisionRequest`

NewWorkspaceProvisionRequest instantiates a new WorkspaceProvisionRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceProvisionRequestWithDefaults

`func NewWorkspaceProvisionRequestWithDefaults() *WorkspaceProvisionRequest`

NewWorkspaceProvisionRequestWithDefaults instantiates a new WorkspaceProvisionRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *WorkspaceProvisionRequest) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *WorkspaceProvisionRequest) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *WorkspaceProvisionRequest) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *WorkspaceProvisionRequest) HasId() bool`

HasId returns a boolean if a field has been set.

### GetUserId

`func (o *WorkspaceProvisionRequest) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *WorkspaceProvisionRequest) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *WorkspaceProvisionRequest) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *WorkspaceProvisionRequest) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetOrgId

`func (o *WorkspaceProvisionRequest) GetOrgId() int32`

GetOrgId returns the OrgId field if non-nil, zero value otherwise.

### GetOrgIdOk

`func (o *WorkspaceProvisionRequest) GetOrgIdOk() (*int32, bool)`

GetOrgIdOk returns a tuple with the OrgId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgId

`func (o *WorkspaceProvisionRequest) SetOrgId(v int32)`

SetOrgId sets OrgId field to given value.

### HasOrgId

`func (o *WorkspaceProvisionRequest) HasOrgId() bool`

HasOrgId returns a boolean if a field has been set.

### GetSshPublicKey

`func (o *WorkspaceProvisionRequest) GetSshPublicKey() string`

GetSshPublicKey returns the SshPublicKey field if non-nil, zero value otherwise.

### GetSshPublicKeyOk

`func (o *WorkspaceProvisionRequest) GetSshPublicKeyOk() (*string, bool)`

GetSshPublicKeyOk returns a tuple with the SshPublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshPublicKey

`func (o *WorkspaceProvisionRequest) SetSshPublicKey(v string)`

SetSshPublicKey sets SshPublicKey field to given value.


### GetStatus

`func (o *WorkspaceProvisionRequest) GetStatus() WorkspaceProvisionRequestStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *WorkspaceProvisionRequest) GetStatusOk() (*WorkspaceProvisionRequestStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *WorkspaceProvisionRequest) SetStatus(v WorkspaceProvisionRequestStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *WorkspaceProvisionRequest) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetState

`func (o *WorkspaceProvisionRequest) GetState() WorkspaceProvisionRequestState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *WorkspaceProvisionRequest) GetStateOk() (*WorkspaceProvisionRequestState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *WorkspaceProvisionRequest) SetState(v WorkspaceProvisionRequestState)`

SetState sets State field to given value.

### HasState

`func (o *WorkspaceProvisionRequest) HasState() bool`

HasState returns a boolean if a field has been set.

### GetMessage

`func (o *WorkspaceProvisionRequest) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *WorkspaceProvisionRequest) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *WorkspaceProvisionRequest) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *WorkspaceProvisionRequest) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetCreatedAt

`func (o *WorkspaceProvisionRequest) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *WorkspaceProvisionRequest) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *WorkspaceProvisionRequest) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *WorkspaceProvisionRequest) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *WorkspaceProvisionRequest) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *WorkspaceProvisionRequest) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *WorkspaceProvisionRequest) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *WorkspaceProvisionRequest) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


