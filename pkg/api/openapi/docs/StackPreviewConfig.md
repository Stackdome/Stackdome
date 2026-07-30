# StackPreviewConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**ProjectId** | Pointer to **string** |  | [optional] [readonly] 
**UserId** | Pointer to **string** |  | [optional] [readonly] 
**Name** | Pointer to **string** |  | [optional] 
**Description** | Pointer to **string** |  | [optional] 
**GitRepository** | Pointer to [**PreviewGitRepository**](PreviewGitRepository.md) |  | [optional] 
**StackfilePath** | Pointer to **string** |  | [optional] 
**MaxActivePreviews** | Pointer to **int32** |  | [optional] 
**Env** | Pointer to [**[]EnvVar**](EnvVar.md) | Env var overrides applied to every preview environment; may use secret references. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewStackPreviewConfig

`func NewStackPreviewConfig() *StackPreviewConfig`

NewStackPreviewConfig instantiates a new StackPreviewConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackPreviewConfigWithDefaults

`func NewStackPreviewConfigWithDefaults() *StackPreviewConfig`

NewStackPreviewConfigWithDefaults instantiates a new StackPreviewConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *StackPreviewConfig) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *StackPreviewConfig) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *StackPreviewConfig) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *StackPreviewConfig) HasId() bool`

HasId returns a boolean if a field has been set.

### GetOrganisationId

`func (o *StackPreviewConfig) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *StackPreviewConfig) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *StackPreviewConfig) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *StackPreviewConfig) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetProjectId

`func (o *StackPreviewConfig) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *StackPreviewConfig) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *StackPreviewConfig) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.

### HasProjectId

`func (o *StackPreviewConfig) HasProjectId() bool`

HasProjectId returns a boolean if a field has been set.

### GetUserId

`func (o *StackPreviewConfig) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *StackPreviewConfig) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *StackPreviewConfig) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *StackPreviewConfig) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetName

`func (o *StackPreviewConfig) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *StackPreviewConfig) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *StackPreviewConfig) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *StackPreviewConfig) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDescription

`func (o *StackPreviewConfig) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *StackPreviewConfig) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *StackPreviewConfig) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *StackPreviewConfig) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetGitRepository

`func (o *StackPreviewConfig) GetGitRepository() PreviewGitRepository`

GetGitRepository returns the GitRepository field if non-nil, zero value otherwise.

### GetGitRepositoryOk

`func (o *StackPreviewConfig) GetGitRepositoryOk() (*PreviewGitRepository, bool)`

GetGitRepositoryOk returns a tuple with the GitRepository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitRepository

`func (o *StackPreviewConfig) SetGitRepository(v PreviewGitRepository)`

SetGitRepository sets GitRepository field to given value.

### HasGitRepository

`func (o *StackPreviewConfig) HasGitRepository() bool`

HasGitRepository returns a boolean if a field has been set.

### GetStackfilePath

`func (o *StackPreviewConfig) GetStackfilePath() string`

GetStackfilePath returns the StackfilePath field if non-nil, zero value otherwise.

### GetStackfilePathOk

`func (o *StackPreviewConfig) GetStackfilePathOk() (*string, bool)`

GetStackfilePathOk returns a tuple with the StackfilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackfilePath

`func (o *StackPreviewConfig) SetStackfilePath(v string)`

SetStackfilePath sets StackfilePath field to given value.

### HasStackfilePath

`func (o *StackPreviewConfig) HasStackfilePath() bool`

HasStackfilePath returns a boolean if a field has been set.

### GetMaxActivePreviews

`func (o *StackPreviewConfig) GetMaxActivePreviews() int32`

GetMaxActivePreviews returns the MaxActivePreviews field if non-nil, zero value otherwise.

### GetMaxActivePreviewsOk

`func (o *StackPreviewConfig) GetMaxActivePreviewsOk() (*int32, bool)`

GetMaxActivePreviewsOk returns a tuple with the MaxActivePreviews field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxActivePreviews

`func (o *StackPreviewConfig) SetMaxActivePreviews(v int32)`

SetMaxActivePreviews sets MaxActivePreviews field to given value.

### HasMaxActivePreviews

`func (o *StackPreviewConfig) HasMaxActivePreviews() bool`

HasMaxActivePreviews returns a boolean if a field has been set.

### GetEnv

`func (o *StackPreviewConfig) GetEnv() []EnvVar`

GetEnv returns the Env field if non-nil, zero value otherwise.

### GetEnvOk

`func (o *StackPreviewConfig) GetEnvOk() (*[]EnvVar, bool)`

GetEnvOk returns a tuple with the Env field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnv

`func (o *StackPreviewConfig) SetEnv(v []EnvVar)`

SetEnv sets Env field to given value.

### HasEnv

`func (o *StackPreviewConfig) HasEnv() bool`

HasEnv returns a boolean if a field has been set.

### GetLabels

`func (o *StackPreviewConfig) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *StackPreviewConfig) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *StackPreviewConfig) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *StackPreviewConfig) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *StackPreviewConfig) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *StackPreviewConfig) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *StackPreviewConfig) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *StackPreviewConfig) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetCreatedAt

`func (o *StackPreviewConfig) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *StackPreviewConfig) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *StackPreviewConfig) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *StackPreviewConfig) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *StackPreviewConfig) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *StackPreviewConfig) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *StackPreviewConfig) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *StackPreviewConfig) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


