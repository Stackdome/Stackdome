# ContainerFailureDetail

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**FailureType** | Pointer to **string** |  | [optional] 
**Reason** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**RestartCount** | Pointer to **int32** |  | [optional] 
**ExitCode** | Pointer to **int32** |  | [optional] 

## Methods

### NewContainerFailureDetail

`func NewContainerFailureDetail() *ContainerFailureDetail`

NewContainerFailureDetail instantiates a new ContainerFailureDetail object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewContainerFailureDetailWithDefaults

`func NewContainerFailureDetailWithDefaults() *ContainerFailureDetail`

NewContainerFailureDetailWithDefaults instantiates a new ContainerFailureDetail object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetFailureType

`func (o *ContainerFailureDetail) GetFailureType() string`

GetFailureType returns the FailureType field if non-nil, zero value otherwise.

### GetFailureTypeOk

`func (o *ContainerFailureDetail) GetFailureTypeOk() (*string, bool)`

GetFailureTypeOk returns a tuple with the FailureType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFailureType

`func (o *ContainerFailureDetail) SetFailureType(v string)`

SetFailureType sets FailureType field to given value.

### HasFailureType

`func (o *ContainerFailureDetail) HasFailureType() bool`

HasFailureType returns a boolean if a field has been set.

### GetReason

`func (o *ContainerFailureDetail) GetReason() string`

GetReason returns the Reason field if non-nil, zero value otherwise.

### GetReasonOk

`func (o *ContainerFailureDetail) GetReasonOk() (*string, bool)`

GetReasonOk returns a tuple with the Reason field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReason

`func (o *ContainerFailureDetail) SetReason(v string)`

SetReason sets Reason field to given value.

### HasReason

`func (o *ContainerFailureDetail) HasReason() bool`

HasReason returns a boolean if a field has been set.

### GetMessage

`func (o *ContainerFailureDetail) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ContainerFailureDetail) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ContainerFailureDetail) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ContainerFailureDetail) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetRestartCount

`func (o *ContainerFailureDetail) GetRestartCount() int32`

GetRestartCount returns the RestartCount field if non-nil, zero value otherwise.

### GetRestartCountOk

`func (o *ContainerFailureDetail) GetRestartCountOk() (*int32, bool)`

GetRestartCountOk returns a tuple with the RestartCount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartCount

`func (o *ContainerFailureDetail) SetRestartCount(v int32)`

SetRestartCount sets RestartCount field to given value.

### HasRestartCount

`func (o *ContainerFailureDetail) HasRestartCount() bool`

HasRestartCount returns a boolean if a field has been set.

### GetExitCode

`func (o *ContainerFailureDetail) GetExitCode() int32`

GetExitCode returns the ExitCode field if non-nil, zero value otherwise.

### GetExitCodeOk

`func (o *ContainerFailureDetail) GetExitCodeOk() (*int32, bool)`

GetExitCodeOk returns a tuple with the ExitCode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExitCode

`func (o *ContainerFailureDetail) SetExitCode(v int32)`

SetExitCode sets ExitCode field to given value.

### HasExitCode

`func (o *ContainerFailureDetail) HasExitCode() bool`

HasExitCode returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


