# GitRepoRevision

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Branch** | Pointer to [**GitRepoRevisionBranch**](GitRepoRevisionBranch.md) |  | [optional] 
**Commit** | Pointer to **string** |  | [optional] 
**Tag** | Pointer to **string** |  | [optional] 

## Methods

### NewGitRepoRevision

`func NewGitRepoRevision() *GitRepoRevision`

NewGitRepoRevision instantiates a new GitRepoRevision object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitRepoRevisionWithDefaults

`func NewGitRepoRevisionWithDefaults() *GitRepoRevision`

NewGitRepoRevisionWithDefaults instantiates a new GitRepoRevision object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetBranch

`func (o *GitRepoRevision) GetBranch() GitRepoRevisionBranch`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GitRepoRevision) GetBranchOk() (*GitRepoRevisionBranch, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GitRepoRevision) SetBranch(v GitRepoRevisionBranch)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *GitRepoRevision) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### GetCommit

`func (o *GitRepoRevision) GetCommit() string`

GetCommit returns the Commit field if non-nil, zero value otherwise.

### GetCommitOk

`func (o *GitRepoRevision) GetCommitOk() (*string, bool)`

GetCommitOk returns a tuple with the Commit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommit

`func (o *GitRepoRevision) SetCommit(v string)`

SetCommit sets Commit field to given value.

### HasCommit

`func (o *GitRepoRevision) HasCommit() bool`

HasCommit returns a boolean if a field has been set.

### GetTag

`func (o *GitRepoRevision) GetTag() string`

GetTag returns the Tag field if non-nil, zero value otherwise.

### GetTagOk

`func (o *GitRepoRevision) GetTagOk() (*string, bool)`

GetTagOk returns a tuple with the Tag field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTag

`func (o *GitRepoRevision) SetTag(v string)`

SetTag sets Tag field to given value.

### HasTag

`func (o *GitRepoRevision) HasTag() bool`

HasTag returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


