# UserSignupRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | User&#39;s name | 
**Username** | Pointer to **string** | User&#39;s username | [optional] 
**Email** | **string** | User&#39;s email address | 
**Password** | **string** | Users desired password | 
**Role** | Pointer to [**UserRole**](UserRole.md) |  | [optional] 
**Organisation** | Pointer to [**Organisation**](Organisation.md) |  | [optional] 
**OrganisationId** | Pointer to **string** | OrganisationID | [optional] 

## Methods

### NewUserSignupRequest

`func NewUserSignupRequest(name string, email string, password string, ) *UserSignupRequest`

NewUserSignupRequest instantiates a new UserSignupRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserSignupRequestWithDefaults

`func NewUserSignupRequestWithDefaults() *UserSignupRequest`

NewUserSignupRequestWithDefaults instantiates a new UserSignupRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *UserSignupRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UserSignupRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UserSignupRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUsername

`func (o *UserSignupRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *UserSignupRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *UserSignupRequest) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *UserSignupRequest) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetEmail

`func (o *UserSignupRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserSignupRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserSignupRequest) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetPassword

`func (o *UserSignupRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *UserSignupRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *UserSignupRequest) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetRole

`func (o *UserSignupRequest) GetRole() UserRole`

GetRole returns the Role field if non-nil, zero value otherwise.

### GetRoleOk

`func (o *UserSignupRequest) GetRoleOk() (*UserRole, bool)`

GetRoleOk returns a tuple with the Role field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRole

`func (o *UserSignupRequest) SetRole(v UserRole)`

SetRole sets Role field to given value.

### HasRole

`func (o *UserSignupRequest) HasRole() bool`

HasRole returns a boolean if a field has been set.

### GetOrganisation

`func (o *UserSignupRequest) GetOrganisation() Organisation`

GetOrganisation returns the Organisation field if non-nil, zero value otherwise.

### GetOrganisationOk

`func (o *UserSignupRequest) GetOrganisationOk() (*Organisation, bool)`

GetOrganisationOk returns a tuple with the Organisation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisation

`func (o *UserSignupRequest) SetOrganisation(v Organisation)`

SetOrganisation sets Organisation field to given value.

### HasOrganisation

`func (o *UserSignupRequest) HasOrganisation() bool`

HasOrganisation returns a boolean if a field has been set.

### GetOrganisationId

`func (o *UserSignupRequest) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *UserSignupRequest) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *UserSignupRequest) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *UserSignupRequest) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


