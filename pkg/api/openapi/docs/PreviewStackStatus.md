# PreviewStackStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Phase** | Pointer to **string** |  | [optional] 
**Reason** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Outputs** | Pointer to [**PreviewStackStatusOutputs**](PreviewStackStatusOutputs.md) |  | [optional] 

## Methods

### NewPreviewStackStatus

`func NewPreviewStackStatus() *PreviewStackStatus`

NewPreviewStackStatus instantiates a new PreviewStackStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPreviewStackStatusWithDefaults

`func NewPreviewStackStatusWithDefaults() *PreviewStackStatus`

NewPreviewStackStatusWithDefaults instantiates a new PreviewStackStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPhase

`func (o *PreviewStackStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PreviewStackStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PreviewStackStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *PreviewStackStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetReason

`func (o *PreviewStackStatus) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *PreviewStackStatus) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *PreviewStackStatus) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *PreviewStackStatus) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetMessage

`func (o *PreviewStackStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *PreviewStackStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *PreviewStackStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *PreviewStackStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetOutputs

`func (o *PreviewStackStatus) GetOutputs() PreviewStackStatusOutputs`

GetOutputs returns the Outputs field if non-nil, zero value otherwise.

### GetOutputsOk

`func (o *PreviewStackStatus) GetOutputsOk() (*PreviewStackStatusOutputs, bool)`

GetOutputsOk returns a tuple with the Outputs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutputs

`func (o *PreviewStackStatus) SetOutputs(v PreviewStackStatusOutputs)`

SetOutputs sets Outputs field to given value.

### HasOutputs

`func (o *PreviewStackStatus) HasOutputs() bool`

HasOutputs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


