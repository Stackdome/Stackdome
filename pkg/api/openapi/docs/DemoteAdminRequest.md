# DemoteAdminRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**TeamName** | **string** | The name of the team to place the demoted user in | 
**Role** | Pointer to [**NullableTeamRole**](TeamRole.md) | The team role to assign (defaults to Viewer) | [optional] 

## Methods

### NewDemoteAdminRequest

`func NewDemoteAdminRequest(teamName string, ) *DemoteAdminRequest`

NewDemoteAdminRequest instantiates a new DemoteAdminRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDemoteAdminRequestWithDefaults

`func NewDemoteAdminRequestWithDefaults() *DemoteAdminRequest`

NewDemoteAdminRequestWithDefaults instantiates a new DemoteAdminRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTeamName

`func (o *DemoteAdminRequest) GetTeamName() string`

GetTeamName returns the TeamName field if non-nil, zero value otherwise.

### GetTeamNameOk

`func (o *DemoteAdminRequest) GetTeamNameOk() (*string, bool)`

GetTeamNameOk returns a tuple with the TeamName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTeamName

`func (o *DemoteAdminRequest) SetTeamName(v string)`

SetTeamName sets TeamName field to given value.


### GetRole

`func (o *DemoteAdminRequest) GetRole() TeamRole`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *DemoteAdminRequest) GetRoleOk() (*TeamRole, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *DemoteAdminRequest) SetRole(v TeamRole)`

SetRole sets Role field to given value.

### HasRole

`func (o *DemoteAdminRequest) HasRole() bool`

HasRole returns a boolean if a field has been set.

### SetRoleNil

`func (o *DemoteAdminRequest) SetRoleNil(b bool)`

 SetRoleNil sets the value for Role to be an explicit nil

### UnsetRole
`func (o *DemoteAdminRequest) UnsetRole()`

UnsetRole ensures that no value is present for Role, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


