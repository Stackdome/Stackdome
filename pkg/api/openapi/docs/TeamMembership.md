# TeamMembership

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**TeamId** | **string** |  | 
**UserId** | **string** |  | 
**Role** | **string** |  | 
**Team** | Pointer to [**Team**](Team.md) |  | [optional] 
**User** | Pointer to [**User**](User.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewTeamMembership

`func NewTeamMembership(teamId string, userId string, role string, ) *TeamMembership`

NewTeamMembership instantiates a new TeamMembership object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTeamMembershipWithDefaults

`func NewTeamMembershipWithDefaults() *TeamMembership`

NewTeamMembershipWithDefaults instantiates a new TeamMembership object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *TeamMembership) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *TeamMembership) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *TeamMembership) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *TeamMembership) HasId() bool`

HasId returns a boolean if a field has been set.

### GetTeamId

`func (o *TeamMembership) GetTeamId() string`

GetTeamId returns the TeamId field if non-nil, zero value otherwise.

### GetTeamIdOk

`func (o *TeamMembership) GetTeamIdOk() (*string, bool)`

GetTeamIdOk returns a tuple with the TeamId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTeamId

`func (o *TeamMembership) SetTeamId(v string)`

SetTeamId sets TeamId field to given value.


### GetUserId

`func (o *TeamMembership) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *TeamMembership) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *TeamMembership) SetUserId(v string)`

SetUserId sets UserId field to given value.


### GetRole

`func (o *TeamMembership) GetRole() string`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *TeamMembership) GetRoleOk() (*string, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *TeamMembership) SetRole(v string)`

SetRole sets Role field to given value.


### GetTeam

`func (o *TeamMembership) GetTeam() Team`

GetTeam returns the Team field if non-nil, zero value otherwise.

### GetTeamOk

`func (o *TeamMembership) GetTeamOk() (*Team, bool)`

GetTeamOk returns a tuple with the Team field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTeam

`func (o *TeamMembership) SetTeam(v Team)`

SetTeam sets Team field to given value.

### HasTeam

`func (o *TeamMembership) HasTeam() bool`

HasTeam returns a boolean if a field has been set.

### GetUser

`func (o *TeamMembership) GetUser() User`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *TeamMembership) GetUserOk() (*User, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *TeamMembership) SetUser(v User)`

SetUser sets User field to given value.

### HasUser

`func (o *TeamMembership) HasUser() bool`

HasUser returns a boolean if a field has been set.

### GetCreatedAt

`func (o *TeamMembership) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *TeamMembership) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *TeamMembership) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *TeamMembership) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


