# StackRelease

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**StackId** | Pointer to **string** |  | [optional] 
**Sequence** | Pointer to **int32** |  | [optional] 
**State** | Pointer to [**StackReleaseState**](StackReleaseState.md) |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Cause** | Pointer to [**ReleaseCause**](ReleaseCause.md) |  | [optional] 
**SnapshotRevision** | Pointer to **string** |  | [optional] 
**ManifestRevision** | Pointer to **string** |  | [optional] 
**RendererVersion** | Pointer to **string** |  | [optional] 
**Pins** | Pointer to [**ReleasePins**](ReleasePins.md) |  | [optional] 
**Outcome** | Pointer to [**ReleaseOutcome**](ReleaseOutcome.md) |  | [optional] 
**CreatedBy** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**UpdatedAt** | Pointer to **time.Time** |  | [optional] 
**RenderedAt** | Pointer to **time.Time** |  | [optional] 
**CompletedAt** | Pointer to **time.Time** |  | [optional] 
**ValidationErrors** | Pointer to [**[]ReleaseValidationError**](ReleaseValidationError.md) |  | [optional] [readonly] 
**LiveStatus** | Pointer to [**ReleaseLiveStatus**](ReleaseLiveStatus.md) |  | [optional] 

## Methods

### NewStackRelease

`func NewStackRelease() *StackRelease`

NewStackRelease instantiates a new StackRelease object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackReleaseWithDefaults

`func NewStackReleaseWithDefaults() *StackRelease`

NewStackReleaseWithDefaults instantiates a new StackRelease object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *StackRelease) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *StackRelease) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *StackRelease) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *StackRelease) HasId() bool`

HasId returns a boolean if a field has been set.

### GetStackId

`func (o *StackRelease) GetStackId() string`

GetStackId returns the StackId field if non-nil, zero value otherwise.

### GetStackIdOk

`func (o *StackRelease) GetStackIdOk() (*string, bool)`

GetStackIdOk returns a tuple with the StackId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackId

`func (o *StackRelease) SetStackId(v string)`

SetStackId sets StackId field to given value.

### HasStackId

`func (o *StackRelease) HasStackId() bool`

HasStackId returns a boolean if a field has been set.

### GetSequence

`func (o *StackRelease) GetSequence() int32`

GetSequence returns the Sequence field if non-nil, zero value otherwise.

### GetSequenceOk

`func (o *StackRelease) GetSequenceOk() (*int32, bool)`

GetSequenceOk returns a tuple with the Sequence field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSequence

`func (o *StackRelease) SetSequence(v int32)`

SetSequence sets Sequence field to given value.

### HasSequence

`func (o *StackRelease) HasSequence() bool`

HasSequence returns a boolean if a field has been set.

### GetState

`func (o *StackRelease) GetState() StackReleaseState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *StackRelease) GetStateOk() (*StackReleaseState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *StackRelease) SetState(v StackReleaseState)`

SetState sets State field to given value.

### HasState

`func (o *StackRelease) HasState() bool`

HasState returns a boolean if a field has been set.

### GetMessage

`func (o *StackRelease) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *StackRelease) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *StackRelease) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *StackRelease) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetCause

`func (o *StackRelease) GetCause() ReleaseCause`

GetCause returns the Cause field if non-nil, zero value otherwise.

### GetCauseOk

`func (o *StackRelease) GetCauseOk() (*ReleaseCause, bool)`

GetCauseOk returns a tuple with the Cause field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCause

`func (o *StackRelease) SetCause(v ReleaseCause)`

SetCause sets Cause field to given value.

### HasCause

`func (o *StackRelease) HasCause() bool`

HasCause returns a boolean if a field has been set.

### GetSnapshotRevision

`func (o *StackRelease) GetSnapshotRevision() string`

GetSnapshotRevision returns the SnapshotRevision field if non-nil, zero value otherwise.

### GetSnapshotRevisionOk

`func (o *StackRelease) GetSnapshotRevisionOk() (*string, bool)`

GetSnapshotRevisionOk returns a tuple with the SnapshotRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSnapshotRevision

`func (o *StackRelease) SetSnapshotRevision(v string)`

SetSnapshotRevision sets SnapshotRevision field to given value.

### HasSnapshotRevision

`func (o *StackRelease) HasSnapshotRevision() bool`

HasSnapshotRevision returns a boolean if a field has been set.

### GetManifestRevision

`func (o *StackRelease) GetManifestRevision() string`

GetManifestRevision returns the ManifestRevision field if non-nil, zero value otherwise.

### GetManifestRevisionOk

