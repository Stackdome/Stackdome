# GitRepoSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RepoUrl** | **string** |  | 
**Revision** | [**GitRepoRevision**](GitRepoRevision.md) |  | 

## Methods

### NewGitRepoSource

`func NewGitRepoSource(repoUrl string, revision GitRepoRevision, ) *GitRepoSource`

NewGitRepoSource instantiates a new GitRepoSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitRepoSourceWithDefaults

`func NewGitRepoSourceWithDefaults() *GitRepoSource`

NewGitRepoSourceWithDefaults instantiates a new GitRepoSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepoUrl

`func (o *GitRepoSource) GetRepoUrl() string`

GetRepoUrl returns the RepoUrl field if non-nil, zero value otherwise.

### GetRepoUrlOk

`func (o *GitRepoSource) GetRepoUrlOk() (*string, bool)`

GetRepoUrlOk returns a tuple with the RepoUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepoUrl

`func (o *GitRepoSource) SetRepoUrl(v string)`

SetRepoUrl sets RepoUrl field to given value.


### GetRevision

`func (o *GitRepoSource) GetRevision() GitRepoRevision`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *GitRepoSource) GetRevisionOk() (*GitRepoRevision, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *GitRepoSource) SetRevision(v GitRepoRevision)`

SetRevision sets Revision field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


