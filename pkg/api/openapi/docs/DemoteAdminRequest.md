# DemoteAdminRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProjectName** | **string** | The name of the project to place the demoted user in | 
**Role** | Pointer to [**NullableProjectRole**](ProjectRole.md) | The project role to assign (defaults to Viewer) | [optional] 

## Methods

### NewDemoteAdminRequest

`func NewDemoteAdminRequest(projectName string, ) *DemoteAdminRequest`

NewDemoteAdminRequest instantiates a new DemoteAdminRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDemoteAdminRequestWithDefaults

`func NewDemoteAdminRequestWithDefaults() *DemoteAdminRequest`

NewDemoteAdminRequestWithDefaults instantiates a new DemoteAdminRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProjectName

`func (o *DemoteAdminRequest) GetProjectName() string`

GetProjectName returns the ProjectName field if non-nil, zero value otherwise.

### GetProjectNameOk

`func (o *DemoteAdminRequest) GetProjectNameOk() (*string, bool)`

GetProjectNameOk returns a tuple with the ProjectName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectName

`func (o *DemoteAdminRequest) SetProjectName(v string)`

SetProjectName sets ProjectName field to given value.


### GetRole

`func (o *DemoteAdminRequest) GetRole() ProjectRole`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *DemoteAdminRequest) GetRoleOk() (*ProjectRole, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *DemoteAdminRequest) SetRole(v ProjectRole)`

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