`func (o *StackRelease) GetManifestRevisionOk() (*string, bool)`

GetManifestRevisionOk returns a tuple with the ManifestRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetManifestRevision

`func (o *StackRelease) SetManifestRevision(v string)`

SetManifestRevision sets ManifestRevision field to given value.

### HasManifestRevision

`func (o *StackRelease) HasManifestRevision() bool`

HasManifestRevision returns a boolean if a field has been set.

### GetRendererVersion

`func (o *StackRelease) GetRendererVersion() string`

GetRendererVersion returns the RendererVersion field if non-nil, zero value otherwise.

### GetRendererVersionOk

`func (o *StackRelease) GetRendererVersionOk() (*string, bool)`

GetRendererVersionOk returns a tuple with the RendererVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRendererVersion

`func (o *StackRelease) SetRendererVersion(v string)`

SetRendererVersion sets RendererVersion field to given value.

### HasRendererVersion

`func (o *StackRelease) HasRendererVersion() bool`

HasRendererVersion returns a boolean if a field has been set.

### GetPins

`func (o *StackRelease) GetPins() ReleasePins`

GetPins returns the Pins field if non-nil, zero value otherwise.

### GetPinsOk

`func (o *StackRelease) GetPinsOk() (*ReleasePins, bool)`

GetPinsOk returns a tuple with the Pins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPins

`func (o *StackRelease) SetPins(v ReleasePins)`

SetPins sets Pins field to given value.

### HasPins

`func (o *StackRelease) HasPins() bool`

HasPins returns a boolean if a field has been set.

### GetOutcome

`func (o *StackRelease) GetOutcome() ReleaseOutcome`

GetOutcome returns the Outcome field if non-nil, zero value otherwise.

### GetOutcomeOk

`func (o *StackRelease) GetOutcomeOk() (*ReleaseOutcome, bool)`

GetOutcomeOk returns a tuple with the Outcome field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutcome

`func (o *StackRelease) SetOutcome(v ReleaseOutcome)`

SetOutcome sets Outcome field to given value.

### HasOutcome

`func (o *StackRelease) HasOutcome() bool`

HasOutcome returns a boolean if a field has been set.

### GetCreatedBy

`func (o *StackRelease) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *StackRelease) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *StackRelease) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *StackRelease) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetCreatedAt

`func (o *StackRelease) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *StackRelease) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *StackRelease) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *StackRelease) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *StackRelease) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *StackRelease) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *StackRelease) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *StackRelease) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetRenderedAt

`func (o *StackRelease) GetRenderedAt() time.Time`

GetRenderedAt returns the RenderedAt field if non-nil, zero value otherwise.

### GetRenderedAtOk

`func (o *StackRelease) GetRenderedAtOk() (*time.Time, bool)`

GetRenderedAtOk returns a tuple with the RenderedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRenderedAt

`func (o *StackRelease) SetRenderedAt(v time.Time)`

SetRenderedAt sets RenderedAt field to given value.

### HasRenderedAt

`func (o *StackRelease) HasRenderedAt() bool`

HasRenderedAt returns a boolean if a field has been set.

### GetCompletedAt

`func (o *StackRelease) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *StackRelease) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *StackRelease) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *StackRelease) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.

### GetValidationErrors

`func (o *StackRelease) GetValidationErrors() []ReleaseValidationError`

GetValidationErrors returns the ValidationErrors field if non-nil, zero value otherwise.

### GetValidationErrorsOk

`func (o *StackRelease) GetValidationErrorsOk() (*[]ReleaseValidationError, bool)`

GetValidationErrorsOk returns a tuple with the ValidationErrors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidationErrors

`func (o *StackRelease) SetValidationErrors(v []ReleaseValidationError)`

SetValidationErrors sets ValidationErrors field to given value.

### HasValidationErrors

`func (o *StackRelease) HasValidationErrors() bool`

HasValidationErrors returns a boolean if a field has been set.

### GetLiveStatus

`func (o *StackRelease) GetLiveStatus() ReleaseLiveStatus`

GetLiveStatus returns the LiveStatus field if non-nil, zero value otherwise.

### GetLiveStatusOk

`func (o *StackRelease) GetLiveStatusOk() (*ReleaseLiveStatus, bool)`

GetLiveStatusOk returns a tuple with the LiveStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLiveStatus

`func (o *StackRelease) SetLiveStatus(v ReleaseLiveStatus)`

SetLiveStatus sets LiveStatus field to given value.

### HasLiveStatus

`func (o *StackRelease) HasLiveStatus() bool`

HasLiveStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


