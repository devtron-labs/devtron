# DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**orchestratorAppListGet**](DefaultApi.md#orchestratorAppListGet) | **GET** /orchestrator/app/list/ | 
[**orchestratorApplicationClusterEnvDetailsGet**](DefaultApi.md#orchestratorApplicationClusterEnvDetailsGet) | **GET** /orchestrator/application/cluster-env-details | 
[**orchestratorApplicationGet**](DefaultApi.md#orchestratorApplicationGet) | **GET** /orchestrator/application/ | 


<a name="orchestratorAppListGet"></a>
# **orchestratorAppListGet**
> AppList orchestratorAppListGet(projectIds, clusterIds, environmentIds, offset, size, sortOrder, sortBy)



    this api gives all devtron applications.

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **projectIds** | [**List**](../Models/Integer.md)| project ids | [default to null]
 **clusterIds** | [**List**](../Models/Integer.md)| cluster ids | [default to null]
 **environmentIds** | [**List**](../Models/Integer.md)| environment ids | [default to null]
 **offset** | **Integer**| offser | [default to null]
 **size** | **Integer**| size | [default to null]
 **sortOrder** | **String**| sortOrder | [default to null] [enum: ASC, DESC]
 **sortBy** | **String**| sortBy | [default to null] [enum: appNameSort]

### Return type

[**AppList**](../Models/AppList.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="orchestratorApplicationClusterEnvDetailsGet"></a>
# **orchestratorApplicationClusterEnvDetailsGet**
> ClusterEnvironmentDetail orchestratorApplicationClusterEnvDetailsGet()



    returns cluster environment namespace mappings

### Parameters
This endpoint does not need any parameter.

### Return type

[**ClusterEnvironmentDetail**](../Models/ClusterEnvironmentDetail.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

<a name="orchestratorApplicationGet"></a>
# **orchestratorApplicationGet**
> AppList orchestratorApplicationGet(clusterIds)



    this api gives all external application+ devtron helm chart applications.

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **clusterIds** | [**List**](../Models/Integer.md)| cluster ids | [default to null]

### Return type

[**AppList**](../Models/AppList.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: text/event-stream

