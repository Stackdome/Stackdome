# ExecutionConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to **[]string** |  | [optional] 
**Args** | Pointer to **[]string** |  | [optional] 
**EnvironmentVariables** | Pointer to [**[]EnvVar**](EnvVar.md) |  | [optional] 
**EnvironmentVariablesFromSecret** | Pointer to [**[]EnvVarFromSecret**](EnvVarFromSecret.md) |  | [optional] 
**EnvFromAddons** | Pointer to [**[]AddonEnvSource**](AddonEnvSource.md) |  | [optional] 

## Methods

### NewExecutionConfig

`func NewExecutionConfig() *ExecutionConfig`

NewExecutionConfig instantiates a new ExecutionConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewExecutionConfigWithDefaults

`func NewExecutionConfigWithDefaults() *ExecutionConfig`

NewExecutionConfigWithDefaults instantiates a new ExecutionConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *ExecutionConfig) GetCommand() []string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *ExecutionConfig) GetCommandOk() (*[]string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *ExecutionConfig) SetCommand(v []string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *ExecutionConfig) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### GetArgs

`func (o *ExecutionConfig) GetArgs() []string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *ExecutionConfig) GetArgsOk() (*[]string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *ExecutionConfig) SetArgs(v []string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *ExecutionConfig) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### GetEnvironmentVariables

`func (o *ExecutionConfig) GetEnvironmentVariables() []EnvVar`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *ExecutionConfig) GetEnvironmentVariablesOk() (*[]EnvVar, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *ExecutionConfig) SetEnvironmentVariables(v []EnvVar)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *ExecutionConfig) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### GetEnvironmentVariablesFromSecret

`func (o *ExecutionConfig) GetEnvironmentVariablesFromSecret() []EnvVarFromSecret`

GetEnvironmentVariablesFromSecret returns the EnvironmentVariablesFromSecret field if non-nil, zero value otherwise.

### GetEnvironmentVariablesFromSecretOk

`func (o *ExecutionConfig) GetEnvironmentVariablesFromSecretOk() (*[]EnvVarFromSecret, bool)`

GetEnvironmentVariablesFromSecretOk returns a tuple with the EnvironmentVariablesFromSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariablesFromSecret

`func (o *ExecutionConfig) SetEnvironmentVariablesFromSecret(v []EnvVarFromSecret)`

SetEnvironmentVariablesFromSecret sets EnvironmentVariablesFromSecret field to given value.

### HasEnvironmentVariablesFromSecret

`func (o *ExecutionConfig) HasEnvironmentVariablesFromSecret() bool`

HasEnvironmentVariablesFromSecret returns a boolean if a field has been set.

### GetEnvFromAddons

`func (o *ExecutionConfig) GetEnvFromAddons() []AddonEnvSource`

GetEnvFromAddons returns the EnvFromAddons field if non-nil, zero value otherwise.

### GetEnvFromAddonsOk

`func (o *ExecutionConfig) GetEnvFromAddonsOk() (*[]AddonEnvSource, bool)`

GetEnvFromAddonsOk returns a tuple with the EnvFromAddons field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvFromAddons

`func (o *ExecutionConfig) SetEnvFromAddons(v []AddonEnvSource)`

SetEnvFromAddons sets EnvFromAddons field to given value.

### HasEnvFromAddons

`func (o *ExecutionConfig) HasEnvFromAddons() bool`

HasEnvFromAddons returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


