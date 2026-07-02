# PreviewGitRepository

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RepoUrl** | **string** |  | 
**BaseBranch** | Pointer to **string** |  | [optional] 
**GitSecretRef** | Pointer to **string** |  | [optional] 

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

### GetGitSecretRef

`func (o *PreviewGitRepository) GetGitSecretRef() string`

GetGitSecretRef returns the GitSecretRef field if non-nil, zero value otherwise.

### GetGitSecretRefOk

`func (o *PreviewGitRepository) GetGitSecretRefOk() (*string, bool)`

GetGitSecretRefOk returns a tuple with the GitSecretRef field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitSecretRef

`func (o *PreviewGitRepository) SetGitSecretRef(v string)`

SetGitSecretRef sets GitSecretRef field to given value.

### HasGitSecretRef

`func (o *PreviewGitRepository) HasGitSecretRef() bool`

HasGitSecretRef returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


