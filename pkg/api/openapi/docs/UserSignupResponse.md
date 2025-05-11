# UserSignupResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** | User&#39;s ID | [optional] 
**Name** | Pointer to **string** | User&#39;s name | [optional] 
**Email** | Pointer to **string** | User&#39;s email address | [optional] 
**Role** | Pointer to [**UserRole**](UserRole.md) |  | [optional] 
**Organisation** | Pointer to [**Organisation**](Organisation.md) |  | [optional] 
**JwtToken** | Pointer to **string** | JWT token for the authenticated user | [optional] 

## Methods

### NewUserSignupResponse

`func NewUserSignupResponse() *UserSignupResponse`

NewUserSignupResponse instantiates a new UserSignupResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserSignupResponseWithDefaults

`func NewUserSignupResponseWithDefaults() *UserSignupResponse`

NewUserSignupResponseWithDefaults instantiates a new UserSignupResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *UserSignupResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *UserSignupResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *UserSignupResponse) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *UserSignupResponse) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *UserSignupResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UserSignupResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UserSignupResponse) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *UserSignupResponse) HasName() bool`

HasName returns a boolean if a field has been set.

### GetEmail

`func (o *UserSignupResponse) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserSignupResponse) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserSignupResponse) SetEmail(v string)`

SetEmail sets Email field to given value.

### HasEmail

`func (o *UserSignupResponse) HasEmail() bool`

HasEmail returns a boolean if a field has been set.

### GetRole

`func (o *UserSignupResponse) GetRole() UserRole`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *UserSignupResponse) GetRoleOk() (*UserRole, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *UserSignupResponse) SetRole(v UserRole)`

SetRole sets Role field to given value.

### HasRole

`func (o *UserSignupResponse) HasRole() bool`

HasRole returns a boolean if a field has been set.

### GetOrganisation

`func (o *UserSignupResponse) GetOrganisation() Organisation`

GetOrganisation returns the Organisation field if non-nil, zero value otherwise.

### GetOrganisationOk

`func (o *UserSignupResponse) GetOrganisationOk() (*Organisation, bool)`

GetOrganisationOk returns a tuple with the Organisation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisation

`func (o *UserSignupResponse) SetOrganisation(v Organisation)`

SetOrganisation sets Organisation field to given value.

### HasOrganisation

`func (o *UserSignupResponse) HasOrganisation() bool`

HasOrganisation returns a boolean if a field has been set.

### GetJwtToken

`func (o *UserSignupResponse) GetJwtToken() string`

GetJwtToken returns the JwtToken field if non-nil, zero value otherwise.

### GetJwtTokenOk

`func (o *UserSignupResponse) GetJwtTokenOk() (*string, bool)`

GetJwtTokenOk returns a tuple with the JwtToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetJwtToken

`func (o *UserSignupResponse) SetJwtToken(v string)`

SetJwtToken sets JwtToken field to given value.

### HasJwtToken

`func (o *UserSignupResponse) HasJwtToken() bool`

HasJwtToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


