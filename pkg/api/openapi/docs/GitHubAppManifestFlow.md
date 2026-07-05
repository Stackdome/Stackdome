# GitHubAppManifestFlow

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Manifest** | Pointer to **map[string]interface{}** | GitHub App manifest to POST to github_url as the manifest form field | [optional] 
**GithubUrl** | Pointer to **string** |  | [optional] 
**State** | Pointer to **string** |  | [optional] 

## Methods

### NewGitHubAppManifestFlow

`func NewGitHubAppManifestFlow() *GitHubAppManifestFlow`

NewGitHubAppManifestFlow instantiates a new GitHubAppManifestFlow object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitHubAppManifestFlowWithDefaults

`func NewGitHubAppManifestFlowWithDefaults() *GitHubAppManifestFlow`

NewGitHubAppManifestFlowWithDefaults instantiates a new GitHubAppManifestFlow object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetManifest

`func (o *GitHubAppManifestFlow) GetManifest() map[string]interface{}`

GetManifest returns the Manifest field if non-nil, zero value otherwise.

### GetManifestOk

`func (o *GitHubAppManifestFlow) GetManifestOk() (*map[string]interface{}, bool)`

GetManifestOk returns a tuple with the Manifest field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetManifest

`func (o *GitHubAppManifestFlow) SetManifest(v map[string]interface{})`

SetManifest sets Manifest field to given value.

### HasManifest

`func (o *GitHubAppManifestFlow) HasManifest() bool`

HasManifest returns a boolean if a field has been set.

### GetGithubUrl

`func (o *GitHubAppManifestFlow) GetGithubUrl() string`

GetGithubUrl returns the GithubUrl field if non-nil, zero value otherwise.

### GetGithubUrlOk

`func (o *GitHubAppManifestFlow) GetGithubUrlOk() (*string, bool)`

GetGithubUrlOk returns a tuple with the GithubUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGithubUrl

`func (o *GitHubAppManifestFlow) SetGithubUrl(v string)`

SetGithubUrl sets GithubUrl field to given value.

### HasGithubUrl

`func (o *GitHubAppManifestFlow) HasGithubUrl() bool`

HasGithubUrl returns a boolean if a field has been set.

### GetState

`func (o *GitHubAppManifestFlow) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *GitHubAppManifestFlow) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *GitHubAppManifestFlow) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *GitHubAppManifestFlow) HasState() bool`

HasState returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


