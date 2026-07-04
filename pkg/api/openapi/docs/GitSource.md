# GitSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RepoUrl** | **string** |  | 
**Branch** | Pointer to **string** | Defaults to the repository&#39;s default branch, resolved and stored at create time | [optional] 
**Tag** | Pointer to **string** | Mutually exclusive with branch | [optional] 
**Commit** | Pointer to **string** | Commit SHA pin; requires branch or tag | [optional] 
**DockerfilePath** | Pointer to **string** |  | [optional] [default to "Dockerfile"]
**BuildContext** | Pointer to **string** |  | [optional] [default to "."]
**IntegrationId** | Pointer to **string** | Org-level git integration override for clone auth | [optional] 
**Credentials** | Pointer to [**InlineCredentials**](InlineCredentials.md) |  | [optional] 
**CredentialsConfigured** | Pointer to **bool** | True when clone credentials are configured for this source | [optional] [readonly] 
**Push** | Pointer to [**PushTarget**](PushTarget.md) |  | [optional] 

## Methods

### NewGitSource

`func NewGitSource(repoUrl string, ) *GitSource`

NewGitSource instantiates a new GitSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitSourceWithDefaults

`func NewGitSourceWithDefaults() *GitSource`

NewGitSourceWithDefaults instantiates a new GitSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepoUrl

`func (o *GitSource) GetRepoUrl() string`

GetRepoUrl returns the RepoUrl field if non-nil, zero value otherwise.

### GetRepoUrlOk

`func (o *GitSource) GetRepoUrlOk() (*string, bool)`

GetRepoUrlOk returns a tuple with the RepoUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepoUrl

`func (o *GitSource) SetRepoUrl(v string)`

SetRepoUrl sets RepoUrl field to given value.


### GetBranch

`func (o *GitSource) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *GitSource) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *GitSource) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *GitSource) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### GetTag

`func (o *GitSource) GetTag() string`

GetTag returns the Tag field if non-nil, zero value otherwise.

### GetTagOk

`func (o *GitSource) GetTagOk() (*string, bool)`

GetTagOk returns a tuple with the Tag field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTag

`func (o *GitSource) SetTag(v string)`

SetTag sets Tag field to given value.

### HasTag

`func (o *GitSource) HasTag() bool`

HasTag returns a boolean if a field has been set.

### GetCommit

`func (o *GitSource) GetCommit() string`

GetCommit returns the Commit field if non-nil, zero value otherwise.

### GetCommitOk

`func (o *GitSource) GetCommitOk() (*string, bool)`

GetCommitOk returns a tuple with the Commit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommit

`func (o *GitSource) SetCommit(v string)`

SetCommit sets Commit field to given value.

### HasCommit

`func (o *GitSource) HasCommit() bool`

HasCommit returns a boolean if a field has been set.

### GetDockerfilePath

`func (o *GitSource) GetDockerfilePath() string`

GetDockerfilePath returns the DockerfilePath field if non-nil, zero value otherwise.

### GetDockerfilePathOk

`func (o *GitSource) GetDockerfilePathOk() (*string, bool)`

GetDockerfilePathOk returns a tuple with the DockerfilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDockerfilePath

`func (o *GitSource) SetDockerfilePath(v string)`

SetDockerfilePath sets DockerfilePath field to given value.

### HasDockerfilePath

`func (o *GitSource) HasDockerfilePath() bool`

HasDockerfilePath returns a boolean if a field has been set.

### GetBuildContext

`func (o *GitSource) GetBuildContext() string`

GetBuildContext returns the BuildContext field if non-nil, zero value otherwise.

### GetBuildContextOk

`func (o *GitSource) GetBuildContextOk() (*string, bool)`

GetBuildContextOk returns a tuple with the BuildContext field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildContext

`func (o *GitSource) SetBuildContext(v string)`

SetBuildContext sets BuildContext field to given value.

### HasBuildContext

`func (o *GitSource) HasBuildContext() bool`

HasBuildContext returns a boolean if a field has been set.

### GetIntegrationId

`func (o *GitSource) GetIntegrationId() string`

GetIntegrationId returns the IntegrationId field if non-nil, zero value otherwise.

### GetIntegrationIdOk

`func (o *GitSource) GetIntegrationIdOk() (*string, bool)`

GetIntegrationIdOk returns a tuple with the IntegrationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIntegrationId

`func (o *GitSource) SetIntegrationId(v string)`

SetIntegrationId sets IntegrationId field to given value.

### HasIntegrationId

`func (o *GitSource) HasIntegrationId() bool`

HasIntegrationId returns a boolean if a field has been set.

### GetCredentials

`func (o *GitSource) GetCredentials() InlineCredentials`

GetCredentials returns the Credentials field if non-nil, zero value otherwise.

### GetCredentialsOk

`func (o *GitSource) GetCredentialsOk() (*InlineCredentials, bool)`

GetCredentialsOk returns a tuple with the Credentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentials

`func (o *GitSource) SetCredentials(v InlineCredentials)`

SetCredentials sets Credentials field to given value.

### HasCredentials

`func (o *GitSource) HasCredentials() bool`

HasCredentials returns a boolean if a field has been set.

### GetCredentialsConfigured

`func (o *GitSource) GetCredentialsConfigured() bool`

GetCredentialsConfigured returns the CredentialsConfigured field if non-nil, zero value otherwise.

### GetCredentialsConfiguredOk

`func (o *GitSource) GetCredentialsConfiguredOk() (*bool, bool)`

GetCredentialsConfiguredOk returns a tuple with the CredentialsConfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialsConfigured

`func (o *GitSource) SetCredentialsConfigured(v bool)`

SetCredentialsConfigured sets CredentialsConfigured field to given value.

### HasCredentialsConfigured

`func (o *GitSource) HasCredentialsConfigured() bool`

HasCredentialsConfigured returns a boolean if a field has been set.

### GetPush

`func (o *GitSource) GetPush() PushTarget`

GetPush returns the Push field if non-nil, zero value otherwise.

### GetPushOk

`func (o *GitSource) GetPushOk() (*PushTarget, bool)`

GetPushOk returns a tuple with the Push field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPush

`func (o *GitSource) SetPush(v PushTarget)`

SetPush sets Push field to given value.

### HasPush

`func (o *GitSource) HasPush() bool`

HasPush returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


