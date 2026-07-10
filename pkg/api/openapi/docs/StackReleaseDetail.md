# StackReleaseDetail

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
**Snapshot** | Pointer to [**StackReleaseSnapshot**](StackReleaseSnapshot.md) |  | [optional] 

## Methods

### NewStackReleaseDetail

`func NewStackReleaseDetail() *StackReleaseDetail`

NewStackReleaseDetail instantiates a new StackReleaseDetail object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackReleaseDetailWithDefaults

`func NewStackReleaseDetailWithDefaults() *StackReleaseDetail`

NewStackReleaseDetailWithDefaults instantiates a new StackReleaseDetail object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *StackReleaseDetail) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *StackReleaseDetail) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *StackReleaseDetail) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *StackReleaseDetail) HasId() bool`

HasId returns a boolean if a field has been set.

### GetStackId

`func (o *StackReleaseDetail) GetStackId() string`

GetStackId returns the StackId field if non-nil, zero value otherwise.

### GetStackIdOk

`func (o *StackReleaseDetail) GetStackIdOk() (*string, bool)`

GetStackIdOk returns a tuple with the StackId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackId

`func (o *StackReleaseDetail) SetStackId(v string)`

SetStackId sets StackId field to given value.

### HasStackId

`func (o *StackReleaseDetail) HasStackId() bool`

HasStackId returns a boolean if a field has been set.

### GetSequence

`func (o *StackReleaseDetail) GetSequence() int32`

GetSequence returns the Sequence field if non-nil, zero value otherwise.

### GetSequenceOk

`func (o *StackReleaseDetail) GetSequenceOk() (*int32, bool)`

GetSequenceOk returns a tuple with the Sequence field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSequence

`func (o *StackReleaseDetail) SetSequence(v int32)`

SetSequence sets Sequence field to given value.

### HasSequence

`func (o *StackReleaseDetail) HasSequence() bool`

HasSequence returns a boolean if a field has been set.

### GetState

`func (o *StackReleaseDetail) GetState() StackReleaseState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *StackReleaseDetail) GetStateOk() (*StackReleaseState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *StackReleaseDetail) SetState(v StackReleaseState)`

SetState sets State field to given value.

### HasState

`func (o *StackReleaseDetail) HasState() bool`

HasState returns a boolean if a field has been set.

### GetMessage

`func (o *StackReleaseDetail) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *StackReleaseDetail) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *StackReleaseDetail) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *StackReleaseDetail) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetCause

`func (o *StackReleaseDetail) GetCause() ReleaseCause`

GetCause returns the Cause field if non-nil, zero value otherwise.

### GetCauseOk

`func (o *StackReleaseDetail) GetCauseOk() (*ReleaseCause, bool)`

GetCauseOk returns a tuple with the Cause field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCause

`func (o *StackReleaseDetail) SetCause(v ReleaseCause)`

SetCause sets Cause field to given value.

### HasCause

`func (o *StackReleaseDetail) HasCause() bool`

HasCause returns a boolean if a field has been set.

### GetSnapshotRevision

`func (o *StackReleaseDetail) GetSnapshotRevision() string`

GetSnapshotRevision returns the SnapshotRevision field if non-nil, zero value otherwise.

### GetSnapshotRevisionOk

`func (o *StackReleaseDetail) GetSnapshotRevisionOk() (*string, bool)`

GetSnapshotRevisionOk returns a tuple with the SnapshotRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSnapshotRevision

`func (o *StackReleaseDetail) SetSnapshotRevision(v string)`

SetSnapshotRevision sets SnapshotRevision field to given value.

### HasSnapshotRevision

`func (o *StackReleaseDetail) HasSnapshotRevision() bool`

HasSnapshotRevision returns a boolean if a field has been set.

### GetManifestRevision

`func (o *StackReleaseDetail) GetManifestRevision() string`

GetManifestRevision returns the ManifestRevision field if non-nil, zero value otherwise.

### GetManifestRevisionOk

`func (o *StackReleaseDetail) GetManifestRevisionOk() (*string, bool)`

GetManifestRevisionOk returns a tuple with the ManifestRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetManifestRevision

`func (o *StackReleaseDetail) SetManifestRevision(v string)`

SetManifestRevision sets ManifestRevision field to given value.

### HasManifestRevision

`func (o *StackReleaseDetail) HasManifestRevision() bool`

HasManifestRevision returns a boolean if a field has been set.

### GetRendererVersion

`func (o *StackReleaseDetail) GetRendererVersion() string`

GetRendererVersion returns the RendererVersion field if non-nil, zero value otherwise.

### GetRendererVersionOk

`func (o *StackReleaseDetail) GetRendererVersionOk() (*string, bool)`

GetRendererVersionOk returns a tuple with the RendererVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRendererVersion

`func (o *StackReleaseDetail) SetRendererVersion(v string)`

SetRendererVersion sets RendererVersion field to given value.

### HasRendererVersion

`func (o *StackReleaseDetail) HasRendererVersion() bool`

HasRendererVersion returns a boolean if a field has been set.

### GetPins

`func (o *StackReleaseDetail) GetPins() ReleasePins`

GetPins returns the Pins field if non-nil, zero value otherwise.

### GetPinsOk

`func (o *StackReleaseDetail) GetPinsOk() (*ReleasePins, bool)`

GetPinsOk returns a tuple with the Pins field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPins

