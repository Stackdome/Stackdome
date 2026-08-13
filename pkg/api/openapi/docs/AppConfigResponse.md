# AppConfigResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**GithubOauth** | Pointer to **bool** | True when GitHub OAuth login is configured on the server. | [optional] 
**Signup** | Pointer to [**SignupConfigResponse**](SignupConfigResponse.md) |  | [optional] 

## Methods

### NewAppConfigResponse

`func NewAppConfigResponse() *AppConfigResponse`

NewAppConfigResponse instantiates a new AppConfigResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAppConfigResponseWithDefaults

`func NewAppConfigResponseWithDefaults() *AppConfigResponse`

NewAppConfigResponseWithDefaults instantiates a new AppConfigResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetGithubOauth

`func (o *AppConfigResponse) GetGithubOauth() bool`

GetGithubOauth returns the GithubOauth field if non-nil, zero value otherwise.

### GetGithubOauthOk

`func (o *AppConfigResponse) GetGithubOauthOk() (*bool, bool)`

GetGithubOauthOk returns a tuple with the GithubOauth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGithubOauth

`func (o *AppConfigResponse) SetGithubOauth(v bool)`

SetGithubOauth sets GithubOauth field to given value.

### HasGithubOauth

`func (o *AppConfigResponse) HasGithubOauth() bool`

HasGithubOauth returns a boolean if a field has been set.

### GetSignup

`func (o *AppConfigResponse) GetSignup() SignupConfigResponse`

GetSignup returns the Signup field if non-nil, zero value otherwise.

### GetSignupOk

`func (o *AppConfigResponse) GetSignupOk() (*SignupConfigResponse, bool)`

GetSignupOk returns a tuple with the Signup field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSignup

`func (o *AppConfigResponse) SetSignup(v SignupConfigResponse)`

SetSignup sets Signup field to given value.

### HasSignup

`func (o *AppConfigResponse) HasSignup() bool`

HasSignup returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


