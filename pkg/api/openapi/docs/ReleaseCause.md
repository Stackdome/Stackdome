# ReleaseCause

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Kind** | Pointer to [**ReleaseCauseKind**](ReleaseCauseKind.md) |  | [optional] 
**Detail** | Pointer to **string** |  | [optional] 

## Methods

### NewReleaseCause

`func NewReleaseCause() *ReleaseCause`

NewReleaseCause instantiates a new ReleaseCause object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewReleaseCauseWithDefaults

`func NewReleaseCauseWithDefaults() *ReleaseCause`

NewReleaseCauseWithDefaults instantiates a new ReleaseCause object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetKind

`func (o *ReleaseCause) GetKind() ReleaseCauseKind`

GetKind returns the Kind field if non-nil, zero value otherwise.

### GetKindOk

`func (o *ReleaseCause) GetKindOk() (*ReleaseCauseKind, bool)`

GetKindOk returns a tuple with the Kind field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetKind

`func (o *ReleaseCause) SetKind(v ReleaseCauseKind)`

SetKind sets Kind field to given value.

### HasKind

`func (o *ReleaseCause) HasKind() bool`

HasKind returns a boolean if a field has been set.

### GetDetail

`func (o *ReleaseCause) GetDetail() string`

GetDetail returns the Detail field if non-nil, zero value otherwise.

### GetDetailOk

`func (o *ReleaseCause) GetDetailOk() (*string, bool)`

GetDetailOk returns a tuple with the Detail field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetail

`func (o *ReleaseCause) SetDetail(v string)`

SetDetail sets Detail field to given value.

### HasDetail

`func (o *ReleaseCause) HasDetail() bool`

HasDetail returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


