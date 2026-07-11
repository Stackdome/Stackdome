# OrgInviteCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Email** | **string** |  | 
**ProjectName** | **string** |  | 
**Role** | **string** |  | 
**ExpiresInDays** | **int32** |  | 

## Methods

### NewOrgInviteCreateRequest

`func NewOrgInviteCreateRequest(email string, projectName string, role string, expiresInDays int32, ) *OrgInviteCreateRequest`

NewOrgInviteCreateRequest instantiates a new OrgInviteCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOrgInviteCreateRequestWithDefaults

`func NewOrgInviteCreateRequestWithDefaults() *OrgInviteCreateRequest`

NewOrgInviteCreateRequestWithDefaults instantiates a new OrgInviteCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEmail

`func (o *OrgInviteCreateRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *OrgInviteCreateRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *OrgInviteCreateRequest) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetProjectName

`func (o *OrgInviteCreateRequest) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *OrgInviteCreateRequest) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *OrgInviteCreateRequest) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetRole

`func (o *OrgInviteCreateRequest) GetRole() string`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *OrgInviteCreateRequest) GetRoleOk() (*string, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *OrgInviteCreateRequest) SetRole(v string)`

SetRole sets Role field to given value.


### GetExpiresInDays

`func (o *OrgInviteCreateRequest) GetExpiresInDays() int32`

GetExpiresInDays returns the ExpiresInDays field if non-nil, zero value otherwise.

### GetExpiresInDaysOk

`func (o *OrgInviteCreateRequest) GetExpiresInDaysOk() (*int32, bool)`

GetExpiresInDaysOk returns a tuple with the ExpiresInDays field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresInDays

`func (o *OrgInviteCreateRequest) SetExpiresInDays(v int32)`

SetExpiresInDays sets ExpiresInDays field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


