# PreviewStack

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**TeamId** | Pointer to **string** |  | [optional] [readonly] 
**UserId** | Pointer to **string** |  | [optional] [readonly] 
**StackPreviewConfigId** | Pointer to **string** |  | [optional] 
**StackId** | Pointer to **string** |  | [optional] 
**Name** | Pointer to **string** |  | [optional] 
**PrNumber** | Pointer to **string** |  | [optional] 
**Branch** | Pointer to **string** |  | [optional] 
**Commit** | Pointer to **string** |  | [optional] [readonly] 
**Source** | Pointer to **string** |  | [optional] 
**Status** | Pointer to [**PreviewStackStatus**](PreviewStackStatus.md) |  | [optional] 
**ImageOverrides** | Pointer to **map[string]string** |  | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 
**DeletionTimestamp** | Pointer to **time.Time** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewPreviewStack

`func NewPreviewStack() *PreviewStack`

NewPreviewStack instantiates a new PreviewStack object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPreviewStackWithDefaults

`func NewPreviewStackWithDefaults() *PreviewStack`

NewPreviewStackWithDefaults instantiates a new PreviewStack object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *PreviewStack) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *PreviewStack) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *PreviewStack) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *PreviewStack) HasId() bool`

HasId returns a boolean if a field has been set.

### GetOrganisationId

`func (o *PreviewStack) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *PreviewStack) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *PreviewStack) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *PreviewStack) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetTeamId

`func (o *PreviewStack) GetTeamId() string`

GetTeamId returns the TeamId field if non-nil, zero value otherwise.

### GetTeamIdOk

`func (o *PreviewStack) GetTeamIdOk() (*string, bool)`

GetTeamIdOk returns a tuple with the TeamId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTeamId

`func (o *PreviewStack) SetTeamId(v string)`

SetTeamId sets TeamId field to given value.

### HasTeamId

`func (o *PreviewStack) HasTeamId() bool`

HasTeamId returns a boolean if a field has been set.

### GetUserId

`func (o *PreviewStack) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *PreviewStack) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *PreviewStack) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *PreviewStack) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetStackPreviewConfigId

`func (o *PreviewStack) GetStackPreviewConfigId() string`

GetStackPreviewConfigId returns the StackPreviewConfigId field if non-nil, zero value otherwise.

### GetStackPreviewConfigIdOk

`func (o *PreviewStack) GetStackPreviewConfigIdOk() (*string, bool)`

GetStackPreviewConfigIdOk returns a tuple with the StackPreviewConfigId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackPreviewConfigId

`func (o *PreviewStack) SetStackPreviewConfigId(v string)`

SetStackPreviewConfigId sets StackPreviewConfigId field to given value.

### HasStackPreviewConfigId

`func (o *PreviewStack) HasStackPreviewConfigId() bool`

HasStackPreviewConfigId returns a boolean if a field has been set.

### GetStackId

`func (o *PreviewStack) GetStackId() string`

GetStackId returns the StackId field if non-nil, zero value otherwise.

### GetStackIdOk

`func (o *PreviewStack) GetStackIdOk() (*string, bool)`

GetStackIdOk returns a tuple with the StackId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackId

`func (o *PreviewStack) SetStackId(v string)`

SetStackId sets StackId field to given value.

### HasStackId

`func (o *PreviewStack) HasStackId() bool`

HasStackId returns a boolean if a field has been set.

### GetName

`func (o *PreviewStack) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *PreviewStack) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *PreviewStack) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *PreviewStack) HasName() bool`

HasName returns a boolean if a field has been set.

### GetPrNumber

`func (o *PreviewStack) GetPrNumber() string`

GetPrNumber returns the PrNumber field if non-nil, zero value otherwise.

### GetPrNumberOk

`func (o *PreviewStack) GetPrNumberOk() (*string, bool)`

GetPrNumberOk returns a tuple with the PrNumber field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPrNumber

`func (o *PreviewStack) SetPrNumber(v string)`

SetPrNumber sets PrNumber field to given value.

### HasPrNumber

`func (o *PreviewStack) HasPrNumber() bool`

HasPrNumber returns a boolean if a field has been set.

### GetBranch

`func (o *PreviewStack) GetBranch() string`

GetBranch returns the Branch field if non-nil, zero value otherwise.

### GetBranchOk

`func (o *PreviewStack) GetBranchOk() (*string, bool)`

GetBranchOk returns a tuple with the Branch field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBranch

`func (o *PreviewStack) SetBranch(v string)`

SetBranch sets Branch field to given value.

### HasBranch

`func (o *PreviewStack) HasBranch() bool`

HasBranch returns a boolean if a field has been set.

### GetCommit

`func (o *PreviewStack) GetCommit() string`

GetCommit returns the Commit field if non-nil, zero value otherwise.

### GetCommitOk

`func (o *PreviewStack) GetCommitOk() (*string, bool)`

GetCommitOk returns a tuple with the Commit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommit

`func (o *PreviewStack) SetCommit(v string)`

SetCommit sets Commit field to given value.

### HasCommit

`func (o *PreviewStack) HasCommit() bool`

HasCommit returns a boolean if a field has been set.

### GetSource

`func (o *PreviewStack) GetSource() string`

GetSource returns the Source field if non-nil, zero value otherwise.

### GetSourceOk

`func (o *PreviewStack) GetSourceOk() (*string, bool)`

GetSourceOk returns a tuple with the Source field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSource

`func (o *PreviewStack) SetSource(v string)`

SetSource sets Source field to given value.

### HasSource

`func (o *PreviewStack) HasSource() bool`

HasSource returns a boolean if a field has been set.

### GetStatus

`func (o *PreviewStack) GetStatus() PreviewStackStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *PreviewStack) GetStatusOk() (*PreviewStackStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *PreviewStack) SetStatus(v PreviewStackStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *PreviewStack) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### GetImageOverrides

`func (o *PreviewStack) GetImageOverrides() map[string]string`

GetImageOverrides returns the ImageOverrides field if non-nil, zero value otherwise.

### GetImageOverridesOk

`func (o *PreviewStack) GetImageOverridesOk() (*map[string]string, bool)`

GetImageOverridesOk returns a tuple with the ImageOverrides field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageOverrides

`func (o *PreviewStack) SetImageOverrides(v map[string]string)`

SetImageOverrides sets ImageOverrides field to given value.

### HasImageOverrides

`func (o *PreviewStack) HasImageOverrides() bool`

HasImageOverrides returns a boolean if a field has been set.

### GetLabels

`func (o *PreviewStack) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *PreviewStack) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *PreviewStack) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *PreviewStack) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *PreviewStack) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *PreviewStack) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *PreviewStack) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *PreviewStack) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetDeletionTimestamp

`func (o *PreviewStack) GetDeletionTimestamp() time.Time`

GetDeletionTimestamp returns the DeletionTimestamp field if non-nil, zero value otherwise.

### GetDeletionTimestampOk

`func (o *PreviewStack) GetDeletionTimestampOk() (*time.Time, bool)`

GetDeletionTimestampOk returns a tuple with the DeletionTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeletionTimestamp

`func (o *PreviewStack) SetDeletionTimestamp(v time.Time)`

SetDeletionTimestamp sets DeletionTimestamp field to given value.

### HasDeletionTimestamp

`func (o *PreviewStack) HasDeletionTimestamp() bool`

HasDeletionTimestamp returns a boolean if a field has been set.

### GetCreatedAt

`func (o *PreviewStack) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *PreviewStack) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *PreviewStack) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *PreviewStack) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *PreviewStack) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *PreviewStack) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *PreviewStack) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *PreviewStack) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


