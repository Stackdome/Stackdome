# WorkspaceUserStatus

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ObservedVersion** | Pointer to **int32** |  | [optional] 
**ProvisionedNamespaces** | Pointer to [**[]WorkspaceUserStatusProvisionedNamespacesInner**](WorkspaceUserStatusProvisionedNamespacesInner.md) |  | [optional] 
**ServiceAccountName** | Pointer to **NullableString** |  | [optional] 
**ServiceaccountToken** | Pointer to **NullableString** |  | [optional] 
**ClusterCaCert** | Pointer to **NullableString** |  | [optional] 
**ClusterUrl** | Pointer to **NullableString** |  | [optional] 
**Conditions** | Pointer to [**[]Condition**](Condition.md) |  | [optional] 

## Methods

### NewWorkspaceUserStatus

`func NewWorkspaceUserStatus() *WorkspaceUserStatus`

NewWorkspaceUserStatus instantiates a new WorkspaceUserStatus object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceUserStatusWithDefaults

`func NewWorkspaceUserStatusWithDefaults() *WorkspaceUserStatus`

NewWorkspaceUserStatusWithDefaults instantiates a new WorkspaceUserStatus object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetObservedVersion

`func (o *WorkspaceUserStatus) GetObservedVersion() int32`

GetObservedVersion returns the ObservedVersion field if non-nil, zero value otherwise.

### GetObservedVersionOk

`func (o *WorkspaceUserStatus) GetObservedVersionOk() (*int32, bool)`

GetObservedVersionOk returns a tuple with the ObservedVersion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetObservedVersion

`func (o *WorkspaceUserStatus) SetObservedVersion(v int32)`

SetObservedVersion sets ObservedVersion field to given value.

### HasObservedVersion

`func (o *WorkspaceUserStatus) HasObservedVersion() bool`

HasObservedVersion returns a boolean if a field has been set.

### GetProvisionedNamespaces

`func (o *WorkspaceUserStatus) GetProvisionedNamespaces() []WorkspaceUserStatusProvisionedNamespacesInner`

GetProvisionedNamespaces returns the ProvisionedNamespaces field if non-nil, zero value otherwise.

### GetProvisionedNamespacesOk

`func (o *WorkspaceUserStatus) GetProvisionedNamespacesOk() (*[]WorkspaceUserStatusProvisionedNamespacesInner, bool)`

GetProvisionedNamespacesOk returns a tuple with the ProvisionedNamespaces field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProvisionedNamespaces

`func (o *WorkspaceUserStatus) SetProvisionedNamespaces(v []WorkspaceUserStatusProvisionedNamespacesInner)`

SetProvisionedNamespaces sets ProvisionedNamespaces field to given value.

### HasProvisionedNamespaces

`func (o *WorkspaceUserStatus) HasProvisionedNamespaces() bool`

HasProvisionedNamespaces returns a boolean if a field has been set.

### GetServiceAccountName

`func (o *WorkspaceUserStatus) GetServiceAccountName() string`

GetServiceAccountName returns the ServiceAccountName field if non-nil, zero value otherwise.

### GetServiceAccountNameOk

`func (o *WorkspaceUserStatus) GetServiceAccountNameOk() (*string, bool)`

GetServiceAccountNameOk returns a tuple with the ServiceAccountName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceAccountName

`func (o *WorkspaceUserStatus) SetServiceAccountName(v string)`

SetServiceAccountName sets ServiceAccountName field to given value.

### HasServiceAccountName

`func (o *WorkspaceUserStatus) HasServiceAccountName() bool`

HasServiceAccountName returns a boolean if a field has been set.

### SetServiceAccountNameNil

`func (o *WorkspaceUserStatus) SetServiceAccountNameNil(b bool)`

 SetServiceAccountNameNil sets the value for ServiceAccountName to be an explicit nil

### UnsetServiceAccountName
`func (o *WorkspaceUserStatus) UnsetServiceAccountName()`

UnsetServiceAccountName ensures that no value is present for ServiceAccountName, not even an explicit nil
### GetServiceaccountToken

