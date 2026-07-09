# GitRepository

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FullName** | Pointer to **string** |  | [optional] 
**CloneUrl** | Pointer to **string** |  | [optional] 
**DefaultBranch** | Pointer to **string** |  | [optional] 
**Private** | Pointer to **bool** |  | [optional] 
**PushedAt** | Pointer to **time.Time** |  | [optional] 
**Owner** | Pointer to **string** |  | [optional] 

## Methods

### NewGitRepository

`func NewGitRepository() *GitRepository`

NewGitRepository instantiates a new GitRepository object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitRepositoryWithDefaults

`func NewGitRepositoryWithDefaults() *GitRepository`

NewGitRepositoryWithDefaults instantiates a new GitRepository object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFullName

`func (o *GitRepository) GetFullName() string`

GetFullName returns the FullName field if non-nil, zero value otherwise.

### GetFullNameOk

`func (o *GitRepository) GetFullNameOk() (*string, bool)`

GetFullNameOk returns a tuple with the FullName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullName

`func (o *GitRepository) SetFullName(v string)`

SetFullName sets FullName field to given value.

### HasFullName

`func (o *GitRepository) HasFullName() bool`

HasFullName returns a boolean if a field has been set.

### GetCloneUrl

`func (o *GitRepository) GetCloneUrl() string`

GetCloneUrl returns the CloneUrl field if non-nil, zero value otherwise.

### GetCloneUrlOk

`func (o *GitRepository) GetCloneUrlOk() (*string, bool)`

GetCloneUrlOk returns a tuple with the CloneUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCloneUrl

`func (o *GitRepository) SetCloneUrl(v string)`

SetCloneUrl sets CloneUrl field to given value.

### HasCloneUrl

`func (o *GitRepository) HasCloneUrl() bool`

HasCloneUrl returns a boolean if a field has been set.

### GetDefaultBranch

`func (o *GitRepository) GetDefaultBranch() string`

GetDefaultBranch returns the DefaultBranch field if non-nil, zero value otherwise.

### GetDefaultBranchOk

`func (o *GitRepository) GetDefaultBranchOk() (*string, bool)`

GetDefaultBranchOk returns a tuple with the DefaultBranch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultBranch

`func (o *GitRepository) SetDefaultBranch(v string)`

SetDefaultBranch sets DefaultBranch field to given value.

### HasDefaultBranch

`func (o *GitRepository) HasDefaultBranch() bool`

HasDefaultBranch returns a boolean if a field has been set.

### GetPrivate

`func (o *GitRepository) GetPrivate() bool`

GetPrivate returns the Private field if non-nil, zero value otherwise.

### GetPrivateOk

`func (o *GitRepository) GetPrivateOk() (*bool, bool)`

GetPrivateOk returns a tuple with the Private field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrivate

`func (o *GitRepository) SetPrivate(v bool)`

SetPrivate sets Private field to given value.

### HasPrivate

`func (o *GitRepository) HasPrivate() bool`

HasPrivate returns a boolean if a field has been set.

### GetPushedAt

`func (o *GitRepository) GetPushedAt() time.Time`

GetPushedAt returns the PushedAt field if non-nil, zero value otherwise.

### GetPushedAtOk

`func (o *GitRepository) GetPushedAtOk() (*time.Time, bool)`

GetPushedAtOk returns a tuple with the PushedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPushedAt

`func (o *GitRepository) SetPushedAt(v time.Time)`

SetPushedAt sets PushedAt field to given value.

### HasPushedAt

`func (o *GitRepository) HasPushedAt() bool`

HasPushedAt returns a boolean if a field has been set.

### GetOwner

`func (o *GitRepository) GetOwner() string`

GetOwner returns the Owner field if non-nil, zero value otherwise.

### GetOwnerOk

`func (o *GitRepository) GetOwnerOk() (*string, bool)`

GetOwnerOk returns a tuple with the Owner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOwner

`func (o *GitRepository) SetOwner(v string)`

SetOwner sets Owner field to given value.

### HasOwner

`func (o *GitRepository) HasOwner() bool`

HasOwner returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


