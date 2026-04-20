# PostgresAddonStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**State** | Pointer to **string** |  | [optional] 
**Message** | Pointer to **string** |  | [optional] 
**Phase** | Pointer to **string** |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**ObservedRevision** | Pointer to **string** |  | [optional] 
**ObservedGeneration** | Pointer to **int32** |  | [optional] 
**ClusterInfo** | Pointer to [**PostgresClusterInfo**](PostgresClusterInfo.md) |  | [optional] 
**ConnectionInfo** | Pointer to [**PostgresConnectionInfo**](PostgresConnectionInfo.md) |  | [optional] 

## Methods

### NewPostgresAddonStatus

`func NewPostgresAddonStatus() *PostgresAddonStatus`

NewPostgresAddonStatus instantiates a new PostgresAddonStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresAddonStatusWithDefaults

`func NewPostgresAddonStatusWithDefaults() *PostgresAddonStatus`

NewPostgresAddonStatusWithDefaults instantiates a new PostgresAddonStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetState

`func (o *PostgresAddonStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *PostgresAddonStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *PostgresAddonStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *PostgresAddonStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetMessage

`func (o *PostgresAddonStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *PostgresAddonStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *PostgresAddonStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *PostgresAddonStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### GetPhase

`func (o *PostgresAddonStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *PostgresAddonStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *PostgresAddonStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *PostgresAddonStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetConditions

`func (o *PostgresAddonStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *PostgresAddonStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *PostgresAddonStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *PostgresAddonStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetObservedRevision

`func (o *PostgresAddonStatus) GetObservedRevision() string`

GetObservedRevision returns the ObservedRevision field if non-nil, zero value otherwise.

### GetObservedRevisionOk

`func (o *PostgresAddonStatus) GetObservedRevisionOk() (*string, bool)`

GetObservedRevisionOk returns a tuple with the ObservedRevision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedRevision

`func (o *PostgresAddonStatus) SetObservedRevision(v string)`

SetObservedRevision sets ObservedRevision field to given value.

### HasObservedRevision

`func (o *PostgresAddonStatus) HasObservedRevision() bool`

HasObservedRevision returns a boolean if a field has been set.

### GetObservedGeneration

`func (o *PostgresAddonStatus) GetObservedGeneration() int32`

GetObservedGeneration returns the ObservedGeneration field if non-nil, zero value otherwise.

### GetObservedGenerationOk

`func (o *PostgresAddonStatus) GetObservedGenerationOk() (*int32, bool)`

GetObservedGenerationOk returns a tuple with the ObservedGeneration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedGeneration

`func (o *PostgresAddonStatus) SetObservedGeneration(v int32)`

SetObservedGeneration sets ObservedGeneration field to given value.

### HasObservedGeneration

`func (o *PostgresAddonStatus) HasObservedGeneration() bool`

HasObservedGeneration returns a boolean if a field has been set.

### GetClusterInfo

`func (o *PostgresAddonStatus) GetClusterInfo() PostgresClusterInfo`

GetClusterInfo returns the ClusterInfo field if non-nil, zero value otherwise.

### GetClusterInfoOk

`func (o *PostgresAddonStatus) GetClusterInfoOk() (*PostgresClusterInfo, bool)`

GetClusterInfoOk returns a tuple with the ClusterInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterInfo

`func (o *PostgresAddonStatus) SetClusterInfo(v PostgresClusterInfo)`

SetClusterInfo sets ClusterInfo field to given value.

### HasClusterInfo

`func (o *PostgresAddonStatus) HasClusterInfo() bool`

HasClusterInfo returns a boolean if a field has been set.

### GetConnectionInfo

`func (o *PostgresAddonStatus) GetConnectionInfo() PostgresConnectionInfo`

GetConnectionInfo returns the ConnectionInfo field if non-nil, zero value otherwise.

### GetConnectionInfoOk

`func (o *PostgresAddonStatus) GetConnectionInfoOk() (*PostgresConnectionInfo, bool)`

GetConnectionInfoOk returns a tuple with the ConnectionInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConnectionInfo

`func (o *PostgresAddonStatus) SetConnectionInfo(v PostgresConnectionInfo)`

SetConnectionInfo sets ConnectionInfo field to given value.

### HasConnectionInfo

`func (o *PostgresAddonStatus) HasConnectionInfo() bool`

HasConnectionInfo returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


