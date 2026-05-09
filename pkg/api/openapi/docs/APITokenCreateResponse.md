# APITokenCreateResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Token** | Pointer to **string** | The raw API token (only returned on creation) | [optional] 
**Id** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**TokenPrefix** | Pointer to **string** | Prefix of the token for identification | [optional] 
**ExpiresAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewAPITokenCreateResponse

`func NewAPITokenCreateResponse() *APITokenCreateResponse`

NewAPITokenCreateResponse instantiates a new APITokenCreateResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAPITokenCreateResponseWithDefaults

`func NewAPITokenCreateResponseWithDefaults() *APITokenCreateResponse`

NewAPITokenCreateResponseWithDefaults instantiates a new APITokenCreateResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToken

`func (o *APITokenCreateResponse) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *APITokenCreateResponse) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *APITokenCreateResponse) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *APITokenCreateResponse) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetId

`func (o *APITokenCreateResponse) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *APITokenCreateResponse) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *APITokenCreateResponse) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *APITokenCreateResponse) HasId() bool`

HasId returns a boolean if a field has been set.

### GetName

`func (o *APITokenCreateResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *APITokenCreateResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *APITokenCreateResponse) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *APITokenCreateResponse) HasName() bool`

HasName returns a boolean if a field has been set.

### GetTokenPrefix

`func (o *APITokenCreateResponse) GetTokenPrefix() string`

GetTokenPrefix returns the TokenPrefix field if non-nil, zero value otherwise.

### GetTokenPrefixOk

`func (o *APITokenCreateResponse) GetTokenPrefixOk() (*string, bool)`

GetTokenPrefixOk returns a tuple with the TokenPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTokenPrefix

`func (o *APITokenCreateResponse) SetTokenPrefix(v string)`

SetTokenPrefix sets TokenPrefix field to given value.

### HasTokenPrefix

`func (o *APITokenCreateResponse) HasTokenPrefix() bool`

HasTokenPrefix returns a boolean if a field has been set.

### GetExpiresAt

`func (o *APITokenCreateResponse) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *APITokenCreateResponse) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *APITokenCreateResponse) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *APITokenCreateResponse) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


