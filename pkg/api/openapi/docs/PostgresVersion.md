# PostgresVersion

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Major** | **int32** | PostgreSQL major version | 
**Minor** | Pointer to **int32** | PostgreSQL minor version | [optional] 
**EnableAutoMinorUpgrade** | Pointer to **bool** | Enable automatic minor version upgrades | [optional] [default to true]
**EnableAutoMajorUpgrade** | Pointer to **bool** | Enable automatic major version upgrades | [optional] [default to false]

## Methods

### NewPostgresVersion

`func NewPostgresVersion(major int32, ) *PostgresVersion`

NewPostgresVersion instantiates a new PostgresVersion object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPostgresVersionWithDefaults

`func NewPostgresVersionWithDefaults() *PostgresVersion`

NewPostgresVersionWithDefaults instantiates a new PostgresVersion object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMajor

`func (o *PostgresVersion) GetMajor() int32`

GetMajor returns the Major field if non-nil, zero value otherwise.

### GetMajorOk

`func (o *PostgresVersion) GetMajorOk() (*int32, bool)`

GetMajorOk returns a tuple with the Major field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMajor

`func (o *PostgresVersion) SetMajor(v int32)`

SetMajor sets Major field to given value.


### GetMinor

`func (o *PostgresVersion) GetMinor() int32`

GetMinor returns the Minor field if non-nil, zero value otherwise.

### GetMinorOk

`func (o *PostgresVersion) GetMinorOk() (*int32, bool)`

GetMinorOk returns a tuple with the Minor field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinor

`func (o *PostgresVersion) SetMinor(v int32)`

SetMinor sets Minor field to given value.

### HasMinor

`func (o *PostgresVersion) HasMinor() bool`

HasMinor returns a boolean if a field has been set.

### GetEnableAutoMinorUpgrade

`func (o *PostgresVersion) GetEnableAutoMinorUpgrade() bool`

GetEnableAutoMinorUpgrade returns the EnableAutoMinorUpgrade field if non-nil, zero value otherwise.

### GetEnableAutoMinorUpgradeOk

`func (o *PostgresVersion) GetEnableAutoMinorUpgradeOk() (*bool, bool)`

GetEnableAutoMinorUpgradeOk returns a tuple with the EnableAutoMinorUpgrade field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnableAutoMinorUpgrade

`func (o *PostgresVersion) SetEnableAutoMinorUpgrade(v bool)`

SetEnableAutoMinorUpgrade sets EnableAutoMinorUpgrade field to given value.

### HasEnableAutoMinorUpgrade

`func (o *PostgresVersion) HasEnableAutoMinorUpgrade() bool`

HasEnableAutoMinorUpgrade returns a boolean if a field has been set.

### GetEnableAutoMajorUpgrade

`func (o *PostgresVersion) GetEnableAutoMajorUpgrade() bool`

GetEnableAutoMajorUpgrade returns the EnableAutoMajorUpgrade field if non-nil, zero value otherwise.

### GetEnableAutoMajorUpgradeOk

`func (o *PostgresVersion) GetEnableAutoMajorUpgradeOk() (*bool, bool)`

GetEnableAutoMajorUpgradeOk returns a tuple with the EnableAutoMajorUpgrade field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnableAutoMajorUpgrade

`func (o *PostgresVersion) SetEnableAutoMajorUpgrade(v bool)`

SetEnableAutoMajorUpgrade sets EnableAutoMajorUpgrade field to given value.

### HasEnableAutoMajorUpgrade

`func (o *PostgresVersion) HasEnableAutoMajorUpgrade() bool`

HasEnableAutoMajorUpgrade returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


