# UserSignupResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**User** | Pointer to [**User**](User.md) |  | [optional] 
**JwtToken** | Pointer to **string** | JWT token for the authenticated user | [optional] 
**RefreshToken** | Pointer to **string** | Refresh token for obtaining new access tokens | [optional] 

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

### GetUser

`func (o *UserSignupResponse) GetUser() User`

GetUser returns the User field if non-nil, zero value otherwise.

### GetUserOk

`func (o *UserSignupResponse) GetUserOk() (*User, bool)`

GetUserOk returns a tuple with the User field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUser

`func (o *UserSignupResponse) SetUser(v User)`

SetUser sets User field to given value.

### HasUser

`func (o *UserSignupResponse) HasUser() bool`

HasUser returns a boolean if a field has been set.

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

### GetRefreshToken

`func (o *UserSignupResponse) GetRefreshToken() string`

GetRefreshToken returns the RefreshToken field if non-nil, zero value otherwise.

### GetRefreshTokenOk

`func (o *UserSignupResponse) GetRefreshTokenOk() (*string, bool)`

GetRefreshTokenOk returns a tuple with the RefreshToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRefreshToken

`func (o *UserSignupResponse) SetRefreshToken(v string)`

SetRefreshToken sets RefreshToken field to given value.

### HasRefreshToken

`func (o *UserSignupResponse) HasRefreshToken() bool`

HasRefreshToken returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


