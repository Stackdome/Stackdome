# PostgresResourcesMemory

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Request** | Pointer to **string** | Memory request (e.g., 256Mi, 1Gi) | [optional] 
**Limit** | Pointer to **string** | Memory limit | [optional] 

## Methods

### NewPostgresResourcesMemory

`func NewPostgresResourcesMemory() *PostgresResourcesMemory`

NewPostgresResourcesMemory instantiates a new PostgresResourcesMemory object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresResourcesMemoryWithDefaults

`func NewPostgresResourcesMemoryWithDefaults() *PostgresResourcesMemory`

NewPostgresResourcesMemoryWithDefaults instantiates a new PostgresResourcesMemory object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetRequest

`func (o *PostgresResourcesMemory) GetRequest() string`

GetRequest returns the Request field if non-nil, zero value otherwise.

### GetRequestOk

`func (o *PostgresResourcesMemory) GetRequestOk() (*string, bool)`

GetRequestOk returns a tuple with the Request field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequest

`func (o *PostgresResourcesMemory) SetRequest(v string)`

SetRequest sets Request field to given value.

### HasRequest

`func (o *PostgresResourcesMemory) HasRequest() bool`

HasRequest returns a boolean if a field has been set.

### GetLimit

`func (o *PostgresResourcesMemory) GetLimit() string`

GetLimit returns the Limit field if non-nil, zero value otherwise.

### GetLimitOk

`func (o *PostgresResourcesMemory) GetLimitOk() (*string, bool)`

GetLimitOk returns a tuple with the Limit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLimit

`func (o *PostgresResourcesMemory) SetLimit(v string)`

SetLimit sets Limit field to given value.

### HasLimit

`func (o *PostgresResourcesMemory) HasLimit() bool`

HasLimit returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


