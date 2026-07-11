# ReleaseSummary

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] 
**Sequence** | Pointer to **int32** |  | [optional] 
**State** | Pointer to [**StackReleaseState**](StackReleaseState.md) |  | [optional] 
**Health** | Pointer to [**ReleaseHealth**](ReleaseHealth.md) |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**CreatedAt** | Pointer to **time.Time** |  | [optional] 
**CompletedAt** | Pointer to **time.Time** |  | [optional] 

## Methods

### NewReleaseSummary

`func NewReleaseSummary() *ReleaseSummary`

NewReleaseSummary instantiates a new ReleaseSummary object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReleaseSummaryWithDefaults

`func NewReleaseSummaryWithDefaults() *ReleaseSummary`

NewReleaseSummaryWithDefaults instantiates a new ReleaseSummary object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *ReleaseSummary) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *ReleaseSummary) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *ReleaseSummary) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *ReleaseSummary) HasId() bool`

HasId returns a boolean if a field has been set.

### GetSequence

`func (o *ReleaseSummary) GetSequence() int32`

GetSequence returns the Sequence field if non-nil, zero value otherwise.

### GetSequenceOk

`func (o *ReleaseSummary) GetSequenceOk() (*int32, bool)`

GetSequenceOk returns a tuple with the Sequence field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSequence

`func (o *ReleaseSummary) SetSequence(v int32)`

SetSequence sets Sequence field to given value.

### HasSequence

`func (o *ReleaseSummary) HasSequence() bool`

HasSequence returns a boolean if a field has been set.

### GetState

`func (o *ReleaseSummary) GetState() StackReleaseState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ReleaseSummary) GetStateOk() (*StackReleaseState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ReleaseSummary) SetState(v StackReleaseState)`

SetState sets State field to given value.

### HasState

`func (o *ReleaseSummary) HasState() bool`

HasState returns a boolean if a field has been set.

### GetHealth

`func (o *ReleaseSummary) GetHealth() ReleaseHealth`

GetHealth returns the Health field if non-nil, zero value otherwise.

### GetHealthOk

`func (o *ReleaseSummary) GetHealthOk() (*ReleaseHealth, bool)`

GetHealthOk returns a tuple with the Health field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHealth

`func (o *ReleaseSummary) SetHealth(v ReleaseHealth)`

SetHealth sets Health field to given value.

### HasHealth

`func (o *ReleaseSummary) HasHealth() bool`

HasHealth returns a boolean if a field has been set.

### GetMessage

`func (o *ReleaseSummary) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ReleaseSummary) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ReleaseSummary) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ReleaseSummary) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetCreatedAt

`func (o *ReleaseSummary) GetCreatedAt() time.Time`

GetCreatedAt returns the CreatedAt field if non-nil, zero value otherwise.

### GetCreatedAtOk

`func (o *ReleaseSummary) GetCreatedAtOk() (*time.Time, bool)`

GetCreatedAtOk returns a tuple with the CreatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreatedAt

`func (o *ReleaseSummary) SetCreatedAt(v time.Time)`

SetCreatedAt sets CreatedAt field to given value.

### HasCreatedAt

`func (o *ReleaseSummary) HasCreatedAt() bool`

HasCreatedAt returns a boolean if a field has been set.

### GetCompletedAt

`func (o *ReleaseSummary) GetCompletedAt() time.Time`

GetCompletedAt returns the CompletedAt field if non-nil, zero value otherwise.

### GetCompletedAtOk

`func (o *ReleaseSummary) GetCompletedAtOk() (*time.Time, bool)`

GetCompletedAtOk returns a tuple with the CompletedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletedAt

`func (o *ReleaseSummary) SetCompletedAt(v time.Time)`

SetCompletedAt sets CompletedAt field to given value.

### HasCompletedAt

`func (o *ReleaseSummary) HasCompletedAt() bool`

HasCompletedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


