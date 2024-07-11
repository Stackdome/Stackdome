# WorkspaceStorageStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 
**Phase** | Pointer to **string** |  | [optional] 
**StorageServerServiceName** | Pointer to **string** |  | [optional] 

## Methods

### NewWorkspaceStorageStatus

`func NewWorkspaceStorageStatus() *WorkspaceStorageStatus`

NewWorkspaceStorageStatus instantiates a new WorkspaceStorageStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceStorageStatusWithDefaults

`func NewWorkspaceStorageStatusWithDefaults() *WorkspaceStorageStatus`

NewWorkspaceStorageStatusWithDefaults instantiates a new WorkspaceStorageStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConditions

`func (o *WorkspaceStorageStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *WorkspaceStorageStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *WorkspaceStorageStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *WorkspaceStorageStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.

### GetPhase

`func (o *WorkspaceStorageStatus) GetPhase() string`

GetPhase returns the Phase field if non-nil, zero value otherwise.

### GetPhaseOk

`func (o *WorkspaceStorageStatus) GetPhaseOk() (*string, bool)`

GetPhaseOk returns a tuple with the Phase field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPhase

`func (o *WorkspaceStorageStatus) SetPhase(v string)`

SetPhase sets Phase field to given value.

### HasPhase

`func (o *WorkspaceStorageStatus) HasPhase() bool`

HasPhase returns a boolean if a field has been set.

### GetStorageServerServiceName

`func (o *WorkspaceStorageStatus) GetStorageServerServiceName() string`

GetStorageServerServiceName returns the StorageServerServiceName field if non-nil, zero value otherwise.

### GetStorageServerServiceNameOk

`func (o *WorkspaceStorageStatus) GetStorageServerServiceNameOk() (*string, bool)`

GetStorageServerServiceNameOk returns a tuple with the StorageServerServiceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageServerServiceName

`func (o *WorkspaceStorageStatus) SetStorageServerServiceName(v string)`

SetStorageServerServiceName sets StorageServerServiceName field to given value.

### HasStorageServerServiceName

`func (o *WorkspaceStorageStatus) HasStorageServerServiceName() bool`

HasStorageServerServiceName returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


