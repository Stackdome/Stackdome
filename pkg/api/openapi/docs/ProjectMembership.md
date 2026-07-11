# ProjectMembership

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**ProjectId** | **string** |  | 
**UserId** | **string** |  | 
**Role** | **string** |  | 
**Project** | Pointer to [**Project**](Project.md) |  | [optional] 
**User** | Pointer to [**User**](User.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewProjectMembership

`func NewProjectMembership(projectId string, userId string, role string, ) *ProjectMembership`

NewProjectMembership instantiates a new ProjectMembership object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectMembershipWithDefaults

`func NewProjectMembershipWithDefaults() *ProjectMembership`

NewProjectMembershipWithDefaults instantiates a new ProjectMembership object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ProjectMembership) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ProjectMembership) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ProjectMembership) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ProjectMembership) HasId() bool`

HasId returns a boolean if a field has been set.

### GetProjectId

`func (o *ProjectMembership) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *ProjectMembership) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *ProjectMembership) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetUserId

`func (o *ProjectMembership) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *ProjectMembership) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *ProjectMembership) SetUserId(v string)`

SetUserId sets UserId field to given value.


### GetRole

`func (o *ProjectMembership) GetRole() string`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *ProjectMembership) GetRoleOk() (*string, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *ProjectMembership) SetRole(v string)`

SetRole sets Role field to given value.


### GetProject

`func (o *ProjectMembership) GetProject() Project`

GetProject returns the Project field if non-nil, zero value otherwise.

### GetProjectOk

`func (o *ProjectMembership) GetProjectOk() (*Project, bool)`

GetProjectOk returns a tuple with the Project field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProject

`func (o *ProjectMembership) SetProject(v Project)`

SetProject sets Project field to given value.

### HasProject

`func (o *ProjectMembership) HasProject() bool`

HasProject returns a boolean if a field has been set.

### GetUser

`func (o *ProjectMembership) GetUser() User`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *ProjectMembership) GetUserOk() (*User, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *ProjectMembership) SetUser(v User)`

SetUser sets User field to given value.

### HasUser

`func (o *ProjectMembership) HasUser() bool`

HasUser returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ProjectMembership) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ProjectMembership) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ProjectMembership) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ProjectMembership) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