`func (o *StackReleaseDetail) SetPins(v ReleasePins)`

SetPins sets Pins field to given value.

### HasPins

`func (o *StackReleaseDetail) HasPins() bool`

HasPins returns a boolean if a field has been set.

### GetOutcome

`func (o *StackReleaseDetail) GetOutcome() ReleaseOutcome`

GetOutcome returns the Outcome field if non-nil, zero value otherwise.

### GetOutcomeOk

`func (o *StackReleaseDetail) GetOutcomeOk() (*ReleaseOutcome, bool)`

GetOutcomeOk returns a tuple with the Outcome field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutcome

`func (o *StackReleaseDetail) SetOutcome(v ReleaseOutcome)`

SetOutcome sets Outcome field to given value.

### HasOutcome

`func (o *StackReleaseDetail) HasOutcome() bool`

HasOutcome returns a boolean if a field has been set.

### GetCreatedBy

`func (o *StackReleaseDetail) GetCreatedBy() string`

GetCreatedBy returns the CreatedBy field if non-nil, zero value otherwise.

### GetCreatedByOk

`func (o *StackReleaseDetail) GetCreatedByOk() (*string, bool)`

GetCreatedByOk returns a tuple with the CreatedBy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedBy

`func (o *StackReleaseDetail) SetCreatedBy(v string)`

SetCreatedBy sets CreatedBy field to given value.

### HasCreatedBy

`func (o *StackReleaseDetail) HasCreatedBy() bool`

HasCreatedBy returns a boolean if a field has been set.

### GetCreatedAt

`func (o *StackReleaseDetail) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *StackReleaseDetail) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *StackReleaseDetail) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *StackReleaseDetail) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetUpdatedAt

`func (o *StackReleaseDetail) GetUpdatedAt() time.Time`

GetUpdatedAt returns the UpdatedAt field if non-nil, zero value otherwise.

### GetUpdatedAtOk

`func (o *StackReleaseDetail) GetUpdatedAtOk() (*time.Time, bool)`

GetUpdatedAtOk returns a tuple with the UpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdatedAt

`func (o *StackReleaseDetail) SetUpdatedAt(v time.Time)`

SetUpdatedAt sets UpdatedAt field to given value.

### HasUpdatedAt

`func (o *StackReleaseDetail) HasUpdatedAt() bool`

HasUpdatedAt returns a boolean if a field has been set.

### GetRenderedAt

`func (o *StackReleaseDetail) GetRenderedAt() time.Time`

GetRenderedAt returns the RenderedAt field if non-nil, zero value otherwise.

### GetRenderedAtOk

`func (o *StackReleaseDetail) GetRenderedAtOk() (*time.Time, bool)`

GetRenderedAtOk returns a tuple with the RenderedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRenderedAt

`func (o *StackReleaseDetail) SetRenderedAt(v time.Time)`

SetRenderedAt sets RenderedAt field to given value.

### HasRenderedAt

`func (o *StackReleaseDetail) HasRenderedAt() bool`

HasRenderedAt returns a boolean if a field has been set.

### GetCompletedAt

`func (o *StackReleaseDetail) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *StackReleaseDetail) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *StackReleaseDetail) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *StackReleaseDetail) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.

### GetValidationErrors

`func (o *StackReleaseDetail) GetValidationErrors() []ReleaseValidationError`

GetValidationErrors returns the ValidationErrors field if non-nil, zero value otherwise.

### GetValidationErrorsOk

`func (o *StackReleaseDetail) GetValidationErrorsOk() (*[]ReleaseValidationError, bool)`

GetValidationErrorsOk returns a tuple with the ValidationErrors field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValidationErrors

`func (o *StackReleaseDetail) SetValidationErrors(v []ReleaseValidationError)`

SetValidationErrors sets ValidationErrors field to given value.

### HasValidationErrors

`func (o *StackReleaseDetail) HasValidationErrors() bool`

HasValidationErrors returns a boolean if a field has been set.

### GetLiveStatus

`func (o *StackReleaseDetail) GetLiveStatus() ReleaseLiveStatus`

GetLiveStatus returns the LiveStatus field if non-nil, zero value otherwise.

### GetLiveStatusOk

`func (o *StackReleaseDetail) GetLiveStatusOk() (*ReleaseLiveStatus, bool)`

GetLiveStatusOk returns a tuple with the LiveStatus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLiveStatus

`func (o *StackReleaseDetail) SetLiveStatus(v ReleaseLiveStatus)`

SetLiveStatus sets LiveStatus field to given value.

### HasLiveStatus

`func (o *StackReleaseDetail) HasLiveStatus() bool`

HasLiveStatus returns a boolean if a field has been set.

### GetSnapshot

`func (o *StackReleaseDetail) GetSnapshot() StackReleaseSnapshot`

GetSnapshot returns the Snapshot field if non-nil, zero value otherwise.

### GetSnapshotOk

`func (o *StackReleaseDetail) GetSnapshotOk() (*StackReleaseSnapshot, bool)`

GetSnapshotOk returns a tuple with the Snapshot field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSnapshot

`func (o *StackReleaseDetail) SetSnapshot(v StackReleaseSnapshot)`

SetSnapshot sets Snapshot field to given value.

### HasSnapshot

`func (o *StackReleaseDetail) HasSnapshot() bool`

HasSnapshot returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


