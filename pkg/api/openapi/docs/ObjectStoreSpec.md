# ObjectStoreSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Configuration** | [**ObjectStoreConfiguration**](ObjectStoreConfiguration.md) |  | 
**DestinationPath** | **string** | Storage destination URL (e.g., s3://bucket/path, https://account.blob.core.windows.net/container/path) | 
**RetentionPolicy** | Pointer to **string** | Retention policy (e.g., &#39;1d&#39;, &#39;7d&#39;, &#39;30d&#39;) | [optional] [default to "7d"]

## Methods

### NewObjectStoreSpec

`func NewObjectStoreSpec(configuration ObjectStoreConfiguration, destinationPath string, ) *ObjectStoreSpec`

NewObjectStoreSpec instantiates a new ObjectStoreSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewObjectStoreSpecWithDefaults

`func NewObjectStoreSpecWithDefaults() *ObjectStoreSpec`

NewObjectStoreSpecWithDefaults instantiates a new ObjectStoreSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfiguration

`func (o *ObjectStoreSpec) GetConfiguration() ObjectStoreConfiguration`

GetConfiguration returns the Configuration field if non-nil, zero value otherwise.

### GetConfigurationOk

`func (o *ObjectStoreSpec) GetConfigurationOk() (*ObjectStoreConfiguration, bool)`

GetConfigurationOk returns a tuple with the Configuration field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfiguration

`func (o *ObjectStoreSpec) SetConfiguration(v ObjectStoreConfiguration)`

SetConfiguration sets Configuration field to given value.


### GetDestinationPath

`func (o *ObjectStoreSpec) GetDestinationPath() string`

GetDestinationPath returns the DestinationPath field if non-nil, zero value otherwise.

### GetDestinationPathOk

`func (o *ObjectStoreSpec) GetDestinationPathOk() (*string, bool)`

GetDestinationPathOk returns a tuple with the DestinationPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDestinationPath

`func (o *ObjectStoreSpec) SetDestinationPath(v string)`

SetDestinationPath sets DestinationPath field to given value.


### GetRetentionPolicy

`func (o *ObjectStoreSpec) GetRetentionPolicy() string`

GetRetentionPolicy returns the RetentionPolicy field if non-nil, zero value otherwise.

### GetRetentionPolicyOk

`func (o *ObjectStoreSpec) GetRetentionPolicyOk() (*string, bool)`

GetRetentionPolicyOk returns a tuple with the RetentionPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRetentionPolicy

`func (o *ObjectStoreSpec) SetRetentionPolicy(v string)`

SetRetentionPolicy sets RetentionPolicy field to given value.

### HasRetentionPolicy

`func (o *ObjectStoreSpec) HasRetentionPolicy() bool`

HasRetentionPolicy returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


