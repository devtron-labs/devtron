# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**OrchestratorAppStoreDeploymentApplicationRollbackPut**](DefaultApi.md#OrchestratorAppStoreDeploymentApplicationRollbackPut) | **Put** /orchestrator/app-store/deployment/application/rollback | 
[**OrchestratorApplicationRollbackPut**](DefaultApi.md#OrchestratorApplicationRollbackPut) | **Put** /orchestrator/application/rollback | 



## OrchestratorAppStoreDeploymentApplicationRollbackPut

> RollbackReleaseResponse OrchestratorAppStoreDeploymentApplicationRollbackPut(ctx).RollbackReleaseRequest(rollbackReleaseRequest).Execute()





### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    rollbackReleaseRequest := *openapiclient.NewRollbackReleaseRequest() // RollbackReleaseRequest | 

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.OrchestratorAppStoreDeploymentApplicationRollbackPut(context.Background()).RollbackReleaseRequest(rollbackReleaseRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorAppStoreDeploymentApplicationRollbackPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorAppStoreDeploymentApplicationRollbackPut`: RollbackReleaseResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorAppStoreDeploymentApplicationRollbackPut`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorAppStoreDeploymentApplicationRollbackPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **rollbackReleaseRequest** | [**RollbackReleaseRequest**](RollbackReleaseRequest.md) |  | 

### Return type

[**RollbackReleaseResponse**](RollbackReleaseResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## OrchestratorApplicationRollbackPut

> RollbackReleaseResponse OrchestratorApplicationRollbackPut(ctx).RollbackReleaseRequest(rollbackReleaseRequest).Execute()





### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    rollbackReleaseRequest := *openapiclient.NewRollbackReleaseRequest() // RollbackReleaseRequest | 

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.OrchestratorApplicationRollbackPut(context.Background()).RollbackReleaseRequest(rollbackReleaseRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApplicationRollbackPut``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApplicationRollbackPut`: RollbackReleaseResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorApplicationRollbackPut`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorApplicationRollbackPutRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **rollbackReleaseRequest** | [**RollbackReleaseRequest**](RollbackReleaseRequest.md) |  | 

### Return type

[**RollbackReleaseResponse**](RollbackReleaseResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

