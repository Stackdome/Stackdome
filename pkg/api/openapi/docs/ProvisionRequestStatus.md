# ProvisionRequestStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**StatusCondition** | Pointer to [**ProvisionRequestStatusCondition**](ProvisionRequestStatusCondition.md) |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewProvisionRequestStatus

`func NewProvisionRequestStatus() *ProvisionRequestStatus`

NewProvisionRequestStatus instantiates a new ProvisionRequestStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProvisionRequestStatusWithDefaults

`func NewProvisionRequestStatusWithDefaults() *ProvisionRequestStatus`

NewProvisionRequestStatusWithDefaults instantiates a new ProvisionRequestStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetStatusCondition

`func (o *ProvisionRequestStatus) GetStatusCondition() ProvisionRequestStatusCondition`

GetStatusCondition returns the StatusCondition field if non-nil, zero value otherwise.

### GetStatusConditionOk

`func (o *ProvisionRequestStatus) GetStatusConditionOk() (*ProvisionRequestStatusCondition, bool)`

GetStatusConditionOk returns a tuple with the StatusCondition field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusCondition

`func (o *ProvisionRequestStatus) SetStatusCondition(v ProvisionRequestStatusCondition)`

SetStatusCondition sets StatusCondition field to given value.

### HasStatusCondition

`func (o *ProvisionRequestStatus) HasStatusCondition() bool`

HasStatusCondition returns a boolean if a field has been set.

### GetMessage

`func (o *ProvisionRequestStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ProvisionRequestStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ProvisionRequestStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ProvisionRequestStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


