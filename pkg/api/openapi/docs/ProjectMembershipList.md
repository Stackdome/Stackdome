# ProjectMembershipList

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Items** | Pointer to [**[]ProjectMembership**](ProjectMembership.md) |  | [optional] 
**Total** | Pointer to **int32** |  | [optional] 

## Methods

### NewProjectMembershipList

`func NewProjectMembershipList() *ProjectMembershipList`

NewProjectMembershipList instantiates a new ProjectMembershipList object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectMembershipListWithDefaults

`func NewProjectMembershipListWithDefaults() *ProjectMembershipList`

NewProjectMembershipListWithDefaults instantiates a new ProjectMembershipList object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetItems

`func (o *ProjectMembershipList) GetItems() []ProjectMembership`

GetItems returns the Items field if non-nil, zero value otherwise.

### GetItemsOk

`func (o *ProjectMembershipList) GetItemsOk() (*[]ProjectMembership, bool)`

GetItemsOk returns a tuple with the Items field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItems

`func (o *ProjectMembershipList) SetItems(v []ProjectMembership)`

SetItems sets Items field to given value.

### HasItems

`func (o *ProjectMembershipList) HasItems() bool`

HasItems returns a boolean if a field has been set.

### GetTotal

`func (o *ProjectMembershipList) GetTotal() int32`

GetTotal returns the Total field if non-nil, zero value otherwise.

### GetTotalOk

`func (o *ProjectMembershipList) GetTotalOk() (*int32, bool)`

GetTotalOk returns a tuple with the Total field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTotal

`func (o *ProjectMembershipList) SetTotal(v int32)`

SetTotal sets Total field to given value.

### HasTotal

`func (o *ProjectMembershipList) HasTotal() bool`

HasTotal returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


