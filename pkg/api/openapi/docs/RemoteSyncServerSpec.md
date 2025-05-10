# RemoteSyncServerSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**VolumeName** | Pointer to **string** |  | [optional] 
**SshPublicKey** | Pointer to **string** |  | [optional] 

## Methods

### NewRemoteSyncServerSpec

`func NewRemoteSyncServerSpec() *RemoteSyncServerSpec`

NewRemoteSyncServerSpec instantiates a new RemoteSyncServerSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRemoteSyncServerSpecWithDefaults

`func NewRemoteSyncServerSpecWithDefaults() *RemoteSyncServerSpec`

NewRemoteSyncServerSpecWithDefaults instantiates a new RemoteSyncServerSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetVolumeName

`func (o *RemoteSyncServerSpec) GetVolumeName() string`

GetVolumeName returns the VolumeName field if non-nil, zero value otherwise.

### GetVolumeNameOk

`func (o *RemoteSyncServerSpec) GetVolumeNameOk() (*string, bool)`

GetVolumeNameOk returns a tuple with the VolumeName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeName

`func (o *RemoteSyncServerSpec) SetVolumeName(v string)`

SetVolumeName sets VolumeName field to given value.

### HasVolumeName

`func (o *RemoteSyncServerSpec) HasVolumeName() bool`

HasVolumeName returns a boolean if a field has been set.

### GetSshPublicKey

`func (o *RemoteSyncServerSpec) GetSshPublicKey() string`

GetSshPublicKey returns the SshPublicKey field if non-nil, zero value otherwise.

### GetSshPublicKeyOk

`func (o *RemoteSyncServerSpec) GetSshPublicKeyOk() (*string, bool)`

GetSshPublicKeyOk returns a tuple with the SshPublicKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshPublicKey

`func (o *RemoteSyncServerSpec) SetSshPublicKey(v string)`

SetSshPublicKey sets SshPublicKey field to given value.

### HasSshPublicKey

`func (o *RemoteSyncServerSpec) HasSshPublicKey() bool`

HasSshPublicKey returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


