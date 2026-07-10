# Stack

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**OrganisationId** | Pointer to **string** |  | [optional] [readonly] 
**TeamId** | Pointer to **string** |  | [optional] [readonly] 
**UserId** | Pointer to **string** |  | [optional] [readonly] 
**Name** | **string** |  | 
**Namespace** | Pointer to **string** |  | [optional] [readonly] 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 
**Revision** | Pointer to **string** |  | [optional] [readonly] 
**Spec** | [**StackSpec**](StackSpec.md) |  | 
**Settings** | Pointer to [**StackSettings**](StackSettings.md) |  | [optional] 
**Lifecycle** | Pointer to [**NullableStackLifecycle**](StackLifecycle.md) |  | [optional] [readonly] 
**CurrentRelease** | Pointer to [**StackCurrentRelease**](StackCurrentRelease.md) |  | [optional] 
**LatestRelease** | Pointer to [**StackCurrentRelease**](StackCurrentRelease.md) |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] [readonly] 

## Methods

### NewStack

`func NewStack(name string, spec StackSpec, ) *Stack`

NewStack instantiates a new Stack object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackWithDefaults

`func NewStackWithDefaults() *Stack`

NewStackWithDefaults instantiates a new Stack object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *Stack) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *Stack) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *Stack) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *Stack) HasId() bool`

HasId returns a boolean if a field has been set.

### GetOrganisationId

`func (o *Stack) GetOrganisationId() string`

GetOrganisationId returns the OrganisationId field if non-nil, zero value otherwise.

### GetOrganisationIdOk

`func (o *Stack) GetOrganisationIdOk() (*string, bool)`

GetOrganisationIdOk returns a tuple with the OrganisationId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOrganisationId

`func (o *Stack) SetOrganisationId(v string)`

SetOrganisationId sets OrganisationId field to given value.

### HasOrganisationId

`func (o *Stack) HasOrganisationId() bool`

HasOrganisationId returns a boolean if a field has been set.

### GetTeamId

`func (o *Stack) GetTeamId() string`

GetTeamId returns the TeamId field if non-nil, zero value otherwise.

### GetTeamIdOk

`func (o *Stack) GetTeamIdOk() (*string, bool)`

GetTeamIdOk returns a tuple with the TeamId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTeamId

`func (o *Stack) SetTeamId(v string)`

SetTeamId sets TeamId field to given value.

### HasTeamId

`func (o *Stack) HasTeamId() bool`

HasTeamId returns a boolean if a field has been set.

### GetUserId

`func (o *Stack) GetUserId() string`

GetUserId returns the UserId field if non-nil, zero value otherwise.

### GetUserIdOk

`func (o *Stack) GetUserIdOk() (*string, bool)`

GetUserIdOk returns a tuple with the UserId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUserId

`func (o *Stack) SetUserId(v string)`

SetUserId sets UserId field to given value.

### HasUserId

`func (o *Stack) HasUserId() bool`

HasUserId returns a boolean if a field has been set.

### GetName

`func (o *Stack) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *Stack) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *Stack) SetName(v string)`

SetName sets Name field to given value.


### GetNamespace

`func (o *Stack) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *Stack) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *Stack) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *Stack) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetLabels

`func (o *Stack) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *Stack) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *Stack) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *Stack) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *Stack) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *Stack) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *Stack) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *Stack) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetRevision

`func (o *Stack) GetRevision() string`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *Stack) GetRevisionOk() (*string, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *Stack) SetRevision(v string)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *Stack) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### GetSpec

`func (o *Stack) GetSpec() StackSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *Stack) GetSpecOk() (*StackSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *Stack) SetSpec(v StackSpec)`

SetSpec sets Spec field to given value.


### GetSettings

`func (o *Stack) GetSettings() StackSettings`

GetSettings returns the Settings field if non-nil, zero value otherwise.

### GetSettingsOk

`func (o *Stack) GetSettingsOk() (*StackSettings, bool)`

GetSettingsOk returns a tuple with the Settings field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSettings

`func (o *Stack) SetSettings(v StackSettings)`

SetSettings sets Settings field to given value.

### HasSettings

`func (o *Stack) HasSettings() bool`

HasSettings returns a boolean if a field has been set.

### GetLifecycle

`func (o *Stack) GetLifecycle() StackLifecycle`

GetLifecycle returns the Lifecycle field if non-nil, zero value otherwise.

### GetLifecycleOk

`func (o *Stack) GetLifecycleOk() (*StackLifecycle, bool)`

GetLifecycleOk returns a tuple with the Lifecycle field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLifecycle

`func (o *Stack) SetLifecycle(v StackLifecycle)`

SetLifecycle sets Lifecycle field to given value.

### HasLifecycle

`func (o *Stack) HasLifecycle() bool`

HasLifecycle returns a boolean if a field has been set.

### SetLifecycleNil

`func (o *Stack) SetLifecycleNil(b bool)`

 SetLifecycleNil sets the value for Lifecycle to be an explicit nil

### UnsetLifecycle
`func (o *Stack) UnsetLifecycle()`

UnsetLifecycle ensures that no value is present for Lifecycle, not even an explicit nil
### GetCurrentRelease

`func (o *Stack) GetCurrentRelease() StackCurrentRelease`

GetCurrentRelease returns the CurrentRelease field if non-nil, zero value otherwise.

### GetCurrentReleaseOk

`func (o *Stack) GetCurrentReleaseOk() (*StackCurrentRelease, bool)`

GetCurrentReleaseOk returns a tuple with the CurrentRelease field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCurrentRelease

`func (o *Stack) SetCurrentRelease(v StackCurrentRelease)`

SetCurrentRelease sets CurrentRelease field to given value.

### HasCurrentRelease

`func (o *Stack) HasCurrentRelease() bool`

HasCurrentRelease returns a boolean if a field has been set.

### GetLatestRelease

`func (o *Stack) GetLatestRelease() StackCurrentRelease`

GetLatestRelease returns the LatestRelease field if non-nil, zero value otherwise.

### GetLatestReleaseOk

`func (o *Stack) GetLatestReleaseOk() (*StackCurrentRelease, bool)`

GetLatestReleaseOk returns a tuple with the LatestRelease field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLatestRelease

`func (o *Stack) SetLatestRelease(v StackCurrentRelease)`

SetLatestRelease sets LatestRelease field to given value.

### HasLatestRelease

`func (o *Stack) HasLatestRelease() bool`

HasLatestRelease returns a boolean if a field has been set.

### GetCreatedAt

`func (o *Stack) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *Stack) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *Stack) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *Stack) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *Stack) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *Stack) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *Stack) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *Stack) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


