# StackResource

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | Pointer to **string** |  | [optional] [readonly] 
**StackId** | Pointer to **string** |  | [optional] [readonly] 
**Name** | **string** |  | 
**Labels** | Pointer to [**[]Label**](Label.md) |  | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) |  | [optional] 
**Revision** | Pointer to **string** |  | [optional] [readonly] 
**BuildSpec** | Pointer to [**StackResourceBuildSpec**](StackResourceBuildSpec.md) |  | [optional] 
**ImageSpec** | Pointer to [**ImageSpec**](ImageSpec.md) |  | [optional] 
**InitSpec** | Pointer to [**InitSpec**](InitSpec.md) |  | [optional] 
**ExecutionConfig** | Pointer to [**ExecutionConfig**](ExecutionConfig.md) |  | [optional] 
**VolumeMounts** | Pointer to [**[]VolumeMount**](VolumeMount.md) |  | [optional] 
**DependsOn** | Pointer to **[]string** |  | [optional] 
**LifecycleConfig** | Pointer to [**LifecycleConfig**](LifecycleConfig.md) |  | [optional] 
**Ports** | Pointer to [**[]Port**](Port.md) |  | [optional] 
**Stateful** | Pointer to **bool** |  | [optional] 
**Status** | Pointer to [**StackResourceStatus**](StackResourceStatus.md) |  | [optional] 

## Methods

### NewStackResource

`func NewStackResource(name string, ) *StackResource`

NewStackResource instantiates a new StackResource object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStackResourceWithDefaults

`func NewStackResourceWithDefaults() *StackResource`

NewStackResourceWithDefaults instantiates a new StackResource object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetId

`func (o *StackResource) GetId() string`

GetId returns the Id field if non-nil, zero value otherwise.

### GetIdOk

`func (o *StackResource) GetIdOk() (*string, bool)`

GetIdOk returns a tuple with the Id field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetId

`func (o *StackResource) SetId(v string)`

SetId sets Id field to given value.

### HasId

`func (o *StackResource) HasId() bool`

HasId returns a boolean if a field has been set.

### GetStackId

`func (o *StackResource) GetStackId() string`

GetStackId returns the StackId field if non-nil, zero value otherwise.

### GetStackIdOk

`func (o *StackResource) GetStackIdOk() (*string, bool)`

GetStackIdOk returns a tuple with the StackId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStackId

`func (o *StackResource) SetStackId(v string)`

SetStackId sets StackId field to given value.

### HasStackId

`func (o *StackResource) HasStackId() bool`

HasStackId returns a boolean if a field has been set.

### GetName

`func (o *StackResource) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *StackResource) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *StackResource) SetName(v string)`

SetName sets Name field to given value.


### GetLabels

`func (o *StackResource) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *StackResource) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *StackResource) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *StackResource) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### GetAnnotations

`func (o *StackResource) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *StackResource) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *StackResource) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *StackResource) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### GetRevision

`func (o *StackResource) GetRevision() string`

GetRevision returns the Revision field if non-nil, zero value otherwise.

### GetRevisionOk

`func (o *StackResource) GetRevisionOk() (*string, bool)`

GetRevisionOk returns a tuple with the Revision field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRevision

`func (o *StackResource) SetRevision(v string)`

SetRevision sets Revision field to given value.

### HasRevision

`func (o *StackResource) HasRevision() bool`

HasRevision returns a boolean if a field has been set.

### GetBuildSpec

`func (o *StackResource) GetBuildSpec() StackResourceBuildSpec`

GetBuildSpec returns the BuildSpec field if non-nil, zero value otherwise.

### GetBuildSpecOk

`func (o *StackResource) GetBuildSpecOk() (*StackResourceBuildSpec, bool)`

GetBuildSpecOk returns a tuple with the BuildSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBuildSpec

`func (o *StackResource) SetBuildSpec(v StackResourceBuildSpec)`

SetBuildSpec sets BuildSpec field to given value.

### HasBuildSpec

`func (o *StackResource) HasBuildSpec() bool`

HasBuildSpec returns a boolean if a field has been set.

### GetImageSpec

`func (o *StackResource) GetImageSpec() ImageSpec`

GetImageSpec returns the ImageSpec field if non-nil, zero value otherwise.

### GetImageSpecOk

`func (o *StackResource) GetImageSpecOk() (*ImageSpec, bool)`

GetImageSpecOk returns a tuple with the ImageSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImageSpec

`func (o *StackResource) SetImageSpec(v ImageSpec)`

