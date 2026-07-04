# InlineCredentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Username** | **string** |  | 
**Password** | **string** |  | 
**SaveToOrg** | Pointer to **bool** | Also save these credentials as an org-level credential | [optional] 

## Methods

### NewInlineCredentials

`func NewInlineCredentials(username string, password string, ) *InlineCredentials`

NewInlineCredentials instantiates a new InlineCredentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInlineCredentialsWithDefaults

`func NewInlineCredentialsWithDefaults() *InlineCredentials`

NewInlineCredentialsWithDefaults instantiates a new InlineCredentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetUsername

`func (o *InlineCredentials) GetUsername() string`

GetUsername returns the Username field if non-nil, zero value otherwise.

### GetUsernameOk

`func (o *InlineCredentials) GetUsernameOk() (*string, bool)`

GetUsernameOk returns a tuple with the Username field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUsername

`func (o *InlineCredentials) SetUsername(v string)`

SetUsername sets Username field to given value.


### GetPassword

`func (o *InlineCredentials) GetPassword() string`

GetPassword returns the Password field if non-nil, zero value otherwise.

### GetPasswordOk

`func (o *InlineCredentials) GetPasswordOk() (*string, bool)`

GetPasswordOk returns a tuple with the Password field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPassword

`func (o *InlineCredentials) SetPassword(v string)`

SetPassword sets Password field to given value.


### GetSaveToOrg

`func (o *InlineCredentials) GetSaveToOrg() bool`

GetSaveToOrg returns the SaveToOrg field if non-nil, zero value otherwise.

### GetSaveToOrgOk

`func (o *InlineCredentials) GetSaveToOrgOk() (*bool, bool)`

GetSaveToOrgOk returns a tuple with the SaveToOrg field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSaveToOrg

`func (o *InlineCredentials) SetSaveToOrg(v bool)`

SetSaveToOrg sets SaveToOrg field to given value.

### HasSaveToOrg

`func (o *InlineCredentials) HasSaveToOrg() bool`

HasSaveToOrg returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


