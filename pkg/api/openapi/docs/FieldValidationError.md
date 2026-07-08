# FieldValidationError

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Field** | Pointer to **string** | JSON path of the offending request field, e.g. \&quot;ports[0].protocol\&quot;. | [optional] 
**Code** | Pointer to **string** | Machine-readable validation code, e.g. \&quot;public_port_not_http\&quot;. | [optional] 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewFieldValidationError

`func NewFieldValidationError() *FieldValidationError`

NewFieldValidationError instantiates a new FieldValidationError object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFieldValidationErrorWithDefaults

`func NewFieldValidationErrorWithDefaults() *FieldValidationError`

NewFieldValidationErrorWithDefaults instantiates a new FieldValidationError object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetField

`func (o *FieldValidationError) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *FieldValidationError) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *FieldValidationError) SetField(v string)`

SetField sets Field field to given value.

### HasField

`func (o *FieldValidationError) HasField() bool`

HasField returns a boolean if a field has been set.

### GetCode

`func (o *FieldValidationError) GetCode() string`

GetCode returns the Code field if non-nil, zero value otherwise.

### GetCodeOk

`func (o *FieldValidationError) GetCodeOk() (*string, bool)`

GetCodeOk returns a tuple with the Code field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCode

`func (o *FieldValidationError) SetCode(v string)`

SetCode sets Code field to given value.

### HasCode

`func (o *FieldValidationError) HasCode() bool`

HasCode returns a boolean if a field has been set.

### GetMessage

`func (o *FieldValidationError) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *FieldValidationError) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *FieldValidationError) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *FieldValidationError) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


