# GCSCredentials

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ServiceAccountCredentials** | [**SecretReference**](SecretReference.md) |  | 

## Methods

### NewGCSCredentials

`func NewGCSCredentials(serviceAccountCredentials SecretReference, ) *GCSCredentials`

NewGCSCredentials instantiates a new GCSCredentials object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewGCSCredentialsWithDefaults

`func NewGCSCredentialsWithDefaults() *GCSCredentials`

NewGCSCredentialsWithDefaults instantiates a new GCSCredentials object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetServiceAccountCredentials

`func (o *GCSCredentials) GetServiceAccountCredentials() SecretReference`

GetServiceAccountCredentials returns the ServiceAccountCredentials field if non-nil, zero value otherwise.

### GetServiceAccountCredentialsOk

`func (o *GCSCredentials) GetServiceAccountCredentialsOk() (*SecretReference, bool)`

GetServiceAccountCredentialsOk returns a tuple with the ServiceAccountCredentials field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceAccountCredentials

`func (o *GCSCredentials) SetServiceAccountCredentials(v SecretReference)`

SetServiceAccountCredentials sets ServiceAccountCredentials field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


