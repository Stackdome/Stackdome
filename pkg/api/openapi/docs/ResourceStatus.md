# ResourceStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**PublicIngress** | Pointer to [**[]Ingress**](Ingress.md) |  | [optional] 
**InternalServiceName** | Pointer to **string** |  | [optional] 
**State** | Pointer to **string** |  | [optional] 
**ObservedVersion** | Pointer to **int32** |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 

## Methods

### NewResourceStatus

`func NewResourceStatus() *ResourceStatus`

NewResourceStatus instantiates a new ResourceStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewResourceStatusWithDefaults

`func NewResourceStatusWithDefaults() *ResourceStatus`

NewResourceStatusWithDefaults instantiates a new ResourceStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPublicIngress

`func (o *ResourceStatus) GetPublicIngress() []Ingress`

GetPublicIngress returns the PublicIngress field if non-nil, zero value otherwise.

### GetPublicIngressOk

`func (o *ResourceStatus) GetPublicIngressOk() (*[]Ingress, bool)`

GetPublicIngressOk returns a tuple with the PublicIngress field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicIngress

`func (o *ResourceStatus) SetPublicIngress(v []Ingress)`

SetPublicIngress sets PublicIngress field to given value.

### HasPublicIngress

`func (o *ResourceStatus) HasPublicIngress() bool`

HasPublicIngress returns a boolean if a field has been set.

### GetInternalServiceName

`func (o *ResourceStatus) GetInternalServiceName() string`

GetInternalServiceName returns the InternalServiceName field if non-nil, zero value otherwise.

### GetInternalServiceNameOk

`func (o *ResourceStatus) GetInternalServiceNameOk() (*string, bool)`

GetInternalServiceNameOk returns a tuple with the InternalServiceName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInternalServiceName

`func (o *ResourceStatus) SetInternalServiceName(v string)`

SetInternalServiceName sets InternalServiceName field to given value.

### HasInternalServiceName

`func (o *ResourceStatus) HasInternalServiceName() bool`

HasInternalServiceName returns a boolean if a field has been set.

### GetState

`func (o *ResourceStatus) GetState() string`

GetState returns the State field if non-nil, zero value otherwise.

### GetStateOk

`func (o *ResourceStatus) GetStateOk() (*string, bool)`

GetStateOk returns a tuple with the State field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetState

`func (o *ResourceStatus) SetState(v string)`

SetState sets State field to given value.

### HasState

`func (o *ResourceStatus) HasState() bool`

HasState returns a boolean if a field has been set.

### GetObservedVersion

`func (o *ResourceStatus) GetObservedVersion() int32`

GetObservedVersion returns the ObservedVersion field if non-nil, zero value otherwise.

### GetObservedVersionOk

`func (o *ResourceStatus) GetObservedVersionOk() (*int32, bool)`

GetObservedVersionOk returns a tuple with the ObservedVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedVersion

`func (o *ResourceStatus) SetObservedVersion(v int32)`

SetObservedVersion sets ObservedVersion field to given value.

### HasObservedVersion

`func (o *ResourceStatus) HasObservedVersion() bool`

HasObservedVersion returns a boolean if a field has been set.

### GetConditions

`func (o *ResourceStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *ResourceStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *ResourceStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *ResourceStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


