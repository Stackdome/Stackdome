# PreviewGitRepository

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RepoUrl** | **string** |  | 
**BaseBranch** | Pointer to **string** |  | [optional] 
**IntegrationId** | Pointer to **string** | Org-level git integration override for clone auth | [optional] 
**Credentials** | Pointer to [**InlineCredentials**](InlineCredentials.md) |  | [optional] 
**CredentialsConfigured** | Pointer to **bool** |  | [optional] [readonly] 

## Methods

### NewPreviewGitRepository

`func NewPreviewGitRepository(repoUrl string, ) *PreviewGitRepository`

NewPreviewGitRepository instantiates a new PreviewGitRepository object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPreviewGitRepositoryWithDefaults

`func NewPreviewGitRepositoryWithDefaults() *PreviewGitRepository`

NewPreviewGitRepositoryWithDefaults instantiates a new PreviewGitRepository object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRepoUrl

`func (o *PreviewGitRepository) GetRepoUrl() string`

GetRepoUrl returns the RepoUrl field if non-nil, zero value otherwise.

### GetRepoUrlOk

`func (o *PreviewGitRepository) GetRepoUrlOk() (*string, bool)`

GetRepoUrlOk returns a tuple with the RepoUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRepoUrl

`func (o *PreviewGitRepository) SetRepoUrl(v string)`

SetRepoUrl sets RepoUrl field to given value.


### GetBaseBranch

`func (o *PreviewGitRepository) GetBaseBranch() string`

GetBaseBranch returns the BaseBranch field if non-nil, zero value otherwise.

### GetBaseBranchOk

`func (o *PreviewGitRepository) GetBaseBranchOk() (*string, bool)`

GetBaseBranchOk returns a tuple with the BaseBranch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBaseBranch

`func (o *PreviewGitRepository) SetBaseBranch(v string)`

SetBaseBranch sets BaseBranch field to given value.

### HasBaseBranch

`func (o *PreviewGitRepository) HasBaseBranch() bool`

HasBaseBranch returns a boolean if a field has been set.

### GetIntegrationId

`func (o *PreviewGitRepository) GetIntegrationId() string`

GetIntegrationId returns the IntegrationId field if non-nil, zero value otherwise.

### GetIntegrationIdOk

`func (o *PreviewGitRepository) GetIntegrationIdOk() (*string, bool)`

GetIntegrationIdOk returns a tuple with the IntegrationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIntegrationId

`func (o *PreviewGitRepository) SetIntegrationId(v string)`

SetIntegrationId sets IntegrationId field to given value.

### HasIntegrationId

`func (o *PreviewGitRepository) HasIntegrationId() bool`

HasIntegrationId returns a boolean if a field has been set.

### GetCredentials

`func (o *PreviewGitRepository) GetCredentials() InlineCredentials`

GetCredentials returns the Credentials field if non-nil, zero value otherwise.

### GetCredentialsOk

`func (o *PreviewGitRepository) GetCredentialsOk() (*InlineCredentials, bool)`

GetCredentialsOk returns a tuple with the Credentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentials

`func (o *PreviewGitRepository) SetCredentials(v InlineCredentials)`

SetCredentials sets Credentials field to given value.

### HasCredentials

`func (o *PreviewGitRepository) HasCredentials() bool`

HasCredentials returns a boolean if a field has been set.

### GetCredentialsConfigured

`func (o *PreviewGitRepository) GetCredentialsConfigured() bool`

GetCredentialsConfigured returns the CredentialsConfigured field if non-nil, zero value otherwise.

### GetCredentialsConfiguredOk

`func (o *PreviewGitRepository) GetCredentialsConfiguredOk() (*bool, bool)`

GetCredentialsConfiguredOk returns a tuple with the CredentialsConfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialsConfigured

`func (o *PreviewGitRepository) SetCredentialsConfigured(v bool)`

SetCredentialsConfigured sets CredentialsConfigured field to given value.

### HasCredentialsConfigured

`func (o *PreviewGitRepository) HasCredentialsConfigured() bool`

HasCredentialsConfigured returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