`func (o *WorkspaceUserStatus) GetServiceaccountToken() string`

GetServiceaccountToken returns the ServiceaccountToken field if non-nil, zero value otherwise.

### GetServiceaccountTokenOk

`func (o *WorkspaceUserStatus) GetServiceaccountTokenOk() (*string, bool)`

GetServiceaccountTokenOk returns a tuple with the ServiceaccountToken field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServiceaccountToken

`func (o *WorkspaceUserStatus) SetServiceaccountToken(v string)`

SetServiceaccountToken sets ServiceaccountToken field to given value.

### HasServiceaccountToken

`func (o *WorkspaceUserStatus) HasServiceaccountToken() bool`

HasServiceaccountToken returns a boolean if a field has been set.

### SetServiceaccountTokenNil

`func (o *WorkspaceUserStatus) SetServiceaccountTokenNil(b bool)`

 SetServiceaccountTokenNil sets the value for ServiceaccountToken to be an explicit nil

### UnsetServiceaccountToken
`func (o *WorkspaceUserStatus) UnsetServiceaccountToken()`

UnsetServiceaccountToken ensures that no value is present for ServiceaccountToken, not even an explicit nil
### GetClusterCaCert

`func (o *WorkspaceUserStatus) GetClusterCaCert() string`

GetClusterCaCert returns the ClusterCaCert field if non-nil, zero value otherwise.

### GetClusterCaCertOk

`func (o *WorkspaceUserStatus) GetClusterCaCertOk() (*string, bool)`

GetClusterCaCertOk returns a tuple with the ClusterCaCert field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterCaCert

`func (o *WorkspaceUserStatus) SetClusterCaCert(v string)`

SetClusterCaCert sets ClusterCaCert field to given value.

### HasClusterCaCert

`func (o *WorkspaceUserStatus) HasClusterCaCert() bool`

HasClusterCaCert returns a boolean if a field has been set.

### SetClusterCaCertNil

`func (o *WorkspaceUserStatus) SetClusterCaCertNil(b bool)`

 SetClusterCaCertNil sets the value for ClusterCaCert to be an explicit nil

### UnsetClusterCaCert
`func (o *WorkspaceUserStatus) UnsetClusterCaCert()`

UnsetClusterCaCert ensures that no value is present for ClusterCaCert, not even an explicit nil
### GetClusterUrl

`func (o *WorkspaceUserStatus) GetClusterUrl() string`

GetClusterUrl returns the ClusterUrl field if non-nil, zero value otherwise.

### GetClusterUrlOk

`func (o *WorkspaceUserStatus) GetClusterUrlOk() (*string, bool)`

GetClusterUrlOk returns a tuple with the ClusterUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterUrl

`func (o *WorkspaceUserStatus) SetClusterUrl(v string)`

SetClusterUrl sets ClusterUrl field to given value.

### HasClusterUrl

`func (o *WorkspaceUserStatus) HasClusterUrl() bool`

HasClusterUrl returns a boolean if a field has been set.

### SetClusterUrlNil

`func (o *WorkspaceUserStatus) SetClusterUrlNil(b bool)`

 SetClusterUrlNil sets the value for ClusterUrl to be an explicit nil

### UnsetClusterUrl
`func (o *WorkspaceUserStatus) UnsetClusterUrl()`

UnsetClusterUrl ensures that no value is present for ClusterUrl, not even an explicit nil
### GetConditions

`func (o *WorkspaceUserStatus) GetConditions() []Condition`

GetConditions returns the Conditions field if non-nil, zero value otherwise.

### GetConditionsOk

`func (o *WorkspaceUserStatus) GetConditionsOk() (*[]Condition, bool)`

GetConditionsOk returns a tuple with the Conditions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConditions

`func (o *WorkspaceUserStatus) SetConditions(v []Condition)`

SetConditions sets Conditions field to given value.

### HasConditions

`func (o *WorkspaceUserStatus) HasConditions() bool`

HasConditions returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


