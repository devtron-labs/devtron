# \DefaultApi

All URIs are relative to *http://localhost*

Method | HTTP request | Description
------------- | ------------- | -------------
[**OrchestratorApplicationRollbackPut**](DefaultApi.md#OrchestratorApplicationRollbackPut) | **Put** /orchestrator/application/rollback | 
[**OrchestratorApplicationTemplateChartPost**](DefaultApi.md#OrchestratorApplicationTemplateChartPost) | **Post** /orchestrator/application/template-chart | 



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


## OrchestratorApplicationTemplateChartPost

> TemplateChartResponse OrchestratorApplicationTemplateChartPost(ctx).TemplateChartRequest(templateChartRequest).Execute()





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
    templateChartRequest := *openapiclient.NewTemplateChartRequest() // TemplateChartRequest | 

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.OrchestratorApplicationTemplateChartPost(context.Background()).TemplateChartRequest(templateChartRequest).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.OrchestratorApplicationTemplateChartPost``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `OrchestratorApplicationTemplateChartPost`: TemplateChartResponse
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.OrchestratorApplicationTemplateChartPost`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiOrchestratorApplicationTemplateChartPostRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **templateChartRequest** | [**TemplateChartRequest**](TemplateChartRequest.md) |  | 

### Return type

[**TemplateChartResponse**](TemplateChartResponse.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: application/json
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

