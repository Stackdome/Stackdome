# StackPreviewConfigCreate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | 
**GitRepository** | [**PreviewGitRepository**](PreviewGitRepository.md) |  | 
**Description** | Pointer to **string** |  | [optional] 
**StackfilePath** | Pointer to **string** |  | [optional] 
**MaxActivePreviews** | Pointer to **int32** |  | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 

## Methods

### NewStackPreviewConfigCreate

`func NewStackPreviewConfigCreate(name string, gitRepository PreviewGitRepository, ) *StackPreviewConfigCreate`

NewStackPreviewConfigCreate instantiates a new StackPreviewConfigCreate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackPreviewConfigCreateWithDefaults

`func NewStackPreviewConfigCreateWithDefaults() *StackPreviewConfigCreate`

NewStackPreviewConfigCreateWithDefaults instantiates a new StackPreviewConfigCreate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *StackPreviewConfigCreate) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *StackPreviewConfigCreate) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *StackPreviewConfigCreate) SetName(v string)`

SetName sets Name field to given value.


### GetGitRepository

`func (o *StackPreviewConfigCreate) GetGitRepository() PreviewGitRepository`

GetGitRepository returns the GitRepository field if non-nil, zero value otherwise.

### GetGitRepositoryOk

`func (o *StackPreviewConfigCreate) GetGitRepositoryOk() (*PreviewGitRepository, bool)`

GetGitRepositoryOk returns a tuple with the GitRepository field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGitRepository

`func (o *StackPreviewConfigCreate) SetGitRepository(v PreviewGitRepository)`

SetGitRepository sets GitRepository field to given value.


### GetDescription

`func (o *StackPreviewConfigCreate) GetDescription() string`

GetDescription returns the Description field if non-nil, zero value otherwise.

### GetDescriptionOk

`func (o *StackPreviewConfigCreate) GetDescriptionOk() (*string, bool)`

GetDescriptionOk returns a tuple with the Description field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDescription

`func (o *StackPreviewConfigCreate) SetDescription(v string)`

SetDescription sets Description field to given value.

### HasDescription

`func (o *StackPreviewConfigCreate) HasDescription() bool`

HasDescription returns a boolean if a field has been set.

### GetStackfilePath

`func (o *StackPreviewConfigCreate) GetStackfilePath() string`

GetStackfilePath returns the StackfilePath field if non-nil, zero value otherwise.

### GetStackfilePathOk

`func (o *StackPreviewConfigCreate) GetStackfilePathOk() (*string, bool)`

GetStackfilePathOk returns a tuple with the StackfilePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackfilePath

`func (o *StackPreviewConfigCreate) SetStackfilePath(v string)`

SetStackfilePath sets StackfilePath field to given value.

### HasStackfilePath

`func (o *StackPreviewConfigCreate) HasStackfilePath() bool`

HasStackfilePath returns a boolean if a field has been set.

### GetMaxActivePreviews

`func (o *StackPreviewConfigCreate) GetMaxActivePreviews() int32`

GetMaxActivePreviews returns the MaxActivePreviews field if non-nil, zero value otherwise.

### GetMaxActivePreviewsOk

`func (o *StackPreviewConfigCreate) GetMaxActivePreviewsOk() (*int32, bool)`

GetMaxActivePreviewsOk returns a tuple with the MaxActivePreviews field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxActivePreviews

`func (o *StackPreviewConfigCreate) SetMaxActivePreviews(v int32)`

SetMaxActivePreviews sets MaxActivePreviews field to given value.

### HasMaxActivePreviews

`func (o *StackPreviewConfigCreate) HasMaxActivePreviews() bool`

HasMaxActivePreviews returns a boolean if a field has been set.

### GetLabels

`func (o *StackPreviewConfigCreate) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *StackPreviewConfigCreate) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *StackPreviewConfigCreate) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *StackPreviewConfigCreate) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *StackPreviewConfigCreate) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *StackPreviewConfigCreate) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *StackPreviewConfigCreate) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *StackPreviewConfigCreate) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


