# StackResourceFailure

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to **string** |  | [optional] 
**Container** | Pointer to [**ContainerFailureDetail**](ContainerFailureDetail.md) |  | [optional] 
**InitContainer** | Pointer to [**ContainerFailureDetail**](ContainerFailureDetail.md) |  | [optional] 
**Build** | Pointer to [**BuildFailureDetail**](BuildFailureDetail.md) |  | [optional] 

## Methods

### NewStackResourceFailure

`func NewStackResourceFailure() *StackResourceFailure`

NewStackResourceFailure instantiates a new StackResourceFailure object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackResourceFailureWithDefaults

`func NewStackResourceFailureWithDefaults() *StackResourceFailure`

NewStackResourceFailureWithDefaults instantiates a new StackResourceFailure object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *StackResourceFailure) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *StackResourceFailure) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *StackResourceFailure) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *StackResourceFailure) HasType() bool`

HasType returns a boolean if a field has been set.

### GetContainer

`func (o *StackResourceFailure) GetContainer() ContainerFailureDetail`

GetContainer returns the Container field if non-nil, zero value otherwise.

### GetContainerOk

`func (o *StackResourceFailure) GetContainerOk() (*ContainerFailureDetail, bool)`

GetContainerOk returns a tuple with the Container field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetContainer

`func (o *StackResourceFailure) SetContainer(v ContainerFailureDetail)`

SetContainer sets Container field to given value.

### HasContainer

`func (o *StackResourceFailure) HasContainer() bool`

HasContainer returns a boolean if a field has been set.

### GetInitContainer

`func (o *StackResourceFailure) GetInitContainer() ContainerFailureDetail`

GetInitContainer returns the InitContainer field if non-nil, zero value otherwise.

### GetInitContainerOk

`func (o *StackResourceFailure) GetInitContainerOk() (*ContainerFailureDetail, bool)`

GetInitContainerOk returns a tuple with the InitContainer field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitContainer

`func (o *StackResourceFailure) SetInitContainer(v ContainerFailureDetail)`

SetInitContainer sets InitContainer field to given value.

### HasInitContainer

`func (o *StackResourceFailure) HasInitContainer() bool`

HasInitContainer returns a boolean if a field has been set.

### GetBuild

`func (o *StackResourceFailure) GetBuild() BuildFailureDetail`

GetBuild returns the Build field if non-nil, zero value otherwise.

### GetBuildOk

`func (o *StackResourceFailure) GetBuildOk() (*BuildFailureDetail, bool)`

GetBuildOk returns a tuple with the Build field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuild

`func (o *StackResourceFailure) SetBuild(v BuildFailureDetail)`

SetBuild sets Build field to given value.

### HasBuild

`func (o *StackResourceFailure) HasBuild() bool`

HasBuild returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


