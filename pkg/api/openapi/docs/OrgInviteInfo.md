# OrgInviteInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**OrgName** | Pointer to **string** |  | [optional] 
**TeamName** | Pointer to **string** |  | [optional] 
**InviterName** | Pointer to **string** |  | [optional] 
**ExpiresAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewOrgInviteInfo

`func NewOrgInviteInfo() *OrgInviteInfo`

NewOrgInviteInfo instantiates a new OrgInviteInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewOrgInviteInfoWithDefaults

`func NewOrgInviteInfoWithDefaults() *OrgInviteInfo`

NewOrgInviteInfoWithDefaults instantiates a new OrgInviteInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOrgName

`func (o *OrgInviteInfo) GetOrgName() string`

GetOrgName returns the OrgName field if non-nil, zero value otherwise.

### GetOrgNameOk

`func (o *OrgInviteInfo) GetOrgNameOk() (*string, bool)`

GetOrgNameOk returns a tuple with the OrgName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrgName

`func (o *OrgInviteInfo) SetOrgName(v string)`

SetOrgName sets OrgName field to given value.

### HasOrgName

`func (o *OrgInviteInfo) HasOrgName() bool`

HasOrgName returns a boolean if a field has been set.

### GetTeamName

`func (o *OrgInviteInfo) GetTeamName() string`

GetTeamName returns the TeamName field if non-nil, zero value otherwise.

### GetTeamNameOk

`func (o *OrgInviteInfo) GetTeamNameOk() (*string, bool)`

GetTeamNameOk returns a tuple with the TeamName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTeamName

`func (o *OrgInviteInfo) SetTeamName(v string)`

SetTeamName sets TeamName field to given value.

### HasTeamName

`func (o *OrgInviteInfo) HasTeamName() bool`

HasTeamName returns a boolean if a field has been set.

### GetInviterName

`func (o *OrgInviteInfo) GetInviterName() string`

GetInviterName returns the InviterName field if non-nil, zero value otherwise.

### GetInviterNameOk

`func (o *OrgInviteInfo) GetInviterNameOk() (*string, bool)`

GetInviterNameOk returns a tuple with the InviterName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInviterName

`func (o *OrgInviteInfo) SetInviterName(v string)`

SetInviterName sets InviterName field to given value.

### HasInviterName

`func (o *OrgInviteInfo) HasInviterName() bool`

HasInviterName returns a boolean if a field has been set.

### GetExpiresAt

`func (o *OrgInviteInfo) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *OrgInviteInfo) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *OrgInviteInfo) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *OrgInviteInfo) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


