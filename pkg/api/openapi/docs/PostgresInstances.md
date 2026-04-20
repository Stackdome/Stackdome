# PostgresInstances

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Count** | **int32** | Number of PostgreSQL instances | 
**Placement** | Pointer to [**PostgresInstancesPlacement**](PostgresInstancesPlacement.md) |  | [optional] 

## Methods

### NewPostgresInstances

`func NewPostgresInstances(count int32, ) *PostgresInstances`

NewPostgresInstances instantiates a new PostgresInstances object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresInstancesWithDefaults

`func NewPostgresInstancesWithDefaults() *PostgresInstances`

NewPostgresInstancesWithDefaults instantiates a new PostgresInstances object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCount

`func (o *PostgresInstances) GetCount() int32`

GetCount returns the Count field if non-nil, zero value otherwise.

### GetCountOk

`func (o *PostgresInstances) GetCountOk() (*int32, bool)`

GetCountOk returns a tuple with the Count field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCount

`func (o *PostgresInstances) SetCount(v int32)`

SetCount sets Count field to given value.


### GetPlacement

`func (o *PostgresInstances) GetPlacement() PostgresInstancesPlacement`

GetPlacement returns the Placement field if non-nil, zero value otherwise.

### GetPlacementOk

`func (o *PostgresInstances) GetPlacementOk() (*PostgresInstancesPlacement, bool)`

GetPlacementOk returns a tuple with the Placement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacement

`func (o *PostgresInstances) SetPlacement(v PostgresInstancesPlacement)`

SetPlacement sets Placement field to given value.

### HasPlacement

`func (o *PostgresInstances) HasPlacement() bool`

HasPlacement returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