SetImageSpec sets ImageSpec field to given value.

### HasImageSpec

`func (o *StackResource) HasImageSpec() bool`

HasImageSpec returns a boolean if a field has been set.

### GetInitSpec

`func (o *StackResource) GetInitSpec() InitSpec`

GetInitSpec returns the InitSpec field if non-nil, zero value otherwise.

### GetInitSpecOk

`func (o *StackResource) GetInitSpecOk() (*InitSpec, bool)`

GetInitSpecOk returns a tuple with the InitSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInitSpec

`func (o *StackResource) SetInitSpec(v InitSpec)`

SetInitSpec sets InitSpec field to given value.

### HasInitSpec

`func (o *StackResource) HasInitSpec() bool`

HasInitSpec returns a boolean if a field has been set.

### GetExecutionConfig

`func (o *StackResource) GetExecutionConfig() ExecutionConfig`

GetExecutionConfig returns the ExecutionConfig field if non-nil, zero value otherwise.

### GetExecutionConfigOk

`func (o *StackResource) GetExecutionConfigOk() (*ExecutionConfig, bool)`

GetExecutionConfigOk returns a tuple with the ExecutionConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExecutionConfig

`func (o *StackResource) SetExecutionConfig(v ExecutionConfig)`

SetExecutionConfig sets ExecutionConfig field to given value.

### HasExecutionConfig

`func (o *StackResource) HasExecutionConfig() bool`

HasExecutionConfig returns a boolean if a field has been set.

### GetVolumeMounts

`func (o *StackResource) GetVolumeMounts() []VolumeMount`

GetVolumeMounts returns the VolumeMounts field if non-nil, zero value otherwise.

### GetVolumeMountsOk

`func (o *StackResource) GetVolumeMountsOk() (*[]VolumeMount, bool)`

GetVolumeMountsOk returns a tuple with the VolumeMounts field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeMounts

`func (o *StackResource) SetVolumeMounts(v []VolumeMount)`

SetVolumeMounts sets VolumeMounts field to given value.

### HasVolumeMounts

`func (o *StackResource) HasVolumeMounts() bool`

HasVolumeMounts returns a boolean if a field has been set.

### GetDependsOn

`func (o *StackResource) GetDependsOn() []string`

GetDependsOn returns the DependsOn field if non-nil, zero value otherwise.

### GetDependsOnOk

`func (o *StackResource) GetDependsOnOk() (*[]string, bool)`

GetDependsOnOk returns a tuple with the DependsOn field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDependsOn

`func (o *StackResource) SetDependsOn(v []string)`

SetDependsOn sets DependsOn field to given value.

### HasDependsOn

`func (o *StackResource) HasDependsOn() bool`

HasDependsOn returns a boolean if a field has been set.

### GetLifecycleConfig

`func (o *StackResource) GetLifecycleConfig() LifecycleConfig`

GetLifecycleConfig returns the LifecycleConfig field if non-nil, zero value otherwise.

### GetLifecycleConfigOk

`func (o *StackResource) GetLifecycleConfigOk() (*LifecycleConfig, bool)`

GetLifecycleConfigOk returns a tuple with the LifecycleConfig field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLifecycleConfig

`func (o *StackResource) SetLifecycleConfig(v LifecycleConfig)`

SetLifecycleConfig sets LifecycleConfig field to given value.

### HasLifecycleConfig

`func (o *StackResource) HasLifecycleConfig() bool`

HasLifecycleConfig returns a boolean if a field has been set.

### GetPorts

`func (o *StackResource) GetPorts() []Port`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *StackResource) GetPortsOk() (*[]Port, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *StackResource) SetPorts(v []Port)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *StackResource) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### GetStateful

`func (o *StackResource) GetStateful() bool`

GetStateful returns the Stateful field if non-nil, zero value otherwise.

### GetStatefulOk

`func (o *StackResource) GetStatefulOk() (*bool, bool)`

GetStatefulOk returns a tuple with the Stateful field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStateful

`func (o *StackResource) SetStateful(v bool)`

SetStateful sets Stateful field to given value.

### HasStateful

`func (o *StackResource) HasStateful() bool`

HasStateful returns a boolean if a field has been set.

### GetStatus

`func (o *StackResource) GetStatus() StackResourceStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *StackResource) GetStatusOk() (*StackResourceStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *StackResource) SetStatus(v StackResourceStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *StackResource) HasStatus() bool`

HasStatus returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


