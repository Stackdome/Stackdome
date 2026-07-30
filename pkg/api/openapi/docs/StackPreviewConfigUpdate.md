# StackPreviewConfigUpdate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Description** | Pointer to **string** |  | [optional] 
**StackfilePath** | Pointer to **string** |  | [optional] 
**MaxActivePreviews** | Pointer to **int32** |  | [optional] 
**GitRepository** | Pointer to [**PreviewGitRepository**](PreviewGitRepository.md) |  | [optional] 
**Env** | Pointer to [**[]EnvVar**](EnvVar.md) | Env var overrides applied to every preview environment; may use secret references. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 

## Methods

### NewStackPreviewConfigUpdate

`func NewStackPreviewConfigUpdate() *StackPreviewConfigUpdate`

NewStackPreviewConfigUpdate instantiates a new StackPreviewConfigUpdate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackPreviewConfigUpdateWithDefaults

`func NewStackPreviewConfigUpdateWithDefaults() *StackPreviewConfigUpdate`

NewStackPreviewConfigUpdateWithDefaults instantiates a new StackPreviewConfigUpdate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDescription

`func (o *StackPreviewConfigUpdate) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *StackPreviewConfigUpdate) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *StackPreviewConfigUpdate) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *StackPreviewConfigUpdate) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetStackfilePath

`func (o *StackPreviewConfigUpdate) GetStackfilePath() string`

GetStackfilePath returns the StackfilePath field if non-nil, zero value otherwise.

### GetStackfilePathOk

`func (o *StackPreviewConfigUpdate) GetStackfilePathOk() (*string, bool)`

GetStackfilePathOk returns a tuple with the StackfilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackfilePath

`func (o *StackPreviewConfigUpdate) SetStackfilePath(v string)`

SetStackfilePath sets StackfilePath field to given value.

### HasStackfilePath

`func (o *StackPreviewConfigUpdate) HasStackfilePath() bool`

HasStackfilePath returns a boolean if a field has been set.

### GetMaxActivePreviews

`func (o *StackPreviewConfigUpdate) GetMaxActivePreviews() int32`

GetMaxActivePreviews returns the MaxActivePreviews field if non-nil, zero value otherwise.

### GetMaxActivePreviewsOk

`func (o *StackPreviewConfigUpdate) GetMaxActivePreviewsOk() (*int32, bool)`

GetMaxActivePreviewsOk returns a tuple with the MaxActivePreviews field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxActivePreviews

`func (o *StackPreviewConfigUpdate) SetMaxActivePreviews(v int32)`

SetMaxActivePreviews sets MaxActivePreviews field to given value.

### HasMaxActivePreviews

`func (o *StackPreviewConfigUpdate) HasMaxActivePreviews() bool`

HasMaxActivePreviews returns a boolean if a field has been set.

### GetGitRepository

`func (o *StackPreviewConfigUpdate) GetGitRepository() PreviewGitRepository`

GetGitRepository returns the GitRepository field if non-nil, zero value otherwise.

### GetGitRepositoryOk

`func (o *StackPreviewConfigUpdate) GetGitRepositoryOk() (*PreviewGitRepository, bool)`

GetGitRepositoryOk returns a tuple with the GitRepository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitRepository

`func (o *StackPreviewConfigUpdate) SetGitRepository(v PreviewGitRepository)`

SetGitRepository sets GitRepository field to given value.

### HasGitRepository

`func (o *StackPreviewConfigUpdate) HasGitRepository() bool`

HasGitRepository returns a boolean if a field has been set.

### GetEnv

`func (o *StackPreviewConfigUpdate) GetEnv() []EnvVar`

GetEnv returns the Env field if non-nil, zero value otherwise.

### GetEnvOk

`func (o *StackPreviewConfigUpdate) GetEnvOk() (*[]EnvVar, bool)`

GetEnvOk returns a tuple with the Env field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnv

`func (o *StackPreviewConfigUpdate) SetEnv(v []EnvVar)`

SetEnv sets Env field to given value.

### HasEnv

`func (o *StackPreviewConfigUpdate) HasEnv() bool`

HasEnv returns a boolean if a field has been set.

### GetLabels

`func (o *StackPreviewConfigUpdate) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *StackPreviewConfigUpdate) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *StackPreviewConfigUpdate) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *StackPreviewConfigUpdate) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *StackPreviewConfigUpdate) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *StackPreviewConfigUpdate) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *StackPreviewConfigUpdate) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *StackPreviewConfigUpdate) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


