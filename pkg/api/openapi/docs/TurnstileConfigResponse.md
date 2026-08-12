# TurnstileConfigResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Enabled** | **bool** | True when password signup requires a Turnstile challenge. | 
**SiteKey** | **string** | Public Turnstile site key used to render the signup widget. | 
**Action** | **string** | Turnstile action submitted by the signup widget. | 

## Methods

### NewTurnstileConfigResponse

`func NewTurnstileConfigResponse(enabled bool, siteKey string, action string, ) *TurnstileConfigResponse`

NewTurnstileConfigResponse instantiates a new TurnstileConfigResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTurnstileConfigResponseWithDefaults

`func NewTurnstileConfigResponseWithDefaults() *TurnstileConfigResponse`

NewTurnstileConfigResponseWithDefaults instantiates a new TurnstileConfigResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnabled

`func (o *TurnstileConfigResponse) GetEnabled() bool`

GetEnabled returns the Enabled field if non-nil, zero value otherwise.

### GetEnabledOk

`func (o *TurnstileConfigResponse) GetEnabledOk() (*bool, bool)`

GetEnabledOk returns a tuple with the Enabled field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnabled

`func (o *TurnstileConfigResponse) SetEnabled(v bool)`

SetEnabled sets Enabled field to given value.


### GetSiteKey

`func (o *TurnstileConfigResponse) GetSiteKey() string`

GetSiteKey returns the SiteKey field if non-nil, zero value otherwise.

### GetSiteKeyOk

`func (o *TurnstileConfigResponse) GetSiteKeyOk() (*string, bool)`

GetSiteKeyOk returns a tuple with the SiteKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSiteKey

`func (o *TurnstileConfigResponse) SetSiteKey(v string)`

SetSiteKey sets SiteKey field to given value.


### GetAction

`func (o *TurnstileConfigResponse) GetAction() string`

GetAction returns the Action field if non-nil, zero value otherwise.

### GetActionOk

`func (o *TurnstileConfigResponse) GetActionOk() (*string, bool)`

GetActionOk returns a tuple with the Action field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAction

`func (o *TurnstileConfigResponse) SetAction(v string)`

SetAction sets Action field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


