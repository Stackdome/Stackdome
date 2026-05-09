# APITokenList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | Pointer to [**[]APIToken**](APIToken.md) |  | [optional] 
**Total** | Pointer to **int32** |  | [optional] 

## Methods

### NewAPITokenList

`func NewAPITokenList() *APITokenList`

NewAPITokenList instantiates a new APITokenList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAPITokenListWithDefaults

`func NewAPITokenListWithDefaults() *APITokenList`

NewAPITokenListWithDefaults instantiates a new APITokenList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *APITokenList) GetItems() []APIToken`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *APITokenList) GetItemsOk() (*[]APIToken, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *APITokenList) SetItems(v []APIToken)`

SetItems sets Items field to given value.

### HasItems

`func (o *APITokenList) HasItems() bool`

HasItems returns a boolean if a field has been set.

### GetTotal

`func (o *APITokenList) GetTotal() int32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *APITokenList) GetTotalOk() (*int32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *APITokenList) SetTotal(v int32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *APITokenList) HasTotal() bool`

HasTotal returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


