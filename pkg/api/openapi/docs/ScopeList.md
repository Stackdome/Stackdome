# ScopeList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FullAccessScope** | Pointer to **string** | The wildcard scope that grants full access (same permissions as the user) | [optional] 
**Items** | Pointer to [**[]ScopeResource**](ScopeResource.md) |  | [optional] 
**Total** | Pointer to **int32** |  | [optional] 

## Methods

### NewScopeList

`func NewScopeList() *ScopeList`

NewScopeList instantiates a new ScopeList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewScopeListWithDefaults

`func NewScopeListWithDefaults() *ScopeList`

NewScopeListWithDefaults instantiates a new ScopeList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFullAccessScope

`func (o *ScopeList) GetFullAccessScope() string`

GetFullAccessScope returns the FullAccessScope field if non-nil, zero value otherwise.

### GetFullAccessScopeOk

`func (o *ScopeList) GetFullAccessScopeOk() (*string, bool)`

GetFullAccessScopeOk returns a tuple with the FullAccessScope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFullAccessScope

`func (o *ScopeList) SetFullAccessScope(v string)`

SetFullAccessScope sets FullAccessScope field to given value.

### HasFullAccessScope

`func (o *ScopeList) HasFullAccessScope() bool`

HasFullAccessScope returns a boolean if a field has been set.

### GetItems

`func (o *ScopeList) GetItems() []ScopeResource`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *ScopeList) GetItemsOk() (*[]ScopeResource, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *ScopeList) SetItems(v []ScopeResource)`

SetItems sets Items field to given value.

### HasItems

`func (o *ScopeList) HasItems() bool`

HasItems returns a boolean if a field has been set.

### GetTotal

`func (o *ScopeList) GetTotal() int32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *ScopeList) GetTotalOk() (*int32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *ScopeList) SetTotal(v int32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *ScopeList) HasTotal() bool`

HasTotal returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


