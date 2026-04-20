# WALConfiguration

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Compression** | Pointer to **string** | Compression type for WAL files | [optional] [default to "gzip"]

## Methods

### NewWALConfiguration

`func NewWALConfiguration() *WALConfiguration`

NewWALConfiguration instantiates a new WALConfiguration object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWALConfigurationWithDefaults

`func NewWALConfigurationWithDefaults() *WALConfiguration`

NewWALConfigurationWithDefaults instantiates a new WALConfiguration object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCompression

`func (o *WALConfiguration) GetCompression() string`

GetCompression returns the Compression field if non-nil, zero value otherwise.

### GetCompressionOk

`func (o *WALConfiguration) GetCompressionOk() (*string, bool)`

GetCompressionOk returns a tuple with the Compression field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompression

`func (o *WALConfiguration) SetCompression(v string)`

SetCompression sets Compression field to given value.

### HasCompression

`func (o *WALConfiguration) HasCompression() bool`

HasCompression returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


