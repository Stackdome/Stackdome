# GitIntegration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**Type** | Pointer to [**GitIntegrationType**](GitIntegrationType.md) |  | [optional] [default to GIT_INTEGRATION_TYPE_CREDENTIALS]
**Host** | **string** | Git host this integration covers, e.g. gitlab.example.com | 
**Status** | Pointer to **string** |  | [optional] [readonly] 
**Auth** | Pointer to [**GitIntegrationAuth**](GitIntegrationAuth.md) |  | [optional] 
**CredentialsConfigured** | Pointer to **bool** |  | [optional] [readonly] 
**InstallUrl** | Pointer to **string** | GitHub install page for adding the app to more accounts (github_app only) | [optional] [readonly] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewGitIntegration

`func NewGitIntegration(host string, ) *GitIntegration`

NewGitIntegration instantiates a new GitIntegration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitIntegrationWithDefaults

`func NewGitIntegrationWithDefaults() *GitIntegration`

NewGitIntegrationWithDefaults instantiates a new GitIntegration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *GitIntegration) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *GitIntegration) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *GitIntegration) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *GitIntegration) HasId() bool`

HasId returns a boolean if a field has been set.

### GetType

`func (o *GitIntegration) GetType() GitIntegrationType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *GitIntegration) GetTypeOk() (*GitIntegrationType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *GitIntegration) SetType(v GitIntegrationType)`

SetType sets Type field to given value.

### HasType

`func (o *GitIntegration) HasType() bool`

HasType returns a boolean if a field has been set.

### GetHost

`func (o *GitIntegration) GetHost() string`

GetHost returns the Host field if non-nil, zero value otherwise.

### GetHostOk

`func (o *GitIntegration) GetHostOk() (*string, bool)`

GetHostOk returns a tuple with the Host field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHost

`func (o *GitIntegration) SetHost(v string)`

SetHost sets Host field to given value.


### GetStatus

`func (o *GitIntegration) GetStatus() string`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *GitIntegration) GetStatusOk() (*string, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *GitIntegration) SetStatus(v string)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *GitIntegration) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetAuth

`func (o *GitIntegration) GetAuth() GitIntegrationAuth`

GetAuth returns the Auth field if non-nil, zero value otherwise.

### GetAuthOk

`func (o *GitIntegration) GetAuthOk() (*GitIntegrationAuth, bool)`

GetAuthOk returns a tuple with the Auth field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuth

`func (o *GitIntegration) SetAuth(v GitIntegrationAuth)`

SetAuth sets Auth field to given value.

### HasAuth

`func (o *GitIntegration) HasAuth() bool`

HasAuth returns a boolean if a field has been set.

### GetCredentialsConfigured

`func (o *GitIntegration) GetCredentialsConfigured() bool`

GetCredentialsConfigured returns the CredentialsConfigured field if non-nil, zero value otherwise.

### GetCredentialsConfiguredOk

`func (o *GitIntegration) GetCredentialsConfiguredOk() (*bool, bool)`

GetCredentialsConfiguredOk returns a tuple with the CredentialsConfigured field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialsConfigured

`func (o *GitIntegration) SetCredentialsConfigured(v bool)`

SetCredentialsConfigured sets CredentialsConfigured field to given value.

### HasCredentialsConfigured

`func (o *GitIntegration) HasCredentialsConfigured() bool`

HasCredentialsConfigured returns a boolean if a field has been set.

### GetInstallUrl

`func (o *GitIntegration) GetInstallUrl() string`

GetInstallUrl returns the InstallUrl field if non-nil, zero value otherwise.

### GetInstallUrlOk

`func (o *GitIntegration) GetInstallUrlOk() (*string, bool)`

GetInstallUrlOk returns a tuple with the InstallUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInstallUrl

`func (o *GitIntegration) SetInstallUrl(v string)`

SetInstallUrl sets InstallUrl field to given value.

### HasInstallUrl

`func (o *GitIntegration) HasInstallUrl() bool`

HasInstallUrl returns a boolean if a field has been set.

### GetOrganisationId

`func (o *GitIntegration) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *GitIntegration) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *GitIntegration) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *GitIntegration) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetCreatedAt

`func (o *GitIntegration) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *GitIntegration) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *GitIntegration) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *GitIntegration) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *GitIntegration) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *GitIntegration) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *GitIntegration) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *GitIntegration) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


