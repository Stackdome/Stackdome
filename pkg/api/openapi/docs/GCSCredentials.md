# GCSCredentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServiceAccountKey** | **string** | Reference to secret containing GCS service account key JSON | 

## Methods

### NewGCSCredentials

`func NewGCSCredentials(serviceAccountKey string, ) *GCSCredentials`

NewGCSCredentials instantiates a new GCSCredentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGCSCredentialsWithDefaults

`func NewGCSCredentialsWithDefaults() *GCSCredentials`

NewGCSCredentialsWithDefaults instantiates a new GCSCredentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServiceAccountKey

`func (o *GCSCredentials) GetServiceAccountKey() string`

GetServiceAccountKey returns the ServiceAccountKey field if non-nil, zero value otherwise.

### GetServiceAccountKeyOk

`func (o *GCSCredentials) GetServiceAccountKeyOk() (*string, bool)`

GetServiceAccountKeyOk returns a tuple with the ServiceAccountKey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceAccountKey

`func (o *GCSCredentials) SetServiceAccountKey(v string)`

SetServiceAccountKey sets ServiceAccountKey field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


