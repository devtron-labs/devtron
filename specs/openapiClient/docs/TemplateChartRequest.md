# TemplateChartRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**EnvironmentId** | Pointer to **int32** | environment Id on which helm template would be performed | [optional] 
**ClusterId** | Pointer to **int32** | If environmentId is not provided, clusterId should be passed | [optional] 
**Namespace** | Pointer to **string** | If environmentId is not provided, namespace should be passed | [optional] 
**ReleaseName** | Pointer to **string** | release name of helm app (if not provided, some random name is picked) | [optional] 
**AppStoreApplicationVersionId** | Pointer to **int32** | App store application version Id | [optional] 
**ValuesYaml** | Pointer to **string** | Values yaml | [optional] 

## Methods

### NewTemplateChartRequest

`func NewTemplateChartRequest() *TemplateChartRequest`

NewTemplateChartRequest instantiates a new TemplateChartRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTemplateChartRequestWithDefaults

`func NewTemplateChartRequestWithDefaults() *TemplateChartRequest`

NewTemplateChartRequestWithDefaults instantiates a new TemplateChartRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetEnvironmentId

`func (o *TemplateChartRequest) GetEnvironmentId() int32`

GetEnvironmentId returns the EnvironmentId field if non-nil, zero value otherwise.

### GetEnvironmentIdOk

`func (o *TemplateChartRequest) GetEnvironmentIdOk() (*int32, bool)`

GetEnvironmentIdOk returns a tuple with the EnvironmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentId

`func (o *TemplateChartRequest) SetEnvironmentId(v int32)`

SetEnvironmentId sets EnvironmentId field to given value.

### HasEnvironmentId

`func (o *TemplateChartRequest) HasEnvironmentId() bool`

HasEnvironmentId returns a boolean if a field has been set.

### GetClusterId

`func (o *TemplateChartRequest) GetClusterId() int32`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *TemplateChartRequest) GetClusterIdOk() (*int32, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *TemplateChartRequest) SetClusterId(v int32)`

SetClusterId sets ClusterId field to given value.

### HasClusterId

`func (o *TemplateChartRequest) HasClusterId() bool`

HasClusterId returns a boolean if a field has been set.

### GetNamespace

`func (o *TemplateChartRequest) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *TemplateChartRequest) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *TemplateChartRequest) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *TemplateChartRequest) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.

### GetReleaseName

`func (o *TemplateChartRequest) GetReleaseName() string`

GetReleaseName returns the ReleaseName field if non-nil, zero value otherwise.

### GetReleaseNameOk

`func (o *TemplateChartRequest) GetReleaseNameOk() (*string, bool)`

GetReleaseNameOk returns a tuple with the ReleaseName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReleaseName

`func (o *TemplateChartRequest) SetReleaseName(v string)`

SetReleaseName sets ReleaseName field to given value.

### HasReleaseName

`func (o *TemplateChartRequest) HasReleaseName() bool`

HasReleaseName returns a boolean if a field has been set.

### GetAppStoreApplicationVersionId

`func (o *TemplateChartRequest) GetAppStoreApplicationVersionId() int32`

GetAppStoreApplicationVersionId returns the AppStoreApplicationVersionId field if non-nil, zero value otherwise.

### GetAppStoreApplicationVersionIdOk

`func (o *TemplateChartRequest) GetAppStoreApplicationVersionIdOk() (*int32, bool)`

GetAppStoreApplicationVersionIdOk returns a tuple with the AppStoreApplicationVersionId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAppStoreApplicationVersionId

`func (o *TemplateChartRequest) SetAppStoreApplicationVersionId(v int32)`

SetAppStoreApplicationVersionId sets AppStoreApplicationVersionId field to given value.

### HasAppStoreApplicationVersionId

`func (o *TemplateChartRequest) HasAppStoreApplicationVersionId() bool`

HasAppStoreApplicationVersionId returns a boolean if a field has been set.

### GetValuesYaml

`func (o *TemplateChartRequest) GetValuesYaml() string`

GetValuesYaml returns the ValuesYaml field if non-nil, zero value otherwise.

### GetValuesYamlOk

`func (o *TemplateChartRequest) GetValuesYamlOk() (*string, bool)`

GetValuesYamlOk returns a tuple with the ValuesYaml field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValuesYaml

`func (o *TemplateChartRequest) SetValuesYaml(v string)`

SetValuesYaml sets ValuesYaml field to given value.

### HasValuesYaml

`func (o *TemplateChartRequest) HasValuesYaml() bool`

HasValuesYaml returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


