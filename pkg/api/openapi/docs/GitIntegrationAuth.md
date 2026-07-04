# GitIntegrationAuth

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Token** | Pointer to **string** |  | [optional] 
**Basic** | Pointer to [**GitIntegrationBasicAuth**](GitIntegrationBasicAuth.md) |  | [optional] 

## Methods

### NewGitIntegrationAuth

`func NewGitIntegrationAuth() *GitIntegrationAuth`

NewGitIntegrationAuth instantiates a new GitIntegrationAuth object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitIntegrationAuthWithDefaults

`func NewGitIntegrationAuthWithDefaults() *GitIntegrationAuth`

NewGitIntegrationAuthWithDefaults instantiates a new GitIntegrationAuth object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetToken

`func (o *GitIntegrationAuth) GetToken() string`

GetToken returns the Token field if non-nil, zero value otherwise.

### GetTokenOk

`func (o *GitIntegrationAuth) GetTokenOk() (*string, bool)`

GetTokenOk returns a tuple with the Token field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetToken

`func (o *GitIntegrationAuth) SetToken(v string)`

SetToken sets Token field to given value.

### HasToken

`func (o *GitIntegrationAuth) HasToken() bool`

HasToken returns a boolean if a field has been set.

### GetBasic

`func (o *GitIntegrationAuth) GetBasic() GitIntegrationBasicAuth`

GetBasic returns the Basic field if non-nil, zero value otherwise.

### GetBasicOk

`func (o *GitIntegrationAuth) GetBasicOk() (*GitIntegrationBasicAuth, bool)`

GetBasicOk returns a tuple with the Basic field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBasic

`func (o *GitIntegrationAuth) SetBasic(v GitIntegrationBasicAuth)`

SetBasic sets Basic field to given value.

### HasBasic

`func (o *GitIntegrationAuth) HasBasic() bool`

HasBasic returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


