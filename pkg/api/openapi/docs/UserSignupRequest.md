# UserSignupRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | User&#39;s name | 
**Email** | **string** | User&#39;s email address | 
**Password** | **string** | Users desired password | 
**Organisation** | Pointer to [**Organisation**](Organisation.md) |  | [optional] 
**InviteToken** | Pointer to **string** | Optional invite token for joining an existing organization | [optional] 
**TurnstileToken** | Pointer to **string** | Turnstile token required for protected public password signup | [optional] 

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

### GetInviteToken

`func (o *UserSignupRequest) GetInviteToken() string`

GetInviteToken returns the InviteToken field if non-nil, zero value otherwise.

### GetInviteTokenOk

`func (o *UserSignupRequest) GetInviteTokenOk() (*string, bool)`

GetInviteTokenOk returns a tuple with the InviteToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInviteToken

`func (o *UserSignupRequest) SetInviteToken(v string)`

SetInviteToken sets InviteToken field to given value.

### HasInviteToken

`func (o *UserSignupRequest) HasInviteToken() bool`

HasInviteToken returns a boolean if a field has been set.

### GetTurnstileToken

`func (o *UserSignupRequest) GetTurnstileToken() string`

GetTurnstileToken returns the TurnstileToken field if non-nil, zero value otherwise.

### GetTurnstileTokenOk

`func (o *UserSignupRequest) GetTurnstileTokenOk() (*string, bool)`

GetTurnstileTokenOk returns a tuple with the TurnstileToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTurnstileToken

`func (o *UserSignupRequest) SetTurnstileToken(v string)`

SetTurnstileToken sets TurnstileToken field to given value.

### HasTurnstileToken

`func (o *UserSignupRequest) HasTurnstileToken() bool`

HasTurnstileToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


