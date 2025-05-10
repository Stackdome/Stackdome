# RemoteSyncServerStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ObservedVersion** | Pointer to **int32** |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**State** | Pointer to [**RemoteSyncServerState**](RemoteSyncServerState.md) |  | [optional] 
**ServiceName** | Pointer to **string** |  | [optional] 

## Methods

### NewRemoteSyncServerStatus

`func NewRemoteSyncServerStatus() *RemoteSyncServerStatus`

NewRemoteSyncServerStatus instantiates a new RemoteSyncServerStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRemoteSyncServerStatusWithDefaults

`func NewRemoteSyncServerStatusWithDefaults() *RemoteSyncServerStatus`

NewRemoteSyncServerStatusWithDefaults instantiates a new RemoteSyncServerStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetObservedVersion

`func (o *RemoteSyncServerStatus) GetObservedVersion() int32`

GetObservedVersion returns the ObservedVersion field if non-nil, zero value otherwise.

### GetObservedVersionOk

`func (o *RemoteSyncServerStatus) GetObservedVersionOk() (*int32, bool)`

GetObservedVersionOk returns a tuple with the ObservedVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedVersion

`func (o *RemoteSyncServerStatus) SetObservedVersion(v int32)`

SetObservedVersion sets ObservedVersion field to given value.

### HasObservedVersion

`func (o *RemoteSyncServerStatus) HasObservedVersion() bool`

HasObservedVersion returns a boolean if a field has been set.

### GetConditions

`func (o *RemoteSyncServerStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *RemoteSyncServerStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *RemoteSyncServerStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *RemoteSyncServerStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetState

`func (o *RemoteSyncServerStatus) GetState() RemoteSyncServerState`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *RemoteSyncServerStatus) GetStateOk() (*RemoteSyncServerState, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *RemoteSyncServerStatus) SetState(v RemoteSyncServerState)`

SetState sets State field to given value.

### HasState

`func (o *RemoteSyncServerStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetServiceName

`func (o *RemoteSyncServerStatus) GetServiceName() string`

GetServiceName returns the ServiceName field if non-nil, zero value otherwise.

### GetServiceNameOk

`func (o *RemoteSyncServerStatus) GetServiceNameOk() (*string, bool)`

GetServiceNameOk returns a tuple with the ServiceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceName

`func (o *RemoteSyncServerStatus) SetServiceName(v string)`

SetServiceName sets ServiceName field to given value.

### HasServiceName

`func (o *RemoteSyncServerStatus) HasServiceName() bool`

HasServiceName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


