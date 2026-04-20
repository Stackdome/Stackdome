# PostgresResourcesCpu

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Request** | Pointer to **string** | CPU request (e.g., 100m, 1) | [optional] 
**Limit** | Pointer to **string** | CPU limit | [optional] 

## Methods

### NewPostgresResourcesCpu

`func NewPostgresResourcesCpu() *PostgresResourcesCpu`

NewPostgresResourcesCpu instantiates a new PostgresResourcesCpu object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresResourcesCpuWithDefaults

`func NewPostgresResourcesCpuWithDefaults() *PostgresResourcesCpu`

NewPostgresResourcesCpuWithDefaults instantiates a new PostgresResourcesCpu object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequest

`func (o *PostgresResourcesCpu) GetRequest() string`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *PostgresResourcesCpu) GetRequestOk() (*string, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *PostgresResourcesCpu) SetRequest(v string)`

SetRequest sets Request field to given value.

### HasRequest

`func (o *PostgresResourcesCpu) HasRequest() bool`

HasRequest returns a boolean if a field has been set.

### GetLimit

`func (o *PostgresResourcesCpu) GetLimit() string`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *PostgresResourcesCpu) GetLimitOk() (*string, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *PostgresResourcesCpu) SetLimit(v string)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *PostgresResourcesCpu) HasLimit() bool`

HasLimit returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


