# ValueRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Output** | Pointer to **string** | Output accessor on the connection&#39;s &#x60;from&#x60; node, such as &#x60;url&#x60; or &#x60;public_url&#x60;. | [optional] 
**Template** | Pointer to **string** | Template used when one target value must be composed from multiple outputs. | [optional] 
**Values** | Pointer to [**map[string]OutputValueRef**](OutputValueRef.md) | Named template inputs, each resolving one output from the connection&#39;s &#x60;from&#x60; node. | [optional] 

## Methods

### NewValueRef

`func NewValueRef() *ValueRef`

NewValueRef instantiates a new ValueRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewValueRefWithDefaults

`func NewValueRefWithDefaults() *ValueRef`

NewValueRefWithDefaults instantiates a new ValueRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetOutput

`func (o *ValueRef) GetOutput() string`

GetOutput returns the Output field if non-nil, zero value otherwise.

### GetOutputOk

`func (o *ValueRef) GetOutputOk() (*string, bool)`

GetOutputOk returns a tuple with the Output field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutput

`func (o *ValueRef) SetOutput(v string)`

SetOutput sets Output field to given value.

### HasOutput

`func (o *ValueRef) HasOutput() bool`

HasOutput returns a boolean if a field has been set.

### GetTemplate

`func (o *ValueRef) GetTemplate() string`

GetTemplate returns the Template field if non-nil, zero value otherwise.

### GetTemplateOk

`func (o *ValueRef) GetTemplateOk() (*string, bool)`

GetTemplateOk returns a tuple with the Template field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTemplate

`func (o *ValueRef) SetTemplate(v string)`

SetTemplate sets Template field to given value.

### HasTemplate

`func (o *ValueRef) HasTemplate() bool`

HasTemplate returns a boolean if a field has been set.

### GetValues

`func (o *ValueRef) GetValues() map[string]OutputValueRef`

GetValues returns the Values field if non-nil, zero value otherwise.

### GetValuesOk

`func (o *ValueRef) GetValuesOk() (*map[string]OutputValueRef, bool)`

GetValuesOk returns a tuple with the Values field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValues

`func (o *ValueRef) SetValues(v map[string]OutputValueRef)`

SetValues sets Values field to given value.

### HasValues

`func (o *ValueRef) HasValues() bool`

HasValues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


