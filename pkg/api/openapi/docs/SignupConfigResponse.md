# SignupConfigResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Turnstile** | [**TurnstileConfigResponse**](TurnstileConfigResponse.md) |  | 

## Methods

### NewSignupConfigResponse

`func NewSignupConfigResponse(turnstile TurnstileConfigResponse, ) *SignupConfigResponse`

NewSignupConfigResponse instantiates a new SignupConfigResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSignupConfigResponseWithDefaults

`func NewSignupConfigResponseWithDefaults() *SignupConfigResponse`

NewSignupConfigResponseWithDefaults instantiates a new SignupConfigResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTurnstile

`func (o *SignupConfigResponse) GetTurnstile() TurnstileConfigResponse`

GetTurnstile returns the Turnstile field if non-nil, zero value otherwise.

### GetTurnstileOk

`func (o *SignupConfigResponse) GetTurnstileOk() (*TurnstileConfigResponse, bool)`

GetTurnstileOk returns a tuple with the Turnstile field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTurnstile

`func (o *SignupConfigResponse) SetTurnstile(v TurnstileConfigResponse)`

SetTurnstile sets Turnstile field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


