# ReleaseEventList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | Pointer to [**[]ReleaseEvent**](ReleaseEvent.md) |  | [optional] 
**NextAfterSequence** | Pointer to **int32** |  | [optional] 

## Methods

### NewReleaseEventList

`func NewReleaseEventList() *ReleaseEventList`

NewReleaseEventList instantiates a new ReleaseEventList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReleaseEventListWithDefaults

`func NewReleaseEventListWithDefaults() *ReleaseEventList`

NewReleaseEventListWithDefaults instantiates a new ReleaseEventList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *ReleaseEventList) GetItems() []ReleaseEvent`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *ReleaseEventList) GetItemsOk() (*[]ReleaseEvent, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *ReleaseEventList) SetItems(v []ReleaseEvent)`

SetItems sets Items field to given value.

### HasItems

`func (o *ReleaseEventList) HasItems() bool`

HasItems returns a boolean if a field has been set.

### GetNextAfterSequence

`func (o *ReleaseEventList) GetNextAfterSequence() int32`

GetNextAfterSequence returns the NextAfterSequence field if non-nil, zero value otherwise.

### GetNextAfterSequenceOk

`func (o *ReleaseEventList) GetNextAfterSequenceOk() (*int32, bool)`

GetNextAfterSequenceOk returns a tuple with the NextAfterSequence field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNextAfterSequence

`func (o *ReleaseEventList) SetNextAfterSequence(v int32)`

SetNextAfterSequence sets NextAfterSequence field to given value.

### HasNextAfterSequence

`func (o *ReleaseEventList) HasNextAfterSequence() bool`

HasNextAfterSequence returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


