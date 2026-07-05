# GitInstallation

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**InstallationId** | Pointer to **int64** |  | [optional] 
**AccountLogin** | Pointer to **string** |  | [optional] 
**AccountType** | Pointer to **string** |  | [optional] 
**RepositorySelection** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewGitInstallation

`func NewGitInstallation() *GitInstallation`

NewGitInstallation instantiates a new GitInstallation object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitInstallationWithDefaults

`func NewGitInstallationWithDefaults() *GitInstallation`

NewGitInstallationWithDefaults instantiates a new GitInstallation object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *GitInstallation) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GitInstallation) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GitInstallation) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *GitInstallation) HasId() bool`

HasId returns a boolean if a field has been set.

### GetInstallationId

`func (o *GitInstallation) GetInstallationId() int64`

GetInstallationId returns the InstallationId field if non-nil, zero value otherwise.

### GetInstallationIdOk

`func (o *GitInstallation) GetInstallationIdOk() (*int64, bool)`

GetInstallationIdOk returns a tuple with the InstallationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstallationId

`func (o *GitInstallation) SetInstallationId(v int64)`

SetInstallationId sets InstallationId field to given value.

### HasInstallationId

`func (o *GitInstallation) HasInstallationId() bool`

HasInstallationId returns a boolean if a field has been set.

### GetAccountLogin

`func (o *GitInstallation) GetAccountLogin() string`

GetAccountLogin returns the AccountLogin field if non-nil, zero value otherwise.

### GetAccountLoginOk

`func (o *GitInstallation) GetAccountLoginOk() (*string, bool)`

GetAccountLoginOk returns a tuple with the AccountLogin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccountLogin

`func (o *GitInstallation) SetAccountLogin(v string)`

SetAccountLogin sets AccountLogin field to given value.

### HasAccountLogin

`func (o *GitInstallation) HasAccountLogin() bool`

HasAccountLogin returns a boolean if a field has been set.

### GetAccountType

`func (o *GitInstallation) GetAccountType() string`

GetAccountType returns the AccountType field if non-nil, zero value otherwise.

### GetAccountTypeOk

`func (o *GitInstallation) GetAccountTypeOk() (*string, bool)`

GetAccountTypeOk returns a tuple with the AccountType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccountType

`func (o *GitInstallation) SetAccountType(v string)`

SetAccountType sets AccountType field to given value.

### HasAccountType

`func (o *GitInstallation) HasAccountType() bool`

HasAccountType returns a boolean if a field has been set.

### GetRepositorySelection

`func (o *GitInstallation) GetRepositorySelection() string`

GetRepositorySelection returns the RepositorySelection field if non-nil, zero value otherwise.

### GetRepositorySelectionOk

`func (o *GitInstallation) GetRepositorySelectionOk() (*string, bool)`

GetRepositorySelectionOk returns a tuple with the RepositorySelection field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepositorySelection

`func (o *GitInstallation) SetRepositorySelection(v string)`

SetRepositorySelection sets RepositorySelection field to given value.

### HasRepositorySelection

`func (o *GitInstallation) HasRepositorySelection() bool`

HasRepositorySelection returns a boolean if a field has been set.

### GetCreatedAt

`func (o *GitInstallation) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GitInstallation) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GitInstallation) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *GitInstallation) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


