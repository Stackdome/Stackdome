# OrgInvite

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**Email** | Pointer to **string** |  | [optional] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**ProjectName** | Pointer to **string** |  | [optional] 
**Role** | Pointer to **string** |  | [optional] 
**Status** | Pointer to [**InviteStatus**](InviteStatus.md) |  | [optional] 
**ExpiresAt** | Pointer to **time.Time** |  | [optional] 
**InvitedBy** | Pointer to **string** |  | [optional] 
**EmailSent** | Pointer to **bool** |  | [optional] 
**EmailError** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**AcceptedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewOrgInvite

`func NewOrgInvite() *OrgInvite`

NewOrgInvite instantiates a new OrgInvite object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOrgInviteWithDefaults

`func NewOrgInviteWithDefaults() *OrgInvite`

NewOrgInviteWithDefaults instantiates a new OrgInvite object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *OrgInvite) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *OrgInvite) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *OrgInvite) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *OrgInvite) HasId() bool`

HasId returns a boolean if a field has been set.

### GetEmail

`func (o *OrgInvite) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *OrgInvite) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *OrgInvite) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *OrgInvite) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetOrganisationId

`func (o *OrgInvite) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *OrgInvite) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *OrgInvite) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *OrgInvite) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetProjectName

`func (o *OrgInvite) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *OrgInvite) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *OrgInvite) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.

### HasProjectName

`func (o *OrgInvite) HasProjectName() bool`

HasProjectName returns a boolean if a field has been set.

### GetRole

`func (o *OrgInvite) GetRole() string`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *OrgInvite) GetRoleOk() (*string, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *OrgInvite) SetRole(v string)`

SetRole sets Role field to given value.

### HasRole

`func (o *OrgInvite) HasRole() bool`

HasRole returns a boolean if a field has been set.

### GetStatus

`func (o *OrgInvite) GetStatus() InviteStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *OrgInvite) GetStatusOk() (*InviteStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *OrgInvite) SetStatus(v InviteStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *OrgInvite) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetExpiresAt

`func (o *OrgInvite) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *OrgInvite) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *OrgInvite) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *OrgInvite) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.

### GetInvitedBy

`func (o *OrgInvite) GetInvitedBy() string`

GetInvitedBy returns the InvitedBy field if non-nil, zero value otherwise.

### GetInvitedByOk

`func (o *OrgInvite) GetInvitedByOk() (*string, bool)`

GetInvitedByOk returns a tuple with the InvitedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInvitedBy

`func (o *OrgInvite) SetInvitedBy(v string)`

SetInvitedBy sets InvitedBy field to given value.

### HasInvitedBy

`func (o *OrgInvite) HasInvitedBy() bool`

HasInvitedBy returns a boolean if a field has been set.

### GetEmailSent

`func (o *OrgInvite) GetEmailSent() bool`

GetEmailSent returns the EmailSent field if non-nil, zero value otherwise.

### GetEmailSentOk

`func (o *OrgInvite) GetEmailSentOk() (*bool, bool)`

GetEmailSentOk returns a tuple with the EmailSent field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmailSent

`func (o *OrgInvite) SetEmailSent(v bool)`

SetEmailSent sets EmailSent field to given value.

### HasEmailSent

`func (o *OrgInvite) HasEmailSent() bool`

HasEmailSent returns a boolean if a field has been set.

### GetEmailError

`func (o *OrgInvite) GetEmailError() string`

GetEmailError returns the EmailError field if non-nil, zero value otherwise.

### GetEmailErrorOk

`func (o *OrgInvite) GetEmailErrorOk() (*string, bool)`

GetEmailErrorOk returns a tuple with the EmailError field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmailError

`func (o *OrgInvite) SetEmailError(v string)`

SetEmailError sets EmailError field to given value.

### HasEmailError

`func (o *OrgInvite) HasEmailError() bool`

HasEmailError returns a boolean if a field has been set.

### GetCreatedAt

`func (o *OrgInvite) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *OrgInvite) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *OrgInvite) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *OrgInvite) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetAcceptedAt

`func (o *OrgInvite) GetAcceptedAt() time.Time`

GetAcceptedAt returns the AcceptedAt field if non-nil, zero value otherwise.

### GetAcceptedAtOk

`func (o *OrgInvite) GetAcceptedAtOk() (*time.Time, bool)`

GetAcceptedAtOk returns a tuple with the AcceptedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAcceptedAt

`func (o *OrgInvite) SetAcceptedAt(v time.Time)`

SetAcceptedAt sets AcceptedAt field to given value.

### HasAcceptedAt

`func (o *OrgInvite) HasAcceptedAt() bool`

HasAcceptedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


