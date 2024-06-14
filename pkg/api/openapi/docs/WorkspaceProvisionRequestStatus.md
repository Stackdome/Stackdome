# WorkspaceProvisionRequestStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**WorkspaceNamespace** | Pointer to **NullableString** |  | [optional] 
**WorkspaceServiceAccountname** | Pointer to **NullableString** |  | [optional] 
**WorkspaceServiceaccountToken** | Pointer to **NullableString** |  | [optional] 
**ClusterCaCert** | Pointer to **NullableString** |  | [optional] 
**ClusterUrl** | Pointer to **NullableString** |  | [optional] 
**Domain** | Pointer to **NullableString** |  | [optional] 
**StatusCondition** | Pointer to [**ProvisionRequestStatusCondition**](ProvisionRequestStatusCondition.md) |  | [optional] 
**Message** | Pointer to **NullableString** |  | [optional] 

## Methods

### NewWorkspaceProvisionRequestStatus

`func NewWorkspaceProvisionRequestStatus() *WorkspaceProvisionRequestStatus`

NewWorkspaceProvisionRequestStatus instantiates a new WorkspaceProvisionRequestStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceProvisionRequestStatusWithDefaults

`func NewWorkspaceProvisionRequestStatusWithDefaults() *WorkspaceProvisionRequestStatus`

NewWorkspaceProvisionRequestStatusWithDefaults instantiates a new WorkspaceProvisionRequestStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkspaceNamespace

`func (o *WorkspaceProvisionRequestStatus) GetWorkspaceNamespace() string`

GetWorkspaceNamespace returns the WorkspaceNamespace field if non-nil, zero value otherwise.

### GetWorkspaceNamespaceOk

`func (o *WorkspaceProvisionRequestStatus) GetWorkspaceNamespaceOk() (*string, bool)`

GetWorkspaceNamespaceOk returns a tuple with the WorkspaceNamespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceNamespace

`func (o *WorkspaceProvisionRequestStatus) SetWorkspaceNamespace(v string)`

SetWorkspaceNamespace sets WorkspaceNamespace field to given value.

### HasWorkspaceNamespace

`func (o *WorkspaceProvisionRequestStatus) HasWorkspaceNamespace() bool`

HasWorkspaceNamespace returns a boolean if a field has been set.

### SetWorkspaceNamespaceNil

`func (o *WorkspaceProvisionRequestStatus) SetWorkspaceNamespaceNil(b bool)`

 SetWorkspaceNamespaceNil sets the value for WorkspaceNamespace to be an explicit nil

### UnsetWorkspaceNamespace
`func (o *WorkspaceProvisionRequestStatus) UnsetWorkspaceNamespace()`

UnsetWorkspaceNamespace ensures that no value is present for WorkspaceNamespace, not even an explicit nil
### GetWorkspaceServiceAccountname

`func (o *WorkspaceProvisionRequestStatus) GetWorkspaceServiceAccountname() string`

GetWorkspaceServiceAccountname returns the WorkspaceServiceAccountname field if non-nil, zero value otherwise.

### GetWorkspaceServiceAccountnameOk

`func (o *WorkspaceProvisionRequestStatus) GetWorkspaceServiceAccountnameOk() (*string, bool)`

GetWorkspaceServiceAccountnameOk returns a tuple with the WorkspaceServiceAccountname field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceServiceAccountname

`func (o *WorkspaceProvisionRequestStatus) SetWorkspaceServiceAccountname(v string)`

SetWorkspaceServiceAccountname sets WorkspaceServiceAccountname field to given value.

### HasWorkspaceServiceAccountname

`func (o *WorkspaceProvisionRequestStatus) HasWorkspaceServiceAccountname() bool`

HasWorkspaceServiceAccountname returns a boolean if a field has been set.

### SetWorkspaceServiceAccountnameNil

`func (o *WorkspaceProvisionRequestStatus) SetWorkspaceServiceAccountnameNil(b bool)`

 SetWorkspaceServiceAccountnameNil sets the value for WorkspaceServiceAccountname to be an explicit nil

### UnsetWorkspaceServiceAccountname
`func (o *WorkspaceProvisionRequestStatus) UnsetWorkspaceServiceAccountname()`

UnsetWorkspaceServiceAccountname ensures that no value is present for WorkspaceServiceAccountname, not even an explicit nil
### GetWorkspaceServiceaccountToken

`func (o *WorkspaceProvisionRequestStatus) GetWorkspaceServiceaccountToken() string`

GetWorkspaceServiceaccountToken returns the WorkspaceServiceaccountToken field if non-nil, zero value otherwise.

### GetWorkspaceServiceaccountTokenOk

`func (o *WorkspaceProvisionRequestStatus) GetWorkspaceServiceaccountTokenOk() (*string, bool)`

GetWorkspaceServiceaccountTokenOk returns a tuple with the WorkspaceServiceaccountToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspaceServiceaccountToken

`func (o *WorkspaceProvisionRequestStatus) SetWorkspaceServiceaccountToken(v string)`

SetWorkspaceServiceaccountToken sets WorkspaceServiceaccountToken field to given value.

### HasWorkspaceServiceaccountToken

`func (o *WorkspaceProvisionRequestStatus) HasWorkspaceServiceaccountToken() bool`

HasWorkspaceServiceaccountToken returns a boolean if a field has been set.

### SetWorkspaceServiceaccountTokenNil

