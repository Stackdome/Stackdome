# AddonEnvSource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Postgres** | Pointer to [**PostgresAddonEnvSource**](PostgresAddonEnvSource.md) |  | [optional] 

## Methods

### NewAddonEnvSource

`func NewAddonEnvSource() *AddonEnvSource`

NewAddonEnvSource instantiates a new AddonEnvSource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAddonEnvSourceWithDefaults

`func NewAddonEnvSourceWithDefaults() *AddonEnvSource`

NewAddonEnvSourceWithDefaults instantiates a new AddonEnvSource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPostgres

`func (o *AddonEnvSource) GetPostgres() PostgresAddonEnvSource`

GetPostgres returns the Postgres field if non-nil, zero value otherwise.

### GetPostgresOk

`func (o *AddonEnvSource) GetPostgresOk() (*PostgresAddonEnvSource, bool)`

GetPostgresOk returns a tuple with the Postgres field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPostgres

`func (o *AddonEnvSource) SetPostgres(v PostgresAddonEnvSource)`

SetPostgres sets Postgres field to given value.

### HasPostgres

`func (o *AddonEnvSource) HasPostgres() bool`

HasPostgres returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


