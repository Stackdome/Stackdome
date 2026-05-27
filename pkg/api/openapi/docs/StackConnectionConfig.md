# StackConnectionConfig

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Database** | Pointer to **string** | Target database name within the addon. | [optional] 
**CredentialScope** | Pointer to **string** | Which credential set to inject. Mutually exclusive with superuser. | [optional] 
**Superuser** | Pointer to **bool** | Use superuser credentials. Mutually exclusive with credential_scope. | [optional] 
**MountPath** | **string** | Absolute path where the volume is mounted in the container. | 
**SubPath** | Pointer to **string** | Sub-path within the volume to mount. | [optional] 
**ReadOnly** | Pointer to **bool** | Mount the volume read-only. | [optional] 
**SourcePath** | **string** | Path within the build output to copy from. | 
**DestinationPath** | Pointer to **string** | Path within the volume to copy to. | [optional] 

## Methods

### NewStackConnectionConfig

`func NewStackConnectionConfig(mountPath string, sourcePath string, ) *StackConnectionConfig`

NewStackConnectionConfig instantiates a new StackConnectionConfig object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackConnectionConfigWithDefaults

`func NewStackConnectionConfigWithDefaults() *StackConnectionConfig`

NewStackConnectionConfigWithDefaults instantiates a new StackConnectionConfig object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDatabase

`func (o *StackConnectionConfig) GetDatabase() string`

GetDatabase returns the Database field if non-nil, zero value otherwise.

### GetDatabaseOk

`func (o *StackConnectionConfig) GetDatabaseOk() (*string, bool)`

GetDatabaseOk returns a tuple with the Database field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDatabase

`func (o *StackConnectionConfig) SetDatabase(v string)`

SetDatabase sets Database field to given value.

### HasDatabase

`func (o *StackConnectionConfig) HasDatabase() bool`

HasDatabase returns a boolean if a field has been set.

### GetCredentialScope

`func (o *StackConnectionConfig) GetCredentialScope() string`

GetCredentialScope returns the CredentialScope field if non-nil, zero value otherwise.

### GetCredentialScopeOk

`func (o *StackConnectionConfig) GetCredentialScopeOk() (*string, bool)`

GetCredentialScopeOk returns a tuple with the CredentialScope field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCredentialScope

`func (o *StackConnectionConfig) SetCredentialScope(v string)`

SetCredentialScope sets CredentialScope field to given value.

### HasCredentialScope

`func (o *StackConnectionConfig) HasCredentialScope() bool`

HasCredentialScope returns a boolean if a field has been set.

### GetSuperuser

`func (o *StackConnectionConfig) GetSuperuser() bool`

GetSuperuser returns the Superuser field if non-nil, zero value otherwise.

### GetSuperuserOk

`func (o *StackConnectionConfig) GetSuperuserOk() (*bool, bool)`

GetSuperuserOk returns a tuple with the Superuser field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSuperuser

`func (o *StackConnectionConfig) SetSuperuser(v bool)`

SetSuperuser sets Superuser field to given value.

### HasSuperuser

`func (o *StackConnectionConfig) HasSuperuser() bool`

HasSuperuser returns a boolean if a field has been set.

### GetMountPath

`func (o *StackConnectionConfig) GetMountPath() string`

GetMountPath returns the MountPath field if non-nil, zero value otherwise.

### GetMountPathOk

`func (o *StackConnectionConfig) GetMountPathOk() (*string, bool)`

GetMountPathOk returns a tuple with the MountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMountPath

`func (o *StackConnectionConfig) SetMountPath(v string)`

SetMountPath sets MountPath field to given value.


### GetSubPath

`func (o *StackConnectionConfig) GetSubPath() string`

GetSubPath returns the SubPath field if non-nil, zero value otherwise.

### GetSubPathOk

`func (o *StackConnectionConfig) GetSubPathOk() (*string, bool)`

GetSubPathOk returns a tuple with the SubPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSubPath

`func (o *StackConnectionConfig) SetSubPath(v string)`

SetSubPath sets SubPath field to given value.

### HasSubPath

`func (o *StackConnectionConfig) HasSubPath() bool`

HasSubPath returns a boolean if a field has been set.

### GetReadOnly

`func (o *StackConnectionConfig) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *StackConnectionConfig) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *StackConnectionConfig) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *StackConnectionConfig) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### GetSourcePath

`func (o *StackConnectionConfig) GetSourcePath() string`

GetSourcePath returns the SourcePath field if non-nil, zero value otherwise.

### GetSourcePathOk

`func (o *StackConnectionConfig) GetSourcePathOk() (*string, bool)`

GetSourcePathOk returns a tuple with the SourcePath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourcePath

`func (o *StackConnectionConfig) SetSourcePath(v string)`

SetSourcePath sets SourcePath field to given value.


### GetDestinationPath

`func (o *StackConnectionConfig) GetDestinationPath() string`

GetDestinationPath returns the DestinationPath field if non-nil, zero value otherwise.

### GetDestinationPathOk

`func (o *StackConnectionConfig) GetDestinationPathOk() (*string, bool)`

GetDestinationPathOk returns a tuple with the DestinationPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDestinationPath

`func (o *StackConnectionConfig) SetDestinationPath(v string)`

SetDestinationPath sets DestinationPath field to given value.

### HasDestinationPath

`func (o *StackConnectionConfig) HasDestinationPath() bool`

HasDestinationPath returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


