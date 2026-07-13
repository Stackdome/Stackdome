# OrgInviteCreateResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Email** | Pointer to **string** |  | [optional] 
**OrganisationId** | Pointer to **string** |  | [optional] 
**ProjectName** | Pointer to **string** |  | [optional] 
**Role** | Pointer to **string** |  | [optional] 
**Status** | Pointer to [**InviteStatus**](InviteStatus.md) |  | [optional] 
**ExpiresAt** | Pointer to **time.Time** |  | [optional] 
**InvitedBy** | Pointer to **string** |  | [optional] 
**EmailSent** | Pointer to **bool** |  | [optional] 
**InviteToken** | Pointer to **string** | Raw invite token (shown only once at creation) | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewOrgInviteCreateResponse

`func NewOrgInviteCreateResponse() *OrgInviteCreateResponse`

NewOrgInviteCreateResponse instantiates a new OrgInviteCreateResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOrgInviteCreateResponseWithDefaults

`func NewOrgInviteCreateResponseWithDefaults() *OrgInviteCreateResponse`

NewOrgInviteCreateResponseWithDefaults instantiates a new OrgInviteCreateResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *OrgInviteCreateResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *OrgInviteCreateResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *OrgInviteCreateResponse) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *OrgInviteCreateResponse) HasId() bool`

HasId returns a boolean if a field has been set.

### GetEmail

`func (o *OrgInviteCreateResponse) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *OrgInviteCreateResponse) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *OrgInviteCreateResponse) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *OrgInviteCreateResponse) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetOrganisationId

`func (o *OrgInviteCreateResponse) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *OrgInviteCreateResponse) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *OrgInviteCreateResponse) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *OrgInviteCreateResponse) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetProjectName

`func (o *OrgInviteCreateResponse) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *OrgInviteCreateResponse) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *OrgInviteCreateResponse) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *OrgInviteCreateResponse) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetRole

`func (o *OrgInviteCreateResponse) GetRole() string`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *OrgInviteCreateResponse) GetRoleOk() (*string, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *OrgInviteCreateResponse) SetRole(v string)`

SetRole sets Role field to given value.

### HasRole

`func (o *OrgInviteCreateResponse) HasRole() bool`

HasRole returns a boolean if a field has been set.

### GetStatus

`func (o *OrgInviteCreateResponse) GetStatus() InviteStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *OrgInviteCreateResponse) GetStatusOk() (*InviteStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *OrgInviteCreateResponse) SetStatus(v InviteStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *OrgInviteCreateResponse) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetExpiresAt

`func (o *OrgInviteCreateResponse) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *OrgInviteCreateResponse) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *OrgInviteCreateResponse) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *OrgInviteCreateResponse) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.

### GetInvitedBy

`func (o *OrgInviteCreateResponse) GetInvitedBy() string`

GetInvitedBy returns the InvitedBy field if non-nil, zero value otherwise.

### GetInvitedByOk

`func (o *OrgInviteCreateResponse) GetInvitedByOk() (*string, bool)`

GetInvitedByOk returns a tuple with the InvitedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvitedBy

`func (o *OrgInviteCreateResponse) SetInvitedBy(v string)`

SetInvitedBy sets InvitedBy field to given value.

### HasInvitedBy

`func (o *OrgInviteCreateResponse) HasInvitedBy() bool`

HasInvitedBy returns a boolean if a field has been set.

### GetEmailSent

`func (o *OrgInviteCreateResponse) GetEmailSent() bool`

GetEmailSent returns the EmailSent field if non-nil, zero value otherwise.

### GetEmailSentOk

`func (o *OrgInviteCreateResponse) GetEmailSentOk() (*bool, bool)`

GetEmailSentOk returns a tuple with the EmailSent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmailSent

`func (o *OrgInviteCreateResponse) SetEmailSent(v bool)`

SetEmailSent sets EmailSent field to given value.

### HasEmailSent

`func (o *OrgInviteCreateResponse) HasEmailSent() bool`

HasEmailSent returns a boolean if a field has been set.

### GetInviteToken

`func (o *OrgInviteCreateResponse) GetInviteToken() string`

GetInviteToken returns the InviteToken field if non-nil, zero value otherwise.

### GetInviteTokenOk

`func (o *OrgInviteCreateResponse) GetInviteTokenOk() (*string, bool)`

GetInviteTokenOk returns a tuple with the InviteToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInviteToken

`func (o *OrgInviteCreateResponse) SetInviteToken(v string)`

SetInviteToken sets InviteToken field to given value.

### HasInviteToken

`func (o *OrgInviteCreateResponse) HasInviteToken() bool`

HasInviteToken returns a boolean if a field has been set.

### GetCreatedAt

`func (o *OrgInviteCreateResponse) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *OrgInviteCreateResponse) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *OrgInviteCreateResponse) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *OrgInviteCreateResponse) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


