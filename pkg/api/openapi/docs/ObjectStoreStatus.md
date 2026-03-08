# ObjectStoreStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 

## Methods

### NewObjectStoreStatus

`func NewObjectStoreStatus() *ObjectStoreStatus`

NewObjectStoreStatus instantiates a new ObjectStoreStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewObjectStoreStatusWithDefaults

`func NewObjectStoreStatusWithDefaults() *ObjectStoreStatus`

NewObjectStoreStatusWithDefaults instantiates a new ObjectStoreStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *ObjectStoreStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ObjectStoreStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ObjectStoreStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ObjectStoreStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetMessage

`func (o *ObjectStoreStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *ObjectStoreStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *ObjectStoreStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *ObjectStoreStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


