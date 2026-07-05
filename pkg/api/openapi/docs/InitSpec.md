# InitSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to **[]string** |  | [optional] 
**Args** | Pointer to **[]string** |  | [optional] 

## Methods

### NewInitSpec

`func NewInitSpec() *InitSpec`

NewInitSpec instantiates a new InitSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInitSpecWithDefaults

`func NewInitSpecWithDefaults() *InitSpec`

NewInitSpecWithDefaults instantiates a new InitSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *InitSpec) GetCommand() []string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *InitSpec) GetCommandOk() (*[]string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *InitSpec) SetCommand(v []string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *InitSpec) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### GetArgs

`func (o *InitSpec) GetArgs() []string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *InitSpec) GetArgsOk() (*[]string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *InitSpec) SetArgs(v []string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *InitSpec) HasArgs() bool`

HasArgs returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


