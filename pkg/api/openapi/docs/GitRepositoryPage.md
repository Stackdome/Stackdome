# GitRepositoryPage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | Pointer to [**[]GitRepository**](GitRepository.md) |  | [optional] 
**Page** | Pointer to **int32** |  | [optional] 
**TotalCount** | Pointer to **int32** |  | [optional] 
**HasNext** | Pointer to **bool** |  | [optional] 

## Methods

### NewGitRepositoryPage

`func NewGitRepositoryPage() *GitRepositoryPage`

NewGitRepositoryPage instantiates a new GitRepositoryPage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGitRepositoryPageWithDefaults

`func NewGitRepositoryPageWithDefaults() *GitRepositoryPage`

NewGitRepositoryPageWithDefaults instantiates a new GitRepositoryPage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *GitRepositoryPage) GetItems() []GitRepository`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *GitRepositoryPage) GetItemsOk() (*[]GitRepository, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *GitRepositoryPage) SetItems(v []GitRepository)`

SetItems sets Items field to given value.

### HasItems

`func (o *GitRepositoryPage) HasItems() bool`

HasItems returns a boolean if a field has been set.

### GetPage

`func (o *GitRepositoryPage) GetPage() int32`

GetPage returns the Page field if non-nil, zero value otherwise.

### GetPageOk

`func (o *GitRepositoryPage) GetPageOk() (*int32, bool)`

GetPageOk returns a tuple with the Page field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPage

`func (o *GitRepositoryPage) SetPage(v int32)`

SetPage sets Page field to given value.

### HasPage

`func (o *GitRepositoryPage) HasPage() bool`

HasPage returns a boolean if a field has been set.

### GetTotalCount

`func (o *GitRepositoryPage) GetTotalCount() int32`

GetTotalCount returns the TotalCount field if non-nil, zero value otherwise.

### GetTotalCountOk

`func (o *GitRepositoryPage) GetTotalCountOk() (*int32, bool)`

GetTotalCountOk returns a tuple with the TotalCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotalCount

`func (o *GitRepositoryPage) SetTotalCount(v int32)`

SetTotalCount sets TotalCount field to given value.

### HasTotalCount

`func (o *GitRepositoryPage) HasTotalCount() bool`

HasTotalCount returns a boolean if a field has been set.

### GetHasNext

`func (o *GitRepositoryPage) GetHasNext() bool`

GetHasNext returns the HasNext field if non-nil, zero value otherwise.

### GetHasNextOk

`func (o *GitRepositoryPage) GetHasNextOk() (*bool, bool)`

GetHasNextOk returns a tuple with the HasNext field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHasNext

`func (o *GitRepositoryPage) SetHasNext(v bool)`

SetHasNext sets HasNext field to given value.

### HasHasNext

`func (o *GitRepositoryPage) HasHasNext() bool`

HasHasNext returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