`func (o *WorkspaceProvisionRequestStatus) SetWorkspaceServiceaccountTokenNil(b bool)`

 SetWorkspaceServiceaccountTokenNil sets the value for WorkspaceServiceaccountToken to be an explicit nil

### UnsetWorkspaceServiceaccountToken
`func (o *WorkspaceProvisionRequestStatus) UnsetWorkspaceServiceaccountToken()`

UnsetWorkspaceServiceaccountToken ensures that no value is present for WorkspaceServiceaccountToken, not even an explicit nil
### GetClusterCaCert

`func (o *WorkspaceProvisionRequestStatus) GetClusterCaCert() string`

GetClusterCaCert returns the ClusterCaCert field if non-nil, zero value otherwise.

### GetClusterCaCertOk

`func (o *WorkspaceProvisionRequestStatus) GetClusterCaCertOk() (*string, bool)`

GetClusterCaCertOk returns a tuple with the ClusterCaCert field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterCaCert

`func (o *WorkspaceProvisionRequestStatus) SetClusterCaCert(v string)`

SetClusterCaCert sets ClusterCaCert field to given value.

### HasClusterCaCert

`func (o *WorkspaceProvisionRequestStatus) HasClusterCaCert() bool`

HasClusterCaCert returns a boolean if a field has been set.

### SetClusterCaCertNil

`func (o *WorkspaceProvisionRequestStatus) SetClusterCaCertNil(b bool)`

 SetClusterCaCertNil sets the value for ClusterCaCert to be an explicit nil

### UnsetClusterCaCert
`func (o *WorkspaceProvisionRequestStatus) UnsetClusterCaCert()`

UnsetClusterCaCert ensures that no value is present for ClusterCaCert, not even an explicit nil
### GetClusterUrl

`func (o *WorkspaceProvisionRequestStatus) GetClusterUrl() string`

GetClusterUrl returns the ClusterUrl field if non-nil, zero value otherwise.

### GetClusterUrlOk

`func (o *WorkspaceProvisionRequestStatus) GetClusterUrlOk() (*string, bool)`

GetClusterUrlOk returns a tuple with the ClusterUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUrl

`func (o *WorkspaceProvisionRequestStatus) SetClusterUrl(v string)`

SetClusterUrl sets ClusterUrl field to given value.

### HasClusterUrl

`func (o *WorkspaceProvisionRequestStatus) HasClusterUrl() bool`

HasClusterUrl returns a boolean if a field has been set.

### SetClusterUrlNil

`func (o *WorkspaceProvisionRequestStatus) SetClusterUrlNil(b bool)`

 SetClusterUrlNil sets the value for ClusterUrl to be an explicit nil

### UnsetClusterUrl
`func (o *WorkspaceProvisionRequestStatus) UnsetClusterUrl()`

UnsetClusterUrl ensures that no value is present for ClusterUrl, not even an explicit nil
### GetDomain

`func (o *WorkspaceProvisionRequestStatus) GetDomain() string`

GetDomain returns the Domain field if non-nil, zero value otherwise.

### GetDomainOk

`func (o *WorkspaceProvisionRequestStatus) GetDomainOk() (*string, bool)`

GetDomainOk returns a tuple with the Domain field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDomain

`func (o *WorkspaceProvisionRequestStatus) SetDomain(v string)`

SetDomain sets Domain field to given value.

### HasDomain

`func (o *WorkspaceProvisionRequestStatus) HasDomain() bool`

HasDomain returns a boolean if a field has been set.

### SetDomainNil

`func (o *WorkspaceProvisionRequestStatus) SetDomainNil(b bool)`

 SetDomainNil sets the value for Domain to be an explicit nil

### UnsetDomain
`func (o *WorkspaceProvisionRequestStatus) UnsetDomain()`

UnsetDomain ensures that no value is present for Domain, not even an explicit nil
### GetStatusCondition

`func (o *WorkspaceProvisionRequestStatus) GetStatusCondition() ProvisionRequestStatusCondition`

GetStatusCondition returns the StatusCondition field if non-nil, zero value otherwise.

### GetStatusConditionOk

`func (o *WorkspaceProvisionRequestStatus) GetStatusConditionOk() (*ProvisionRequestStatusCondition, bool)`

GetStatusConditionOk returns a tuple with the StatusCondition field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatusCondition

`func (o *WorkspaceProvisionRequestStatus) SetStatusCondition(v ProvisionRequestStatusCondition)`

SetStatusCondition sets StatusCondition field to given value.

### HasStatusCondition

`func (o *WorkspaceProvisionRequestStatus) HasStatusCondition() bool`

HasStatusCondition returns a boolean if a field has been set.

### GetMessage

`func (o *WorkspaceProvisionRequestStatus) GetMessage() string`

GetMessage returns the Message field if non-nil, zero value otherwise.

### GetMessageOk

`func (o *WorkspaceProvisionRequestStatus) GetMessageOk() (*string, bool)`

GetMessageOk returns a tuple with the Message field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMessage

`func (o *WorkspaceProvisionRequestStatus) SetMessage(v string)`

SetMessage sets Message field to given value.

### HasMessage

`func (o *WorkspaceProvisionRequestStatus) HasMessage() bool`

HasMessage returns a boolean if a field has been set.

### SetMessageNil

`func (o *WorkspaceProvisionRequestStatus) SetMessageNil(b bool)`

 SetMessageNil sets the value for Message to be an explicit nil

### UnsetMessage
`func (o *WorkspaceProvisionRequestStatus) UnsetMessage()`

UnsetMessage ensures that no value is present for Message, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


