# UserCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | User&#39;s name | 
**Username** | Pointer to **string** | User&#39;s username | [optional] 
**Email** | **string** | User&#39;s email address | 
**Password** | **string** | Users desired password | 
**Organisation** | Pointer to **string** | User&#39;s organisation | [optional] 
**OrganisationId** | **string** | OrganisationID | 

## Methods

### NewUserCreateRequest

`func NewUserCreateRequest(name string, email string, password string, organisationId string, ) *UserCreateRequest`

NewUserCreateRequest instantiates a new UserCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewUserCreateRequestWithDefaults

`func NewUserCreateRequestWithDefaults() *UserCreateRequest`

NewUserCreateRequestWithDefaults instantiates a new UserCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *UserCreateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *UserCreateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *UserCreateRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUsername

`func (o *UserCreateRequest) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *UserCreateRequest) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *UserCreateRequest) SetUsername(v string)`

SetUsername sets Username field to given value.

### HasUsername

`func (o *UserCreateRequest) HasUsername() bool`

HasUsername returns a boolean if a field has been set.

### GetEmail

`func (o *UserCreateRequest) GetEmail() string`

GetEmail returns the Email field if non-nil, zero value otherwise.

### GetEmailOk

`func (o *UserCreateRequest) GetEmailOk() (*string, bool)`

GetEmailOk returns a tuple with the Email field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmail

`func (o *UserCreateRequest) SetEmail(v string)`

SetEmail sets Email field to given value.


### GetPassword

`func (o *UserCreateRequest) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *UserCreateRequest) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *UserCreateRequest) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetOrganisation

`func (o *UserCreateRequest) GetOrganisation() string`

GetOrganisation returns the Organisation field if non-nil, zero value otherwise.

### GetOrganisationOk

`func (o *UserCreateRequest) GetOrganisationOk() (*string, bool)`

GetOrganisationOk returns a tuple with the Organisation field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisation

`func (o *UserCreateRequest) SetOrganisation(v string)`

SetOrganisation sets Organisation field to given value.

### HasOrganisation

`func (o *UserCreateRequest) HasOrganisation() bool`

HasOrganisation returns a boolean if a field has been set.

### GetOrganisationId

`func (o *UserCreateRequest) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *UserCreateRequest) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *UserCreateRequest) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


