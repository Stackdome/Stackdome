# APITokenCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | Display name for the token | 
**Scopes** | **[]string** | Permission scopes for the token | 
**ResourceIds** | Pointer to **[]string** | Optional list of resource IDs to restrict the token to | [optional] 
**ExpiresAt** | Pointer to **time.Time** | Optional expiration time in RFC3339 format | [optional] 

## Methods

### NewAPITokenCreateRequest

`func NewAPITokenCreateRequest(name string, scopes []string, ) *APITokenCreateRequest`

NewAPITokenCreateRequest instantiates a new APITokenCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAPITokenCreateRequestWithDefaults

`func NewAPITokenCreateRequestWithDefaults() *APITokenCreateRequest`

NewAPITokenCreateRequestWithDefaults instantiates a new APITokenCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *APITokenCreateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *APITokenCreateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *APITokenCreateRequest) SetName(v string)`

SetName sets Name field to given value.


### GetScopes

`func (o *APITokenCreateRequest) GetScopes() []string`

GetScopes returns the Scopes field if non-nil, zero value otherwise.

### GetScopesOk

`func (o *APITokenCreateRequest) GetScopesOk() (*[]string, bool)`

GetScopesOk returns a tuple with the Scopes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetScopes

`func (o *APITokenCreateRequest) SetScopes(v []string)`

SetScopes sets Scopes field to given value.


### GetResourceIds

`func (o *APITokenCreateRequest) GetResourceIds() []string`

GetResourceIds returns the ResourceIds field if non-nil, zero value otherwise.

### GetResourceIdsOk

`func (o *APITokenCreateRequest) GetResourceIdsOk() (*[]string, bool)`

GetResourceIdsOk returns a tuple with the ResourceIds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceIds

`func (o *APITokenCreateRequest) SetResourceIds(v []string)`

SetResourceIds sets ResourceIds field to given value.

### HasResourceIds

`func (o *APITokenCreateRequest) HasResourceIds() bool`

HasResourceIds returns a boolean if a field has been set.

### GetExpiresAt

`func (o *APITokenCreateRequest) GetExpiresAt() time.Time`

GetExpiresAt returns the ExpiresAt field if non-nil, zero value otherwise.

### GetExpiresAtOk

`func (o *APITokenCreateRequest) GetExpiresAtOk() (*time.Time, bool)`

GetExpiresAtOk returns a tuple with the ExpiresAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExpiresAt

`func (o *APITokenCreateRequest) SetExpiresAt(v time.Time)`

SetExpiresAt sets ExpiresAt field to given value.

### HasExpiresAt

`func (o *APITokenCreateRequest) HasExpiresAt() bool`

HasExpiresAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


