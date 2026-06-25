# PreviewStackList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | Pointer to [**[]PreviewStack**](PreviewStack.md) |  | [optional] 
**Total** | Pointer to **int32** | Total number of records | [optional] 
**Page** | Pointer to **int32** | Current page number | [optional] 
**PageSize** | Pointer to **int32** | Number of items per page | [optional] 
**TotalPages** | Pointer to **int32** | Total number of pages | [optional] 

## Methods

### NewPreviewStackList

`func NewPreviewStackList() *PreviewStackList`

NewPreviewStackList instantiates a new PreviewStackList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPreviewStackListWithDefaults

`func NewPreviewStackListWithDefaults() *PreviewStackList`

NewPreviewStackListWithDefaults instantiates a new PreviewStackList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *PreviewStackList) GetItems() []PreviewStack`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *PreviewStackList) GetItemsOk() (*[]PreviewStack, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *PreviewStackList) SetItems(v []PreviewStack)`

SetItems sets Items field to given value.

### HasItems

`func (o *PreviewStackList) HasItems() bool`

HasItems returns a boolean if a field has been set.

### GetTotal

`func (o *PreviewStackList) GetTotal() int32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *PreviewStackList) GetTotalOk() (*int32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *PreviewStackList) SetTotal(v int32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *PreviewStackList) HasTotal() bool`

HasTotal returns a boolean if a field has been set.

### GetPage

`func (o *PreviewStackList) GetPage() int32`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *PreviewStackList) GetPageOk() (*int32, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *PreviewStackList) SetPage(v int32)`

SetPage sets Page field to given value.

### HasPage

`func (o *PreviewStackList) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetPageSize

`func (o *PreviewStackList) GetPageSize() int32`

GetPageSize returns the PageSize field if non-nil, zero value otherwise.

### GetPageSizeOk

`func (o *PreviewStackList) GetPageSizeOk() (*int32, bool)`

GetPageSizeOk returns a tuple with the PageSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPageSize

`func (o *PreviewStackList) SetPageSize(v int32)`

SetPageSize sets PageSize field to given value.

### HasPageSize

`func (o *PreviewStackList) HasPageSize() bool`

HasPageSize returns a boolean if a field has been set.

### GetTotalPages

`func (o *PreviewStackList) GetTotalPages() int32`

GetTotalPages returns the TotalPages field if non-nil, zero value otherwise.

### GetTotalPagesOk

`func (o *PreviewStackList) GetTotalPagesOk() (*int32, bool)`

GetTotalPagesOk returns a tuple with the TotalPages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalPages

`func (o *PreviewStackList) SetTotalPages(v int32)`

SetTotalPages sets TotalPages field to given value.

### HasTotalPages

`func (o *PreviewStackList) HasTotalPages() bool`

HasTotalPages returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


